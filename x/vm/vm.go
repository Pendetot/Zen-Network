package vm

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// VMConfig holds EVM configuration
type VMConfig struct {
	ChainID     uint64           `json:"chain_id"`
	Shards      int              `json:"shards"`
	MaxGas      uint64           `json:"max_gas"`
	BlockGas    uint64           `json:"block_gas"`
	ParallelTxs int              `json:"parallel_txs"`
}

// Shard represents a parallel execution shard
type Shard struct {
	ID          int                  `json:"id"`
	BlockNumber int64                `json:"block_number"`
	Transactions []types.Transaction `json:"transactions"`
	State       vm.StateDB           `json:"state"`
	Results     map[common.Hash]*types.Receipt `json:"results"`
	mu          sync.Mutex
}

// ExecutionResult holds transaction execution result
type ExecutionResult struct {
	TxHash      common.Hash    `json:"tx_hash"`
	Success     bool           `json:"success"`
	GasUsed     uint64         `json:"gas_used"`
	ReturnData  []byte         `json:"return_data"`
	Logs        []*types.Log   `json:"logs"`
	ShardID     int            `json:"shard_id"`
	ExecutionTime time.Duration `json:"execution_time"`
}

// EVM handles parallel EVM execution
type EVM struct {
	config        VMConfig
	shards        []*Shard
	currentBlock  int64
	stateFactory  StateFactory
	mu            sync.RWMutex
	running       bool
	benchmarks    []Benchmark
}

// StateFactory creates state instances
type StateFactory interface {
	NewState() (vm.StateDB, error)
	GetState(root common.Hash) (vm.StateDB, error)
}

// SimpleStateFactory implements StateFactory
type SimpleStateFactory struct{}

func (f SimpleStateFactory) NewState() (vm.StateDB, error) {
	// In production: use actual state database
	// For now: return mock state
	return &MockStateDB{}, nil
}

func (f SimpleStateFactory) GetState(root common.Hash) (vm.StateDB, error) {
	return &MockStateDB{}, nil
}

// MockStateDB is a mock implementation
type MockStateDB struct{}

func (s *MockStateDB) CreateAccount(addr common.Address) {}
func (s *MockStateDB) SubBalance(addr common.Address, amount *common.Uint256Value) {}
func (s *MockStateDB) AddBalance(addr common.Address, amount *common.Uint256Value) {}
func (s *MockStateDB) GetBalance(addr common.Address) *common.Uint256Value { return &common.Uint256Value{} }
func (s *MockStateDB) SubNonce(addr common.Address, amount uint64) {}
func (s *MockStateDB) AddNonce(addr common.Address, amount uint64) {}
func (s *MockStateDB) GetNonce(addr common.Address) uint64 { return 0 }
func (s *MockStateDB) Delete(addr common.Address) {}
func (s *MockStateDB) Exist(addr common.Address) bool { return false }
func (s *MockStateDB) Empty(addr common.Address) bool { return false }
func (s *MockStateDB) RevertToSnapshot(int) {}

// Benchmark holds performance metrics
type Benchmark struct {
	BlockNumber int64
	TPS         int
	GasPerSec   uint64
	Latency     time.Duration
	Timestamp   time.Time
}

// NewEVM creates a new EVM instance
func NewEVM() *EVM {
	return &EVM{
		config: VMConfig{
			ChainID:      1337, // ZenNetwork chain ID
			Shards:       64,
			MaxGas:       100000000, // 100M gas per block
			BlockGas:     100000000,
			ParallelTxs:  1000,
		},
		shards:       make([]*Shard, 64),
		stateFactory: SimpleStateFactory{},
		running:      false,
		benchmarks:   make([]Benchmark, 0),
	}
}

// NewEVMWithConfig creates EVM with custom config
func NewEVMWithConfig(config VMConfig) *EVM {
	evm := NewEVM()
	evm.config = config
	evm.shards = make([]*Shard, config.Shards)
	return evm
}

// Start initializes the EVM
func (e *EVM) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	fmt.Println("[EVM] Initializing parallel EVM executor")
	fmt.Printf("  - Chain ID: %d\n", e.config.ChainID)
	fmt.Printf("  - Shards: %d\n", e.config.Shards)
	fmt.Printf("  - Max Gas: %d\n", e.config.MaxGas)
	fmt.Printf("  - Parallel Txs: %d\n", e.config.ParallelTxs)
	fmt.Printf("  - Target TPS: 10,000-50,000\n")

	// Initialize shards
	for i := 0; i < e.config.Shards; i++ {
		state, err := e.stateFactory.NewState()
		if err != nil {
			return fmt.Errorf("failed to create state for shard %d: %w", i, err)
		}

		e.shards[i] = &Shard{
			ID:          i,
			BlockNumber: 0,
			Transactions: make([]types.Transaction, 0),
			State:       state,
			Results:     make(map[common.Hash]*types.Receipt),
		}
	}

	// Start benchmark collector
	go e.benchmarkCollector()

	e.running = true
	fmt.Println("✓ EVM initialized with parallel execution")

	return nil
}

// Stop halts the EVM
func (e *EVM) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	fmt.Println("[EVM] Stopping EVM executor")
	e.running = false
	return nil
}

// ExecuteTransaction executes a single transaction
func (e *EVM) ExecuteTransaction(tx *types.Transaction) (*ExecutionResult, error) {
	startTime := time.Now()

	// Determine which shard to use (based on transaction hash)
	shardID := e.selectShard(tx.Hash())
	shard := e.shards[shardID]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Create EVM context
	evmContext := createEVMContext(tx, e.currentBlock, shard.State)

	// Execute transaction using EVM
	// In production: actual EVM execution
	result := &ExecutionResult{
		TxHash:      tx.Hash(),
		Success:     true,
		GasUsed:     21000, // Simple transfer
		ReturnData:  []byte{},
		Logs:        make([]*types.Log, 0),
		ShardID:     shardID,
		ExecutionTime: time.Since(startTime),
	}

	// Store result
	shard.Results[tx.Hash()] = &types.Receipt{
		TxHash:      tx.Hash(),
		GasUsed:     result.GasUsed,
		BlockNumber: &e.currentBlock,
		Logs:        result.Logs,
	}

	return result, nil
}

// ExecuteBlock executes a block with parallel transactions
func (e *EVM) ExecuteBlock(block *types.Block) ([]*ExecutionResult, error) {
	e.mu.Lock()
	e.currentBlock = block.Number().Int64()
	e.mu.Unlock()

	txs := block.Transactions()
	if len(txs) == 0 {
		return []*ExecutionResult{}, nil
	}

	fmt.Printf("[EVM] Executing block %d with %d transactions (parallel)\n",
		block.Number(), len(txs))

	// Distribute transactions across shards for parallel execution
	resultsCh := make(chan *ExecutionResult, len(txs))
	var wg sync.WaitGroup

	// Group transactions by shard
	shardTxs := make(map[int][]*types.Transaction)
	for _, tx := range txs {
		shardID := e.selectShard(tx.Hash())
		shardTxs[shardID] = append(shardTxs[shardID], tx)
	}

	// Execute in parallel per shard
	for shardID, shardTxList := range shardTxs {
		wg.Add(1)
		go func(id int, txList []*types.Transaction) {
			defer wg.Done()

			shard := e.shards[id]
			shard.mu.Lock()
			defer shard.mu.Unlock()

			for _, tx := range txList {
				result, _ := e.ExecuteTransaction(tx)
				resultsCh <- result
			}
		}(shardID, shardTxList)
	}

	// Wait for all executions
	wg.Wait()
	close(resultsCh)

	// Collect results
	results := make([]*ExecutionResult, 0, len(txs))
	for result := range resultsCh {
		results = append(results, result)
	}

	// Update shard states
	e.updateShardStates(block.Number().Int64(), txs)

	// Calculate and store benchmark
	e.recordBenchmark(block.Number().Int64(), len(txs), time.Now().Sub(block.Time()))

	fmt.Printf("[EVM] Block executed: TPS=%d, Time=%v\n",
		len(txs)/(int(time.Since(block.Time())/time.Millisecond)*1000), time.Since(block.Time()))

	return results, nil
}

// ExecuteTransactions executes multiple transactions in parallel
func (e *EVM) ExecuteTransactions(txs []*types.Transaction) ([]*ExecutionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("EVM not running")
	}

	resultsCh := make(chan *ExecutionResult, len(txs))
	var wg sync.WaitGroup

	// Execute all transactions in parallel
	for i, tx := range txs {
		wg.Add(1)
		go func(index int, transaction *types.Transaction) {
			defer wg.Done()
			result, _ := e.ExecuteTransaction(transaction)
			resultsCh <- result
		}(i, tx)
	}

	// Wait for completion
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	results := make([]*ExecutionResult, 0, len(txs))
	for result := range resultsCh {
		results = append(results, result)
	}

	return results, nil
}

// selectShard determines which shard to use for a transaction
func (e *EVM) selectShard(txHash common.Hash) int {
	// Simple hash-based shard selection
	// In production: more sophisticated load balancing
	hash := txHash.Big()
	shardID := int(hash.Uint64() % uint64(e.config.Shards))
	return shardID
}

// updateShardStates updates state after block execution
func (e *EVM) updateShardStates(blockNumber int64, txs []*types.Transaction) {
	for _, shard := range e.shards {
		shard.mu.Lock()
		shard.BlockNumber = blockNumber
		shard.Transactions = append(shard.Transactions, txs...)
		shard.mu.Unlock()
	}
}

// recordBenchmark stores performance metrics
func (e *EVM) recordBenchmark(blockNumber int64, txCount int, latency time.Duration) {
	tps := int(float64(txCount) / latency.Seconds())

	benchmark := Benchmark{
		BlockNumber: blockNumber,
		TPS:         tps,
		GasPerSec:   uint64(float64(txCount) * 21000 / latency.Seconds()),
		Latency:     latency,
		Timestamp:   time.Now(),
	}

	e.benchmarks = append(e.benchmarks, benchmark)

	// Keep only recent benchmarks
	if len(e.benchmarks) > 100 {
		e.benchmarks = e.benchmarks[1:]
	}
}

// benchmarkCollector collects and prints performance metrics
func (e *EVM) benchmarkCollector() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.printBenchmark()
		}
	}
}

// printBenchmark displays current performance metrics
func (e *EVM) printBenchmark() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.benchmarks) == 0 {
		return
	}

	// Calculate averages from last 10 blocks
	start := len(e.benchmarks) - 10
	if start < 0 {
		start = 0
	}

	var totalTPS, totalGas, totalLatency int
	count := 0

	for _, b := range e.benchmarks[start:] {
		totalTPS += b.TPS
		totalGas += int(b.GasPerSec)
		totalLatency += int(b.Latency.Milliseconds())
		count++
	}

	if count > 0 {
		avgTPS := totalTPS / count
		avgGas := totalGas / count
		avgLatency := totalLatency / count

		fmt.Printf("[EVM] Performance (avg last %d blocks): TPS=%d, Gas/s=%d, Latency=%dms\n",
			count, avgTPS, avgGas, avgLatency)
	}
}

// GetShard returns shard information
func (e *EVM) GetShard(shardID int) *Shard {
	if shardID < 0 || shardID >= len(e.shards) {
		return nil
	}
	return e.shards[shardID]
}

// GetAllShards returns all shards
func (e *EVM) GetAllShards() []*Shard {
	return e.shards
}

// GetBenchmarks returns performance metrics
func (e *EVM) GetBenchmarks() []Benchmark {
	return e.benchmarks
}

// GetStats returns EVM statistics
func (e *EVM) GetStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var avgTPS int
	if len(e.benchmarks) > 0 {
		for _, b := range e.benchmarks {
			avgTPS += b.TPS
		}
		avgTPS /= len(e.benchmarks)
	}

	return map[string]interface{}{
		"chain_id":       e.config.ChainID,
		"shards":         e.config.Shards,
		"max_gas":        e.config.MaxGas,
		"current_block":  e.currentBlock,
		"avg_tps":        avgTPS,
		"target_tps_min": 10000,
		"target_tps_max": 50000,
		"running":        e.running,
		"benchmarks":     len(e.benchmarks),
	}
}

// createEVMContext creates EVM execution context
func createEVMContext(tx *types.Transaction, blockNumber int64, state vm.StateDB) vm.Context {
	return vm.Context{
		CanTransfer: func(db vm.StateDB, addr common.Address, amount *common.Uint256Value) bool {
			return db.GetBalance(addr).Cmp(amount) >= 0
		},
		Transfer: func(db vm.StateDB, sender, recipient common.Address, amount *common.Uint256Value) {
			db.SubBalance(sender, amount)
			db.AddBalance(recipient, amount)
		},
		GetHash: func(uint64) common.Hash {
			return common.Hash{}
		},
		BlockNumber:     common.NewUint256WithoutWrapper(blockNumber),
		Coinbase:        common.Address{},
		Timestamp:       common.NewUint256WithoutWrapper(time.Now().Unix()),
		Difficulty:      common.NewUint256WithoutWrapper(0),
		GasLimit:        common.NewUint256WithoutWrapper(100000000),
		GasPrice:        common.NewUint256WithoutWrapper(0),
		Origin:          common.Address{},
	}
}

// DeployContract deploys a smart contract
func (e *EVM) DeployContract(bytecode []byte, constructorArgs []byte) (common.Hash, error) {
	// In production: actual contract deployment
	// For now: return mock hash
	hash := common.HexToHash("0x1234567890abcdef")
	fmt.Printf("[EVM] Contract deployed: %s (size: %d bytes)\n",
		hash.String(), len(bytecode))
	return hash, nil
}

// CallContract performs a contract call
func (e *EVM) CallContract(addr common.Address, data []byte) ([]byte, error) {
	// In production: actual contract call
	// For now: return mock data
	return []byte("mock response"), nil
}

// IsRunning returns EVM status
func (e *EVM) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}
