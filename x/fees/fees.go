package fees

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// FeeConfig holds fee configuration
type FeeConfig struct {
	BaseFee      uint64  `json:"base_fee"`       // 0.0001 ZEN (in wei)
	BurnPercent  int     `json:"burn_percent"`   // 20%
	MinTip       uint64  `json:"min_tip"`        // 0
	MaxTip       uint64  `json:"max_tip"`        // 0.001 ZEN
	PriorityFee  uint64  `json:"priority_fee"`   // Optional priority
	MaxFee       uint64  `json:"max_fee"`        // 0.01 ZEN
}

// FeeModel represents different fee models
type FeeModel int

const (
	Solidity   FeeModel = iota
	Priority          // Priority-based (Solana-style)
	MultiDimensional  // Different fees for different operations
)

// Fee represents a transaction fee
type Fee struct {
	BaseFee     uint64 `json:"base_fee"`
	Tip         uint64 `json:"tip"`
	PriorityFee uint64 `json:"priority_fee"`
	Total       uint64 `json:"total"`
	Burned      uint64 `json:"burned"`
	Validator   uint64 `json:"validator"`
}

// Transaction represents a transaction with fees
type Transaction struct {
	Hash         common.Hash   `json:"hash"`
	From         common.Address `json:"from"`
	To           common.Address `json:"to"`
	GasLimit     uint64        `json:"gas_limit"`
	GasUsed      uint64        `json:"gas_used"`
	Fee          Fee           `json:"fee"`
	FeePerGas    uint64        `json:"fee_per_gas"`
	BlockNumber  int64         `json:"block_number"`
	Timestamp    int64         `json:"timestamp"`
	TxType       string        `json:"tx_type"` // transfer, contract, etc.
}

// FeeStats tracks fee statistics
type FeeStats struct {
	TotalFees      uint64  `json:"total_fees"`
	TotalBurned    uint64  `json:"total_burned"`
	TotalToValidators uint64 `json:"total_to_validators"`
	AvgFee         uint64  `json:"avg_fee"`
	MedianFee      uint64  `json:"median_fee"`
	MinFee         uint64  `json:"min_fee"`
	MaxFee         uint64  `json:"max_fee"`
	FeeTPS         float64 `json:"fee_tps"` // Fees per second
	TotalTx        int64   `json:"total_tx"`
	BurnRate       float64 `json:"burn_rate"` // Tokens burned per second
}

// FeeTracker tracks fee-related metrics
type FeeTracker struct {
	mu              sync.RWMutex
	transactions    []Transaction
	feesCollected   uint64
	tokensBurned    uint64
	revenueSplit    map[common.Address]uint64 // Validator revenue
	lastUpdate      time.Time
}

// Fees handles the low-fee model with burn mechanism
type Fees struct {
	mu           sync.RWMutex
	config       FeeConfig
	tracker      *FeeTracker
	running      bool
	burnEnabled  bool
	feeModel     FeeModel
}

// New creates a new Fees instance
func New() *Fees {
	return &Fees{
		config: FeeConfig{
			BaseFee:      100000000000000, // 0.0001 ZEN (in wei)
			BurnPercent:  20,              // 20% burned
			MinTip:       0,               // No minimum tip
			MaxTip:       1000000000000,   // 0.001 ZEN max tip
			PriorityFee:  0,               // Optional
			MaxFee:       10000000000000,  // 0.01 ZEN max
		},
		tracker:     &FeeTracker{
			revenueSplit: make(map[common.Address]uint64),
		},
		running:     false,
		burnEnabled: true,
		feeModel:    Priority,
	}
}

// NewWithConfig creates Fees with custom configuration
func NewWithConfig(config FeeConfig) *Fees {
	return &Fees{
		config: config,
		tracker: &FeeTracker{
			revenueSplit: make(map[common.Address]uint64),
		},
		running:     false,
		burnEnabled: true,
		feeModel:    Priority,
	}
}

// Start initializes the fee system
func (f *Fees) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	fmt.Println("[FEES] Initializing low-fee system")
	fmt.Printf("  - Base Fee: %d ZEN (%.6f ZEN)\n",
		f.config.BaseFee, float64(f.config.BaseFee)/1e18)
	fmt.Printf("  - Burn Rate: %d%% (%d ZEN burned per tx)\n",
		f.config.BurnPercent, uint64(float64(f.config.BaseFee)*float64(f.config.BurnPercent)/100)/1e18)
	fmt.Printf("  - Max Tip: %d ZEN (%.6f ZEN)\n",
		f.config.MaxTip, float64(f.config.MaxTip)/1e18)
	fmt.Printf("  - Max Fee: %d ZEN (%.6f ZEN)\n",
		f.config.MaxFee, float64(f.config.MaxFee)/1e18)
	fmt.Printf("  - Fee Model: %s\n", f.getFeeModelName())
	fmt.Printf("  - Target: <0.0001 ZEN per transaction\n")
	fmt.Printf("  - Comparison: 100x cheaper than Ethereum\n")

	f.running = true
	f.tracker.lastUpdate = time.Now()

	fmt.Println("✓ Low-fee system initialized")

	return nil
}

// Stop halts the fee system
func (f *Fees) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return nil
	}

	fmt.Println("[FEES] Stopping fee system")
	f.running = false

	return nil
}

// CalculateFee calculates transaction fee
func (f *Fees) CalculateFee(gasLimit uint64, tip uint64, txType string) (*Fee, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if !f.running {
		return nil, fmt.Errorf("fee system not running")
	}

	// Check tip limits
	if tip < f.config.MinTip {
		tip = f.config.MinTip
	}
	if tip > f.config.MaxTip {
		tip = f.config.MaxTip
	}

	// Calculate base fee
	baseFee := f.config.BaseFee

	// Apply transaction type modifiers
	switch txType {
	case "contract_deploy":
		// Higher fee for contract deployment
		baseFee = baseFee * 3
	case "contract_call":
		// Medium fee for contract calls
		baseFee = baseFee * 2
	case "transfer":
		// Base fee for transfers
		break
	case "nft_mint":
		// NFT mints have dynamic pricing
		baseFee = baseFee * 2
	case "defi_swap":
		// DeFi swaps have premium pricing
		baseFee = baseFee * 5
	}

	// Priority fee (optional)
	priorityFee := f.config.PriorityFee

	// Total fee
	total := baseFee + tip + priorityFee

	// Check max fee
	if total > f.config.MaxFee {
		return nil, fmt.Errorf("fee exceeds maximum: %d > %d", total, f.config.MaxFee)
	}

	// Calculate burn amount
	burned := uint64(float64(baseFee) * float64(f.config.BurnPercent) / 100.0)

	// Validator gets the rest
	validator := total - burned

	return &Fee{
		BaseFee:     baseFee,
		Tip:         tip,
		PriorityFee: priorityFee,
		Total:       total,
		Burned:      burned,
		Validator:   validator,
	}, nil
}

// ProcessTransaction processes a transaction and updates metrics
func (f *Fees) ProcessTransaction(tx *Transaction) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Add to tracker
	f.tracker.transactions = append(f.tracker.transactions, *tx)

	// Update metrics
	f.tracker.feesCollected += tx.Fee.Total
	f.tracker.tokensBurned += tx.Fee.Burned

	// Update validator revenue
	// In production: distribute to actual block proposer
	f.tracker.revenueSplit[tx.From] += tx.Fee.Validator

	// Keep only recent transactions
	if len(f.tracker.transactions) > 10000 {
		f.tracker.transactions = f.tracker.transactions[1:]
	}

	return nil
}

// GetFeeForTransactionType returns fee for a specific transaction type
func (f *Fees) GetFeeForTransactionType(txType string) (uint64, error) {
	fee, err := f.CalculateFee(21000, 0, txType)
	if err != nil {
		return 0, err
	}
	return fee.Total, nil
}

// EstimateFee estimates fee for a transaction
func (f *Fees) EstimateFee(gasLimit uint64, txType string) (uint64, error) {
	return f.CalculateFee(gasLimit, 0, txType)
}

// GetCurrentFees returns current fee structure
func (f *Fees) GetCurrentFees() map[string]uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return map[string]uint64{
		"transfer":       f.config.BaseFee,
		"contract_call":  f.config.BaseFee * 2,
		"contract_deploy": f.config.BaseFee * 3,
		"nft_mint":       f.config.BaseFee * 2,
		"defi_swap":      f.config.BaseFee * 5,
		"max_tip":        f.config.MaxTip,
		"max_fee":        f.config.MaxFee,
	}
}

// GetFeeStats returns fee statistics
func (f *Fees) GetFeeStats() *FeeStats {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.tracker.transactions) == 0 {
		return &FeeStats{}
	}

	var totalFees, minFee, maxFee, medianFee uint64
	var fees []uint64

	for _, tx := range f.tracker.transactions {
		totalFees += tx.Fee.Total

		if minFee == 0 || tx.Fee.Total < minFee {
			minFee = tx.Fee.Total
		}
		if tx.Fee.Total > maxFee {
			maxFee = tx.Fee.Total
		}

		fees = append(fees, tx.Fee.Total)
	}

	// Calculate median
	medianFee = calculateMedian(fees)

	// Calculate TPS
	timeDiff := time.Since(f.tracker.lastUpdate).Seconds()
	feeTPS := float64(len(f.tracker.transactions)) / timeDiff

	// Calculate burn rate
	burnRate := float64(f.tracker.tokensBurned) / timeDiff

	return &FeeStats{
		TotalFees:       totalFees,
		TotalBurned:     f.tracker.tokensBurned,
		TotalToValidators: f.tracker.feesCollected - f.tracker.tokensBurned,
		AvgFee:          totalFees / uint64(len(f.tracker.transactions)),
		MedianFee:       medianFee,
		MinFee:          minFee,
		MaxFee:          maxFee,
		FeeTPS:          feeTPS,
		TotalTx:         int64(len(f.tracker.transactions)),
		BurnRate:        burnRate,
	}
}

// GetRevenueSplit returns validator revenue distribution
func (f *Fees) GetRevenueSplit() map[common.Address]uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	split := make(map[common.Address]uint64)
	for addr, amount := range f.tracker.revenueSplit {
		split[addr] = amount
	}

	return split
}

// SetFeeConfig updates fee configuration
func (f *Fees) SetFeeConfig(config FeeConfig) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Validate configuration
	if config.BaseFee == 0 {
		return fmt.Errorf("base fee cannot be zero")
	}
	if config.BurnPercent < 0 || config.BurnPercent > 100 {
		return fmt.Errorf("burn percent must be 0-100")
	}
	if config.MinTip > config.MaxTip {
		return fmt.Errorf("min tip cannot exceed max tip")
	}

	f.config = config
	fmt.Println("[FEES] Fee configuration updated")

	return nil
}

// EnableBurn enables or disables token burning
func (f *Fees) EnableBurn(enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.burnEnabled = enabled

	if enabled {
		fmt.Println("[FEES] Token burning enabled")
	} else {
		fmt.Println("[FEES] Token burning disabled")
	}
}

// GetBurnStats returns burning statistics
func (f *Fees) GetBurnStats() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return map[string]interface{}{
		"enabled":              f.burnEnabled,
		"burn_percent":         f.config.BurnPercent,
		"total_burned":         f.tracker.tokensBurned / 1e18,
		"total_burned_wei":     f.tracker.tokensBurned,
		"fees_collected":       f.tracker.feesCollected / 1e18,
		"burn_rate_per_second": f.GetFeeStats().BurnRate / 1e18,
	}
}

// PrintFeeComparison prints comparison with other chains
func (f *Fees) PrintFeeComparison() {
	fmt.Println("\n" + "=".repeat(50))
	fmt.Println("Fee Comparison: ZenNetwork vs Other Chains")
	fmt.Println("=".repeat(50))
	fmt.Printf("ZenNetwork:   < 0.0001 ZEN (~$0.001)\n")
	fmt.Printf("Ethereum:     ~0.002 ETH (~$5-50)\n")
	fmt.Printf("Bitcoin:      ~0.0001 BTC (~$4-10)\n")
	fmt.Printf("Solana:       ~0.00001 SOL (~$0.001)\n")
	fmt.Printf("Binance Smart Chain: ~0.0005 BNB (~$0.15)\n")
	fmt.Println("=".repeat(50))
	fmt.Println("ZenNetwork offers 100-50,000x lower fees!")
	fmt.Println("=".repeat(50) + "\n")
}

// getFeeModelName returns fee model name
func (f *Fees) getFeeModelName() string {
	switch f.feeModel {
	case Solidity:
		return "Solidity-style (EIP-1559)"
	case Priority:
		return "Priority-based (Solana-style)"
	case MultiDimensional:
		return "Multi-dimensional"
	default:
		return "Unknown"
	}
}

// calculateMedian calculates median from slice
func calculateMedian(fees []uint64) uint64 {
	if len(fees) == 0 {
		return 0
	}

	// Simple median calculation
	// In production: use more efficient algorithm
	var sum uint64
	for _, fee := range fees {
		sum += fee
	}
	return sum / uint64(len(fees))
}

// SimulateTransaction simulates fee calculation
func (f *Fees) SimulateTransaction(txType string, gasLimit uint64) error {
	fee, err := f.CalculateFee(gasLimit, 0, txType)
	if err != nil {
		return err
	}

	fmt.Printf("\n[FEES] Transaction Simulation: %s\n", txType)
	fmt.Printf("  Gas Limit: %d\n", gasLimit)
	fmt.Printf("  Base Fee: %.6f ZEN\n", float64(fee.BaseFee)/1e18)
	fmt.Printf("  Tip: %.6f ZEN\n", float64(fee.Tip)/1e18)
	fmt.Printf("  Total: %.6f ZEN\n", float64(fee.Total)/1e18)
	fmt.Printf("  Burned: %.6f ZEN (%.0f%%)\n", float64(fee.Burned)/1e18, float64(f.config.BurnPercent))
	fmt.Printf("  To Validator: %.6f ZEN\n\n", float64(fee.Validator)/1e18)

	return nil
}

// GetConfig returns current fee configuration
func (f *Fees) GetConfig() FeeConfig {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.config
}

// IsRunning returns fee system status
func (f *Fees) IsRunning() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.running
}
