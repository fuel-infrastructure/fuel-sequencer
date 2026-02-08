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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
	blobhub "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"golang.org/x/crypto/blake2b"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/metrics"
)

type blobhubClient struct {
	logger         log.Logger
	blobpool       *Blobpool // reference to the blobpool, where blobs are stored
	blobhubAddress string    // configurable blobhub address (HTTP base URL)

	client              *blobhub.Client
	validatorID         string // validator ID derived from consensus key (empty for non-validator nodes)
	chunkMode           bool   // enable chunk-mode attestation
	chunkValidatorIndex int    // which chunk this validator attests

	privKey ed25519.PrivateKey // Ed25519 private key for signing attestations
	pubKey  ed25519.PublicKey  // Ed25519 public key registered with hub
}

// normalizeBlobhubAddress converts a blobhub address to HTTP URL format.
// If the address already contains a scheme (http:// or https://), it's returned as-is.
// Otherwise, it's assumed to be host:port and http:// is prepended.
func normalizeBlobhubAddress(address string) string {
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		return address
	}
	return "http://" + address
}

// newBlobhubClient creates a blobhub client. If client is nil, a new one is created.
// If validatorID is provided, ACK (signature submission) will be performed after blob sync.
// privKey/pubKey should be the consensus Ed25519 key pair; if nil, a random pair is generated (for testing).
func newBlobhubClient(ctx context.Context, logger log.Logger, blobpool *Blobpool, blobhubAddress string, client *blobhub.Client, validatorID string, chunkMode bool, chunkValidatorIndex int, privKey ed25519.PrivateKey, pubKey ed25519.PublicKey) (*blobhubClient, error) {
	// Normalize address to HTTP URL format
	httpURL := normalizeBlobhubAddress(blobhubAddress)

	// Create client if not provided
	if client == nil {
		config := &blobhub.ClientConfig{
			BaseURL: httpURL,
			Logger:  nil, // Use no-op logger from client library
		}
		client = blobhub.NewClient(config)
	}

	// Use provided consensus key pair, or generate random (for testing/non-validator nodes)
	pub, priv := pubKey, privKey
	if priv == nil {
		var keyErr error
		pub, priv, keyErr = ed25519.GenerateKey(nil)
		if keyErr != nil {
			return nil, fmt.Errorf("failed to generate Ed25519 keypair: %w", keyErr)
		}
	}

	blobhubClient := &blobhubClient{
		logger:              logger.With("module", "blobhub_sync"),
		blobpool:            blobpool,
		blobhubAddress:      httpURL,
		client:              client,
		validatorID:         validatorID,
		chunkMode:           chunkMode,
		chunkValidatorIndex: chunkValidatorIndex,
		privKey:             priv,
		pubKey:              pub,
	}

	// Register validator with Blobhub before starting sync (required for signing)
	if validatorID != "" {
		if err := client.RegisterWithKey(ctx, validatorID, pub); err != nil {
			blobhubClient.logger.Error("failed to register validator with Blobhub", "error", err, "validator_id", validatorID)
			return nil, fmt.Errorf("failed to register validator: %w", err)
		}
		blobhubClient.logger.Info("registered validator with Blobhub", "validator_id", validatorID, "pubkey", hex.EncodeToString(pub))
	}

	blobhubClient.logger.Info("created new blobhub client", "url", httpURL, "validator_id", validatorID, "chunk_mode", chunkMode, "chunk_index", chunkValidatorIndex)
	if chunkMode {
		go blobhubClient.syncChunks(ctx)
	} else {
		go blobhubClient.sync(ctx)
	}

	return blobhubClient, nil
}

func (c *blobhubClient) sync(ctx context.Context) {
	defer func() {
		c.logger.Info("closing blobhub connection")
		metrics.SetBlobhubConnectionStatus(false)
	}()

	c.logger.Info("starting blob sync", "url", c.blobhubAddress)

	const reconnectBackoff = 5 * time.Second

	// Retry loop for reconnection
	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping sync")
			return
		default:
			if attempt > 1 {
				c.logger.Info("blobhub reconnect attempt", "attempt", attempt, "url", c.blobhubAddress)
			}

			start := time.Now()

			// StreamBlobs establishes WebSocket connection and returns a channel
			blobChan, err := c.client.StreamBlobs(ctx)
			if err != nil {
				c.logger.Error("failed to connect to blobhub", "error", err, "url", c.blobhubAddress, "attempt", attempt)
				metrics.IncrementBlobhubErrors()
				metrics.SetBlobhubConnectionStatus(false)

				c.logger.Info("backing off before reconnect", "duration", reconnectBackoff, "attempt", attempt)
				select {
				case <-ctx.Done():
					return
				case <-time.After(reconnectBackoff):
					continue
				}
			}

			// Record successful connection metrics
			connectionTime := time.Since(start)
			metrics.SetBlobhubConnectionStatus(true)
			metrics.ObserveBlobSyncLatency(connectionTime)
			if attempt > 1 {
				c.logger.Info("connected to blobhub successfully", "url", c.blobhubAddress, "attempt", attempt, "reconnected_after_failures", attempt-1)
			} else {
				c.logger.Info("connected to blobhub successfully", "url", c.blobhubAddress, "attempt", attempt)
			}

			// Process blobs from the stream; meter ingestion rate to spot sequencer backpressure
			const ingestionRateLogInterval = 30 * time.Second
			lastRateLog := time.Now()
			blobsInWindow := 0

			for {
				select {
				case <-ctx.Done():
					c.logger.Info("context cancelled, stopping sync")
					return
				case blob, ok := <-blobChan:
					if !ok {
						// Channel closed by remote or transport (client API does not expose close reason)
						c.logger.Warn("blob stream channel closed, reconnecting", "url", c.blobhubAddress)
						metrics.SetBlobhubConnectionStatus(false)
						metrics.IncrementBlobhubReconnections()
						break // Break inner loop to retry connection
					}

					if blob == nil {
						c.logger.Debug("received nil blob, skipping")
						continue
					}

					// Skip if we already have this blob
					if c.blobpool.Has(ctx, blob.Key) {
						c.logger.Debug("skipping existing blob", "id", blob.Key.String())
						continue
					}

					blobsInWindow++
					elapsed := time.Since(lastRateLog)
					if elapsed >= ingestionRateLogInterval {
						rate := float32(blobsInWindow) / float32(elapsed.Seconds())
						metrics.SetBlobhubIngestionRate(rate)
						c.logger.Info("blobhub ingestion rate", "blobs_per_sec", rate, "blobs_in_window", blobsInWindow, "window_sec", elapsed.Seconds())
						lastRateLog = time.Now()
						blobsInWindow = 0
					}

					// Store the blob
					c.logger.Info("storing new blob", "id", blob.Key.String(), "size", len(blob.Data))
					c.blobpool.Insert(ctx, blob.Data)

					// Submit signature (ACK) if validator ID is set
					if c.validatorID != "" {
						c.logger.Info("submitting signature for blob", "id", blob.Key.String(), "validator_id", c.validatorID)
						c.signBlobAsync(ctx, blob.Key)
					} else {
						c.logger.Info("skipping signature (non-validator node)", "id", blob.Key.String())
					}
				}
			}
		}
	}
}

// signBlob submits a signature for a blob to Blobhub.
// This method is called asynchronously and includes retry logic matching
// existing patterns in the sync loop.
func (c *blobhubClient) signBlob(ctx context.Context, key store.Key) error {
	start := time.Now()

	// Retry logic matching sync reconnection pattern
	maxRetries := 3
	retryDelay := 5 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("retrying signature submission", "attempt", attempt+1, "key", key.String())
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		err := c.client.Sign(ctx, c.validatorID, key)
		if err == nil {
			// Success - record metrics
			latency := time.Since(start)
			metrics.IncrementACKSubmissions()
			metrics.ObserveACKLatency(latency)
			c.logger.Info("signature submitted to blobhub", "key", key.String(), "validator_id", c.validatorID, "latency_ms", latency.Milliseconds())
			return nil
		}

		lastErr = err
		c.logger.Warn("failed to submit signature", "error", err, "key", key.String(), "attempt", attempt+1)
		metrics.IncrementACKErrors()
	}

	// All retries failed
	c.logger.Error("failed to submit signature after retries", "error", lastErr, "key", key.String())
	return lastErr
}

// signBlobAsync submits a signature asynchronously in a goroutine.
// This ensures that signature submission doesn't block the sync flow.
func (c *blobhubClient) signBlobAsync(ctx context.Context, key store.Key) {
	go func() {
		if err := c.signBlob(ctx, key); err != nil {
			// Log error but don't block sync - errors are already logged in signBlob
			c.logger.Error("async signature submission failed", "error", err, "key", key.String())
		}
	}()
}

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

	const reconnectBackoff = 5 * time.Second

	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping chunk sync")
			return
		default:
			if attempt > 1 {
				c.logger.Info("blobhub chunk reconnect attempt", "attempt", attempt, "url", c.blobhubAddress)
			}

			start := time.Now()

			notifChan, _, err := c.client.StreamChunks(ctx, c.chunkValidatorIndex)
			if err != nil {
				c.logger.Error("failed to connect to blobhub for chunks", "error", err, "url", c.blobhubAddress, "attempt", attempt)
				metrics.IncrementBlobhubErrors()
				metrics.SetBlobhubConnectionStatus(false)

				c.logger.Info("backing off before reconnect", "duration", reconnectBackoff, "attempt", attempt)
				select {
				case <-ctx.Done():
					return
				case <-time.After(reconnectBackoff):
					continue
				}
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

			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectBackoff):
			}
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
		return blobhub.ChunkAttestation{}, false
	}
	resp, err := chunkHTTPClient.Do(req)
	if err != nil {
		return blobhub.ChunkAttestation{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return blobhub.ChunkAttestation{}, false
	}
	buf := chunkBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer chunkBufPool.Put(buf)
	if _, err = buf.ReadFrom(resp.Body); err != nil {
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
