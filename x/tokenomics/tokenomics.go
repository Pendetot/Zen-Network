package tokenomics

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Distribution represents token distribution
type Distribution struct {
	Category          string  `json:"category"`
	AllocationPercent float64 `json:"allocation_percent"`
	Amount            string  `json:"amount"` // in wei
	Locked            bool    `json:"locked"`
	UnlockDate        int64   `json:"unlock_date"`
	Address           common.Address `json:"address"`
}

// TotalSupply represents the fixed total supply
type TotalSupply struct {
	Fixed     bool   `json:"fixed"`
	Immutable bool   `json:"immutable"`
	Amount    string `json:"amount"` // 1,000,000,000 ZEN
}

// BurnEvent represents a token burn event
type BurnEvent struct {
	Amount    string `json:"amount"`
	TxHash    common.Hash `json:"tx_hash"`
	Reason    string `json:"reason"`
	Timestamp int64  `json:"timestamp"`
	Block     int64  `json:"block"`
}

// Tokenomics holds the complete tokenomics configuration
type Tokenomics struct {
	mu           sync.RWMutex
	totalSupply  TotalSupply
	distributions []Distribution
	burnEvents   []BurnEvent
	minting      MintingConfig
}

// MintingConfig holds minting configuration
type MintingConfig struct {
	Enabled        bool    `json:"enabled"`
	MaxSupply      string  `json:"max_supply"`
	InflationRate  float64 `json:"inflation_rate"`
	HardCapped     bool    `json:"hard_capped"`
}

// New creates a new tokenomics instance with fixed 1B ZEN supply
func New() *Tokenomics {
	return &Tokenomics{
		totalSupply: TotalSupply{
			Fixed:     true,
			Immutable: true,
			Amount:    "1000000000000000000000000000", // 1B ZEN (1e27 wei)
		},
		distributions: getInitialDistributions(),
		burnEvents:    make([]BurnEvent, 0),
		minting: MintingConfig{
			Enabled:        false, // Minting disabled
			MaxSupply:      "1000000000000000000000000000",
			InflationRate:  0.0, // No inflation
			HardCapped:     true,
		},
	}
}

// ValidateSupply validates the total supply is exactly 1B ZEN
func (t *Tokenomics) ValidateSupply() error {
	if !t.totalSupply.Fixed {
		return fmt.Errorf("supply must be fixed")
	}
	if !t.totalSupply.Immutable {
		return fmt.Errorf("supply must be immutable")
	}

	// Check if supply equals exactly 1B ZEN
	expected := "1000000000000000000000000000"
	if t.totalSupply.Amount != expected {
		return fmt.Errorf("invalid total supply: %s (expected: %s)", t.totalSupply.Amount, expected)
	}

	return nil
}

// GetTotalSupply returns total supply
func (t *Tokenomics) GetTotalSupply() TotalSupply {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.totalSupply
}

// GetDistributions returns all token distributions
func (t *Tokenomics) GetDistributions() []Distribution {
	t.mu.RLock()
	defer t.mu.RUnlock()

	distributions := make([]Distribution, len(t.distributions))
	copy(distributions, t.distributions)
	return distributions
}

// GetSupplyByCategory returns supply for a specific category
func (t *Tokenomics) GetSupplyByCategory(category string) (Distribution, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, dist := range t.distributions {
		if dist.Category == category {
			return dist, nil
		}
	}

	return Distribution{}, fmt.Errorf("category not found: %s", category)
}

// BurnTokens burns tokens (fee burning mechanism)
func (t *Tokenomics) BurnTokens(amount string, txHash common.Hash, reason string, block int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Parse amount
	// In production: use big.Int for precise calculation
	// For now: just record the event

	event := BurnEvent{
		Amount:    amount,
		TxHash:    txHash,
		Reason:    reason,
		Timestamp: time.Now().Unix(),
		Block:     block,
	}

	t.burnEvents = append(t.burnEvents, event)

	// Keep only last 10000 burn events
	if len(t.burnEvents) > 10000 {
		t.burnEvents = t.burnEvents[1:]
	}

	return nil
}

// GetBurnEvents returns all burn events
func (t *Tokenomics) GetBurnEvents(limit int) []BurnEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if limit <= 0 || limit > len(t.burnEvents) {
		limit = len(t.burnEvents)
	}

	events := make([]BurnEvent, limit)
	start := len(t.burnEvents) - limit
	copy(events, t.burnEvents[start:])

	return events
}

// GetMintingConfig returns minting configuration
func (t *Tokenomics) GetMintingConfig() MintingConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.minting
}

// IsMintingEnabled checks if minting is enabled
func (t *Tokenomics) IsMintingEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.minting.Enabled
}

// AttemptMint attempts to mint tokens (should always fail)
func (t *Tokenomics) AttemptMint(amount string, to common.Address) error {
	if t.IsMintingEnabled() {
		return fmt.Errorf("minting is disabled in tokenomics")
	}

	// CRITICAL: Prevent any minting attempts
	return fmt.Errorf("ZEN token supply is fixed and immutable (1,000,000,000 ZEN). Minting is permanently disabled")
}

// GetCirculatingSupply calculates circulating supply
func (t *Tokenomics) GetCirculatingSupply() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Start with total supply
	circulating := t.totalSupply.Amount

	// Subtract locked amounts
	// In production: calculate from actual locked amounts
	// For now: estimate

	return circulating
}

// GetBurnStats returns burning statistics
func (t *Tokenomics) GetBurnStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var totalBurned float64
	now := time.Now()

	for _, event := range t.burnEvents {
		// In production: parse actual amount
		totalBurned += 0 // Mock
	}

	return map[string]interface{}{
		"total_events":   len(t.burnEvents),
		"total_burned":   totalBurned,
		"burn_rate":      totalBurned / float64(len(t.burnEvents)),
		"last_burn":      time.Unix(t.burnEvents[len(t.burnEvents)-1].Timestamp, 0).Unix(),
	}
}

// PrintSummary prints tokenomics summary
func (t *Tokenomics) PrintSummary() {
	fmt.Println("\n" + "=".repeat(60))
	fmt.Println("ZenNetwork Tokenomics Summary (ZEN)")
	fmt.Println("=".repeat(60))
	fmt.Printf("Total Supply: 1,000,000,000 ZEN (Fixed & Immutable)\n")
	fmt.Printf("Decimals: 18\n")
	fmt.Printf("Minting: DISABLED (Hard-capped)\n")
	fmt.Printf("Burning: ENABLED (20%% of all fees)\n")
	fmt.Println("=".repeat(60))
	fmt.Println("\nDistribution:")

	for _, dist := range t.distributions {
		fmt.Printf("  %-20s: %8.1f%%  (%s ZEN)\n",
			dist.Category, dist.AllocationPercent, t.formatAmount(dist.Amount))
	}
	fmt.Println("=".repeat(60))
	fmt.Println("\nHalving System:")
	fmt.Println("  - Adaptive Exponential Halving (AEH)")
	fmt.Println("  - Total Reward Pool: 200M ZEN")
	fmt.Println("  - Initial Reward: 1000 ZEN/block")
	fmt.Println("  - Reduction: 5% per quarter")
	fmt.Println("  - Habis: ~2033")
	fmt.Println("=".repeat(60))
}

// formatAmount formats amount for display
func (t *Tokenomics) formatAmount(amount string) string {
	// Simple formatting
	return fmt.Sprintf("%s", amount)
}

// GetAllocationPercent returns allocation percentage for category
func (t *Tokenomics) GetAllocationPercent(category string) (float64, error) {
	dist, err := t.GetSupplyByCategory(category)
	if err != nil {
		return 0, err
	}
	return dist.AllocationPercent, nil
}

// IsSupplyFixed checks if supply is fixed
func (t *Tokenomics) IsSupplyFixed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.totalSupply.Fixed
}

// IsSupplyImmutable checks if supply is immutable
func (t *Tokenomics) IsSupplyImmutable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.totalSupply.Immutable
}

// getInitialDistributions returns the initial token distribution
func getInitialDistributions() []Distribution {
	return []Distribution{
		{
			Category:          "Community",
			AllocationPercent: 40.0,
			Amount:            "400000000000000000000000000", // 40%
			Locked:            true,
			UnlockDate:        0, // Immediately available
			Address:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		{
			Category:          "Team",
			AllocationPercent: 20.0,
			Amount:            "200000000000000000000000000", // 20%
			Locked:            true,
			UnlockDate:        time.Now().Add(4 * 365 * 24 * time.Hour).Unix(), // 4 years
			Address:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		{
			Category:          "Ecosystem",
			AllocationPercent: 20.0,
			Amount:            "200000000000000000000000000", // 20%
			Locked:            true,
			UnlockDate:        0, // Distributed via rewards
			Address:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		{
			Category:          "Liquidity",
			AllocationPercent: 10.0,
			Amount:            "100000000000000000000000000", // 10%
			Locked:            true,
			UnlockDate:        0,
			Address:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		{
			Category:          "Foundation",
			AllocationPercent: 10.0,
			Amount:            "100000000000000000000000000", // 10%
			Locked:            true,
			UnlockDate:        time.Now().Add(2 * 365 * 24 * time.Hour).Unix(), // 2 years
			Address:           common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
	}
}
