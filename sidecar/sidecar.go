package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"go.uber.org/zap"
)

var _ Sidecar = (*SidecarImpl)(nil)

// Oracle defines the expected interface for an oracle. It is consumed by the oracle server.
//
//go:generate mockery --name Sidecar --filename mock_sidecar.go
type Sidecar interface {
	IsRunning() bool
	QueryBlockEvents(blockNumber *big.Int) ([]GenericEvent, error)
	Start(ctx context.Context) error
	Stop()
}

// SidecarImpl is the core component responsible for fetching ethereum blocks and processing events.
type SidecarImpl struct {
	logger          *zap.Logger
	client          *ethclient.Client
	contractAddress common.Address
	contractABI     abi.ABI
	startBlock      *big.Int
	blocksMap       map[string]*EthereumBlock
	mu              sync.Mutex
	updateInterval  time.Duration

	// running is the current status of the main oracle process (running or not).
	running atomic.Bool
}

// NewSidecar creates a new Oracle instance.
func NewSidecar(
	client *ethclient.Client,
	contractAddress common.Address,
	contractAbi abi.ABI,
	startBlock *big.Int,
	logger *zap.Logger,
) *SidecarImpl {
	return &SidecarImpl{
		client:          client,
		contractAddress: contractAddress,
		contractABI:     contractAbi,
		startBlock:      startBlock,
		logger:          logger,
		blocksMap:       make(map[string]*EthereumBlock),
		updateInterval:  10 * time.Second, // Adjust the interval as needed
	}
}

// Start begins the process of querying and storing events from the Ethereum blockchain.
func (o *SidecarImpl) Start(ctx context.Context) error {

	// Initial check to verify Ethereum client connectivity and log fetching capability
	query := ethereum.FilterQuery{
		FromBlock: o.startBlock,
		ToBlock:   o.startBlock,
		Addresses: []common.Address{o.contractAddress},
	}

	// Attempt to fetch logs as a connectivity and configuration check
	_, err := o.client.FilterLogs(ctx, query)
	if err != nil {
		o.logger.Error("Failed to fetch logs for initial check", zap.Error(err))
		return err
	}

	o.running.Store(true)
	go o.queryAndStoreEvents(ctx)

	// Side car started succesfully
	return nil
}

// Stop signals the oracle to stop processing.
func (o *SidecarImpl) Stop() {
	o.running.Store(false)
}

// IsRunning checks if the oracle process is currently running.
// It returns true if the oracle is running, false otherwise.
func (o *SidecarImpl) IsRunning() bool {
	return o.running.Load()
}

// QueryBlockEvents queries the `blocksMap` for events associated with a specific block number.
func (o *SidecarImpl) QueryBlockEvents(blockNumber *big.Int) ([]GenericEvent, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	blockHash := blockNumber.String() // Assuming block number is used directly as key; adjust if necessary

	block, exists := o.blocksMap[blockHash]
	if !exists {
		return nil, fmt.Errorf("no events found for block number %s", blockNumber)
	}

	// Assuming your EthereumBlock structure has a slice of Events
	// You might need to convert or map these events to your expected return type.
	return block.Events, nil
}

// queryAndStoreEvents continuously fetches logs from the Ethereum blockchain and processes them.
func (o *SidecarImpl) queryAndStoreEvents(ctx context.Context) {
	lastQueriedBlock := o.startBlock
	ticker := time.NewTicker(o.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !o.IsRunning() {
				return
			}

			query := ethereum.FilterQuery{
				FromBlock: lastQueriedBlock,
				Addresses: []common.Address{o.contractAddress},
			}

			logs, err := o.client.FilterLogs(context.Background(), query)
			if err != nil {
				o.logger.Error("Error fetching logs", zap.Error(err))
				continue
			}

			if len(logs) > 0 {
				lastLogBlock := logs[len(logs)-1].BlockNumber
				lastQueriedBlock = new(big.Int).SetUint64(lastLogBlock + 1)
			}

			o.mu.Lock()
			for _, vLog := range logs {
				processAndStoreLog(vLog, o.contractABI, o.blocksMap)
			}
			o.mu.Unlock()
		}
	}
}
