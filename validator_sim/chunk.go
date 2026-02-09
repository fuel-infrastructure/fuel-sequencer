// Chunk-mode attestation: mirror x/blob/keeper/blobhub.go syncChunks,
// runChunkPipeline, verifyChunk, verifyMerkleProof, hashPairBytes.
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	blobclient "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"golang.org/x/crypto/blake2b"
)

const (
	chunkVerifyWorkers = 16
	chunkBatchTicker   = 10 * time.Millisecond
	chunkMaxBatch      = 500
	chunkSendWorkers   = 8
)

// chunkHTTPClient is used as fallback when binary data is not provided in the stream.
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

var chunkBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 2*1024*1024)
		return bytes.NewBuffer(buf)
	},
}

// chunkWork pairs a notification with its binary chunk data (if streamed inline).
type chunkWork struct {
	notif *blobclient.ChunkNotification
	data  []byte // non-nil when Binary was true in the notification
}

func runChunkMode(ctx context.Context, client *blobclient.Client, blobhubURL, validatorID string, chunkValidatorIndex int, privKey ed25519.PrivateKey) {
	log.Printf("starting chunk-mode sync url=%s validator_index=%d", blobhubURL, chunkValidatorIndex)
	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			log.Print("context cancelled, stopping chunk sync")
			return
		default:
		}
		if attempt > 1 {
			log.Printf("blobhub chunk reconnect attempt attempt=%d url=%s", attempt, blobhubURL)
		}
		notifChan, dataChan, err := client.StreamChunks(ctx, chunkValidatorIndex)
		if err != nil {
			log.Printf("failed to connect to blobhub for chunks error=%v url=%s attempt=%d", err, blobhubURL, attempt)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectBackoff):
			}
			continue
		}
		log.Printf("connected to blobhub (chunk-mode) url=%s", blobhubURL)
		runChunkPipeline(ctx, client, blobhubURL, validatorID, privKey, notifChan, dataChan)
		log.Printf("chunk stream closed, reconnecting url=%s", blobhubURL)
		select {
		case <-ctx.Done():
			return
		case <-time.After(reconnectBackoff):
		}
	}
}

func runChunkPipeline(ctx context.Context, client *blobclient.Client, blobhubURL, validatorID string, privKey ed25519.PrivateKey, notifChan <-chan *blobclient.ChunkNotification, dataChan <-chan []byte) {
	var signed, errors, verified, verifyFailed atomic.Int64

	// Status logging (matches keeper's 10s ticker)
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
				log.Printf("chunk attestation status signed=%d verified=%d verify_failed=%d errors=%d",
					signed.Load(), verified.Load(), verifyFailed.Load(), errors.Load())
			}
		}
	}()

	verifyCh := make(chan chunkWork, 1024)
	attestCh := make(chan blobclient.ChunkAttestation, 1024)
	sendCh := make(chan []blobclient.ChunkAttestation, 64)

	// 16 verify workers
	verifier := &chunkVerifier{blobhubURL: blobhubURL, privKey: privKey}
	var verifyWg sync.WaitGroup
	for i := 0; i < chunkVerifyWorkers; i++ {
		verifyWg.Add(1)
		go func() {
			defer verifyWg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case work, ok := <-verifyCh:
					if !ok {
						return
					}
					att, ok := verifier.verifyChunk(ctx, work)
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
		ticker := time.NewTicker(chunkBatchTicker)
		defer ticker.Stop()
		var batch []blobclient.ChunkAttestation
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
				for len(atts) > chunkMaxBatch {
					part := make([]blobclient.ChunkAttestation, chunkMaxBatch)
					copy(part, atts[:chunkMaxBatch])
					select {
					case sendCh <- part:
					case <-ctx.Done():
						return
					}
					atts = atts[chunkMaxBatch:]
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
	var loggedFirst atomic.Bool
	for i := 0; i < chunkSendWorkers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case atts := <-sendCh:
					results, err := client.SignChunkBatch(ctx, validatorID, atts)
					if err != nil {
						errors.Add(int64(len(atts)))
						if loggedFirst.CompareAndSwap(false, true) {
							log.Printf("SignChunkBatch error: %v", err)
						}
					} else {
						var ok, fail int
						for _, r := range results {
							if r.Status == "ok" || r.Status == "duplicate" {
								ok++
							} else {
								fail++
								if loggedFirst.CompareAndSwap(false, true) {
									log.Printf("attestation rejected: blob_key=%s chunk=%d status=%s error=%s",
										r.BlobKey, r.ChunkIndex, r.Status, r.Error)
								}
							}
						}
						signed.Add(int64(ok))
						errors.Add(int64(fail))
					}
				}
			}
		}()
	}

	// Read notifications and pair with binary data from the stream.
	// When Binary is true, the next item on dataChan is the chunk data for
	// that notification. We MUST consume dataChan to prevent the WebSocket
	// reader goroutine from blocking.
	for {
		select {
		case <-ctx.Done():
			close(verifyCh)
			verifyWg.Wait()
			log.Printf("chunk pipeline shutting down signed=%d verified=%d verify_failed=%d errors=%d",
				signed.Load(), verified.Load(), verifyFailed.Load(), errors.Load())
			return
		case notif, ok := <-notifChan:
			if !ok {
				close(verifyCh)
				verifyWg.Wait()
				log.Printf("chunk notification channel closed signed=%d verified=%d verify_failed=%d errors=%d",
					signed.Load(), verified.Load(), verifyFailed.Load(), errors.Load())
				return
			}
			if notif == nil {
				continue
			}
			work := chunkWork{notif: notif}
			if notif.Binary {
				select {
				case data, ok := <-dataChan:
					if !ok {
						close(verifyCh)
						verifyWg.Wait()
						log.Printf("chunk data channel closed signed=%d verified=%d verify_failed=%d errors=%d",
							signed.Load(), verified.Load(), verifyFailed.Load(), errors.Load())
						return
					}
					work.data = data
				case <-ctx.Done():
					close(verifyCh)
					verifyWg.Wait()
					return
				}
			}
			select {
			case verifyCh <- work:
			case <-ctx.Done():
				return
			}
		}
	}
}

type chunkVerifier struct {
	blobhubURL string
	privKey    ed25519.PrivateKey
}

func (v *chunkVerifier) verifyChunk(ctx context.Context, work chunkWork) (blobclient.ChunkAttestation, bool) {
	msg := work.notif
	var chunkData []byte

	if work.data != nil {
		// Use binary data provided inline via the WebSocket stream.
		chunkData = work.data
	} else {
		// Fallback: download chunk via HTTP GET.
		data, ok := v.fetchChunk(ctx, msg.BlobKey, msg.ChunkIndex)
		if !ok {
			return blobclient.ChunkAttestation{}, false
		}
		chunkData = data
	}

	h := blake2b.Sum256(chunkData)
	computedHash := hex.EncodeToString(h[:])

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
					return blobclient.ChunkAttestation{}, false
				}
			}
		}
	}

	blobKeyBytes, err := hex.DecodeString(msg.BlobKey)
	if err != nil || len(blobKeyBytes) != 32 {
		return blobclient.ChunkAttestation{}, false
	}
	attMsg := make([]byte, 68)
	copy(attMsg[0:32], blobKeyBytes)
	binary.BigEndian.PutUint32(attMsg[32:36], uint32(msg.ChunkIndex))
	copy(attMsg[36:68], h[:])
	sig := ed25519.Sign(v.privKey, attMsg)

	return blobclient.ChunkAttestation{
		BlobKey:    msg.BlobKey,
		ChunkIndex: msg.ChunkIndex,
		ChunkHash:  computedHash,
		Signature:  hex.EncodeToString(sig),
	}, true
}

func (v *chunkVerifier) fetchChunk(ctx context.Context, blobKey string, chunkIndex int) ([]byte, bool) {
	chunkURL := fmt.Sprintf("%s/chunks/%s/%d", v.blobhubURL, blobKey, chunkIndex)
	req, err := http.NewRequestWithContext(ctx, "GET", chunkURL, nil)
	if err != nil {
		return nil, false
	}
	resp, err := chunkHTTPClient.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, false
	}
	buf := chunkBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer chunkBufPool.Put(buf)
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return nil, false
	}
	// Copy out of pooled buffer before returning it.
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	return data, true
}

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
