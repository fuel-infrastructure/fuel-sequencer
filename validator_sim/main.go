// validator_sim runs the off-consensus part of x/blob: stream blobs (or chunk
// notifications) from blobhub, sync to blobpool or attest chunks, and ack/sign
// back to blobhub. Two modes: whole-blob (sync → blobpool, sign blob key) and
// chunk-mode (stream chunks → verify → sign chunk attestations). No
// MsgBlobMetadataTx or blobpool HTTP server.
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	blobclient "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/blob-storage/pkg/store/sqlite"
)

const (
	reconnectBackoff = 5 * time.Second
	signMaxRetries  = 3
	signRetryDelay  = 5 * time.Second
)

func main() {
	chunkMode := flag.Bool("chunk-mode", false, "chunk-mode attestation: stream chunk notifications, verify and sign chunks only")
	chunkIndex := flag.Int("index", 0, "chunk index this validator attests (used when -chunk-mode)")
	blobpoolPath := flag.String("blobpool", "./data/blobpool.db", "SQLite path for blob store (whole-blob mode only)")
	keyPath := flag.String("key", "", "path to Ed25519 private key file (hex or raw 32 bytes)")
	validatorIDFlag := flag.String("validator-id", "", "validator ID (fallback when not using positionals)")
	blobhubFlag := flag.String("blobhub", "http://localhost:31035", "blobhub URL (fallback when not using positionals)")
	flag.Parse()

	// Positionals: first two non-flag args = validator-id, blobhub URL (script style)
	var validatorID, blobhubURL string
	if args := flag.Args(); len(args) >= 2 {
		validatorID = args[0]
		blobhubURL = args[1]
	} else {
		validatorID = *validatorIDFlag
		blobhubURL = *blobhubFlag
	}

	if *chunkMode && validatorID == "" {
		log.Fatal("validator ID is required when -chunk-mode is set (use positionals or -validator-id)")
	}

	blobhubURL = normalizeBlobhubAddress(blobhubURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		cancel()
	}()

	client := blobclient.NewClient(&blobclient.ClientConfig{
		BaseURL: blobhubURL,
		Logger:  nil,
	})

	var privKey ed25519.PrivateKey
	var pubKey ed25519.PublicKey
	if validatorID != "" {
		var err error
		privKey, pubKey, err = loadOrGenerateKey(*keyPath)
		if err != nil {
			log.Fatalf("key: %v", err)
		}
		if err := client.RegisterWithKey(ctx, validatorID, pubKey); err != nil {
			log.Fatalf("register validator: %v", err)
		}
		log.Printf("registered validator with blobhub validator_id=%s", validatorID)
	}

	if *chunkMode {
		runChunkMode(ctx, client, blobhubURL, validatorID, *chunkIndex, privKey)
		return
	}

	// Whole-blob mode
	store, err := sqlite.Store(ctx, *blobpoolPath, false)
	if err != nil {
		log.Fatalf("blobpool: %v", err)
	}
	store.Prune(ctx, 20*time.Minute)
	runWholeBlobSync(ctx, client, store, blobhubURL, validatorID, privKey)
}

func normalizeBlobhubAddress(address string) string {
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		return address
	}
	return "http://" + address
}

func runWholeBlobSync(ctx context.Context, client *blobclient.Client, st store.Store, blobhubURL, validatorID string, _ ed25519.PrivateKey) {
	log.Printf("starting blob sync url=%s", blobhubURL)
	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			log.Print("context cancelled, stopping sync")
			return
		default:
		}
		if attempt > 1 {
			log.Printf("blobhub reconnect attempt attempt=%d url=%s", attempt, blobhubURL)
		}
		blobChan, err := client.StreamBlobs(ctx)
		if err != nil {
			log.Printf("failed to connect to blobhub error=%v url=%s attempt=%d", err, blobhubURL, attempt)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectBackoff):
			}
			continue
		}
		log.Printf("connected to blobhub url=%s", blobhubURL)
	stream:
		for {
			select {
			case <-ctx.Done():
				return
			case blob, ok := <-blobChan:
				if !ok {
					log.Printf("blob stream channel closed, reconnecting url=%s", blobhubURL)
					break stream
				}
				if blob == nil {
					continue
				}
				has, err := st.Has(ctx, blob.Key)
				if err != nil {
					log.Printf("store Has error: %v", err)
					continue
				}
				if has {
					continue
				}
				if _, err := st.Put(ctx, blob.Data); err != nil {
					log.Printf("store Put error: %v", err)
					continue
				}
				log.Printf("stored blob id=%s size=%d", blob.Key.String(), len(blob.Data))
				if validatorID != "" {
					go signBlobAsync(ctx, client, validatorID, blob.Key)
				}
			}
		}
	}
}

func signBlobAsync(ctx context.Context, client *blobclient.Client, validatorID string, key store.Key) {
	if err := signBlob(ctx, client, validatorID, key); err != nil {
		log.Printf("async signature failed key=%s error=%v", key.String(), err)
	}
}

func signBlob(ctx context.Context, client *blobclient.Client, validatorID string, key store.Key) error {
	var lastErr error
	for attempt := 0; attempt < signMaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(signRetryDelay):
			}
		}
		err := client.Sign(ctx, validatorID, key)
		if err == nil {
			log.Printf("signature submitted key=%s validator_id=%s", key.String(), validatorID)
			return nil
		}
		lastErr = err
		log.Printf("failed to submit signature key=%s attempt=%d error=%v", key.String(), attempt+1, err)
	}
	return lastErr
}

func loadOrGenerateKey(keyPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if keyPath != "" {
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, nil, fmt.Errorf("read key file: %w", err)
		}
		raw := data
		if hex.DecodedLen(len(data)) == 32 {
			raw = make([]byte, 32)
			_, err = hex.Decode(raw, bytesTrimSpace(data))
			if err != nil {
				return nil, nil, fmt.Errorf("decode hex key: %w", err)
			}
		} else {
			raw = bytesTrimSpace(data)
		}
		if len(raw) != 32 {
			return nil, nil, fmt.Errorf("key must be 32 bytes (got %d)", len(raw))
		}
		priv := ed25519.NewKeyFromSeed(raw)
		pub := priv.Public().(ed25519.PublicKey)
		return priv, pub, nil
	}
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	// Seed is first 32 bytes of standard 64-byte Ed25519 private key
	seed := priv
	if len(seed) > 32 {
		seed = priv[:32]
	}
	log.Printf("generated Ed25519 key (save for reuse): %s", hex.EncodeToString(seed))
	return priv, pub, nil
}

func bytesTrimSpace(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == ' ' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}
