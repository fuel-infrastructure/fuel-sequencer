package keeper

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	blobhub "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"golang.org/x/crypto/blake2b"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/metrics"
)

// --- Chunk-mode attestation ---

// chunkHTTPClient has high connection limits to avoid socket exhaustion under load.
var chunkHTTPClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
	Timeout: 30 * time.Second,
}

// chunkBufPool reuses byte buffers to avoid 1.25 MiB allocation per chunk verification.
var chunkBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 2*1024*1024)
		return bytes.NewBuffer(buf)
	},
}

// syncChunks runs the chunk-mode attestation pipeline.
// Pipeline: StreamChunks → 16 verify workers → batch collector → 8 HTTP senders
func (c *blobhubClient) syncChunks(ctx context.Context) {
	defer func() {
		c.logger.Info("closing chunk-mode blobhub connection")
		metrics.SetBlobhubConnectionStatus(false)
	}()

	c.logger.Info("starting chunk-mode sync", "url", c.blobhubAddress, "validator_index", c.chunkValidatorIndex)

	const (
		initialReconnectBackoff = 2 * time.Second
		maxReconnectBackoff     = 60 * time.Second
	)

	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping chunk sync")
			return
		default:
			if attempt > 1 {
				backoff := initialReconnectBackoff
				for i := 0; i < attempt-2; i++ {
					backoff *= 2
					if backoff > maxReconnectBackoff {
						backoff = maxReconnectBackoff
						break
					}
				}
				c.logger.Info("blobhub chunk reconnect attempt", "attempt", attempt, "url", c.blobhubAddress, "backoff", backoff)
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
				}
			}

			start := time.Now()

			notifChan, _, err := c.client.StreamChunks(ctx, c.chunkValidatorIndex)
			if err != nil {
				c.logger.Error("failed to connect to blobhub for chunks", "error", err, "url", c.blobhubAddress, "attempt", attempt)
				metrics.IncrementBlobhubErrors()
				metrics.SetBlobhubConnectionStatus(false)
				continue
			}

			connectionTime := time.Since(start)
			metrics.SetBlobhubConnectionStatus(true)
			metrics.ObserveBlobSyncLatency(connectionTime)
			if attempt > 1 {
				c.logger.Info("connected to blobhub (chunk-mode) successfully", "url", c.blobhubAddress, "attempt", attempt)
			} else {
				c.logger.Info("connected to blobhub (chunk-mode) successfully", "url", c.blobhubAddress)
			}

			c.runChunkPipeline(ctx, notifChan)

			// If we get here, the stream closed — reconnect
			c.logger.Warn("chunk stream closed, reconnecting", "url", c.blobhubAddress)
			metrics.SetBlobhubConnectionStatus(false)
			metrics.IncrementBlobhubReconnections()
		}
	}
}

// runChunkPipeline processes chunk notifications through the verify → batch → send pipeline.
func (c *blobhubClient) runChunkPipeline(ctx context.Context, notifChan <-chan *blobhub.ChunkNotification) {
	var signed atomic.Int64
	var errors atomic.Int64
	var verified atomic.Int64
	var verifyFailed atomic.Int64

	// Status logging
	statusCtx, statusCancel := context.WithCancel(ctx)
	defer statusCancel()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-statusCtx.Done():
				return
			case <-ticker.C:
				c.logger.Info("chunk attestation status",
					"signed", signed.Load(),
					"verified", verified.Load(),
					"verify_failed", verifyFailed.Load(),
					"errors", errors.Load(),
				)
			}
		}
	}()

	verifyCh := make(chan *blobhub.ChunkNotification, 1024)
	attestCh := make(chan blobhub.ChunkAttestation, 1024)
	sendCh := make(chan []blobhub.ChunkAttestation, 64)

	// 16 verify workers
	var verifyWg sync.WaitGroup
	for i := 0; i < 16; i++ {
		verifyWg.Add(1)
		go func() {
			defer verifyWg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-verifyCh:
					if !ok {
						return
					}
					att, ok := c.verifyChunk(ctx, msg)
					if !ok {
						verifyFailed.Add(1)
						continue
					}
					verified.Add(1)
					select {
					case attestCh <- att:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Batch collector
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		var batch []blobhub.ChunkAttestation
		for {
			select {
			case <-ctx.Done():
				return
			case att := <-attestCh:
				batch = append(batch, att)
			case <-ticker.C:
				if len(batch) == 0 {
					continue
				}
				atts := batch
				batch = nil
				const maxBatch = 500
				for len(atts) > maxBatch {
					part := make([]blobhub.ChunkAttestation, maxBatch)
					copy(part, atts[:maxBatch])
					select {
					case sendCh <- part:
					case <-ctx.Done():
						return
					}
					atts = atts[maxBatch:]
				}
				if len(atts) > 0 {
					select {
					case sendCh <- atts:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	// 8 concurrent HTTP senders
	for i := 0; i < 8; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case atts := <-sendCh:
					if _, err := c.client.SignChunkBatch(ctx, c.validatorID, atts); err != nil {
						errors.Add(int64(len(atts)))
						metrics.IncrementACKErrors()
					} else {
						signed.Add(int64(len(atts)))
						metrics.IncrementACKSubmissions()
					}
				}
			}
		}()
	}

	// Read from notification channel until it closes
	for {
		select {
		case <-ctx.Done():
			close(verifyCh)
			verifyWg.Wait()
			c.logger.Info("chunk pipeline shutting down",
				"signed", signed.Load(),
				"verified", verified.Load(),
				"verify_failed", verifyFailed.Load(),
				"errors", errors.Load(),
			)
			return
		case msg, ok := <-notifChan:
			if !ok {
				close(verifyCh)
				verifyWg.Wait()
				c.logger.Info("chunk notification channel closed",
					"signed", signed.Load(),
					"verified", verified.Load(),
					"verify_failed", verifyFailed.Load(),
					"errors", errors.Load(),
				)
				return
			}
			if msg == nil {
				continue
			}
			select {
			case verifyCh <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}

// verifyChunk downloads a chunk, verifies its hash and Merkle proof, returns the attestation.
func (c *blobhubClient) verifyChunk(ctx context.Context, msg *blobhub.ChunkNotification) (blobhub.ChunkAttestation, bool) {
	chunkURL := fmt.Sprintf("%s/chunks/%s/%d", c.blobhubAddress, msg.BlobKey, msg.ChunkIndex)
	req, err := http.NewRequestWithContext(ctx, "GET", chunkURL, nil)
	if err != nil {
		c.logger.Debug("verifyChunk: failed to create request", "error", err, "blob_key", msg.BlobKey, "chunk_index", msg.ChunkIndex)
		return blobhub.ChunkAttestation{}, false
	}
	resp, err := chunkHTTPClient.Do(req)
	if err != nil {
		c.logger.Debug("verifyChunk: HTTP request failed", "error", err, "blob_key", msg.BlobKey, "chunk_index", msg.ChunkIndex)
		return blobhub.ChunkAttestation{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		_, _ = io.Copy(io.Discard, resp.Body)
		c.logger.Debug("verifyChunk: non-200 status", "status", resp.StatusCode, "blob_key", msg.BlobKey, "chunk_index", msg.ChunkIndex)
		return blobhub.ChunkAttestation{}, false
	}
	buf := chunkBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer chunkBufPool.Put(buf)
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		c.logger.Debug("verifyChunk: failed to read body", "error", err, "blob_key", msg.BlobKey, "chunk_index", msg.ChunkIndex)
		return blobhub.ChunkAttestation{}, false
	}
	chunkData := buf.Bytes()

	// Hash chunk (computed locally, not from notification)
	h := blake2b.Sum256(chunkData)
	computedHash := hex.EncodeToString(h[:])

	// Verify Merkle proof
	if msg.MerkleRoot != "" && len(msg.MerkleProof) > 0 {
		var root [32]byte
		rootBytes, err := hex.DecodeString(msg.MerkleRoot)
		if err == nil && len(rootBytes) == 32 {
			copy(root[:], rootBytes)
			proof := make([][32]byte, len(msg.MerkleProof))
			validProof := true
			for i, p := range msg.MerkleProof {
				pBytes, pErr := hex.DecodeString(p)
				if pErr != nil || len(pBytes) != 32 {
					validProof = false
					break
				}
				copy(proof[i][:], pBytes)
			}
			if validProof {
				if !verifyMerkleProof(root, h, msg.ChunkIndex, msg.TotalChunks, proof) {
					c.logger.Warn("merkle proof verification failed", "chunk_index", msg.ChunkIndex, "blob_key", msg.BlobKey)
					return blobhub.ChunkAttestation{}, false
				}
			}
		}
	}

	// Ed25519 sign the attestation message: blob_key(32) || chunk_index(4 BE) || chunk_hash(32)
	blobKeyBytes, err := hex.DecodeString(msg.BlobKey)
	if err != nil || len(blobKeyBytes) != 32 {
		c.logger.Debug("verifyChunk: invalid blob key", "error", err, "blob_key", msg.BlobKey, "chunk_index", msg.ChunkIndex)
		return blobhub.ChunkAttestation{}, false
	}
	attMsg := make([]byte, 68)
	copy(attMsg[0:32], blobKeyBytes)
	binary.BigEndian.PutUint32(attMsg[32:36], uint32(msg.ChunkIndex))
	copy(attMsg[36:68], h[:])
	sig := ed25519.Sign(c.privKey, attMsg)

	return blobhub.ChunkAttestation{
		BlobKey:    msg.BlobKey,
		ChunkIndex: msg.ChunkIndex,
		ChunkHash:  computedHash,
		Signature:  hex.EncodeToString(sig),
	}, true
}

// verifyMerkleProof replays the Merkle tree to verify chunk inclusion.
func verifyMerkleProof(root, chunkHash [32]byte, idx, totalLeaves int, proof [][32]byte) bool {
	current := chunkHash
	pos := idx
	levelSize := totalLeaves
	proofIdx := 0

	for levelSize > 1 {
		if pos%2 == 0 {
			if pos+1 < levelSize {
				if proofIdx >= len(proof) {
					return false
				}
				current = hashPairBytes(current, proof[proofIdx])
				proofIdx++
			}
		} else {
			if proofIdx >= len(proof) {
				return false
			}
			current = hashPairBytes(proof[proofIdx], current)
			proofIdx++
		}
		pos = pos / 2
		levelSize = (levelSize + 1) / 2
	}
	return proofIdx == len(proof) && current == root
}

func hashPairBytes(left, right [32]byte) [32]byte {
	var combined [64]byte
	copy(combined[:32], left[:])
	copy(combined[32:], right[:])
	return blake2b.Sum256(combined[:])
}
