package halving

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// HalvingPhase represents the current halving phase
type HalvingPhase struct {
	Phase            int       `json:"phase"`
	StartBlock       int64     `json:"start_block"`
	EndBlock         int64     `json:"end_block"`
	InitialReward    uint64    `json:"initial_reward"` // in wei (ZEN base unit)
	CurrentReward    uint64    `json:"current_reward"`
	TotalDistributed uint64    `json:"total_distributed"`
	RemainingPool    uint64    `json:"remaining_pool"`
	NextHalving      int64     `json:"next_halving"`
}

// AEHConfig holds Adaptive Exponential Halving configuration
type AEHConfig struct {
	TotalPool         uint64  `json:"total_pool"`         // 200M ZEN total pool
	InitialReward     uint64  `json:"initial_reward"`     // Initial reward per block
	HalvingFactor     float64 `json:"halving_factor"`     // 0.95 (5% reduction)
	HalvingInterval   int64   `json:"halving_interval"`   // ~3 months in blocks
	AdaptiveEnabled   bool    `json:"adaptive_enabled"`   // AI-based adjustment
	AdaptiveThreshold float64 `json:"adaptive_threshold"` // 50% TVL threshold
}

// RewardRecord tracks reward distribution
type RewardRecord struct {
	BlockNumber   int64     `json:"block_number"`
	Validator     []byte    `json:"validator"`
	Amount        uint64    `json:"amount"`
	Phase         int       `json:"phase"`
	Timestamp     int64     `json:"timestamp"`
}

// Halving handles Adaptive Exponential Halving (AEH)
type Halving struct {
	mu             sync.RWMutex
	config         AEHConfig
	phases         []HalvingPhase
	currentPhase   int
	currentBlock   int64
	rewardPool     uint64
	distributed    uint64
	rewardHistory  []RewardRecord
	aiAdapter      *AIAdapter
	adaptiveActive bool
}

// AIAdapter handles AI-based adaptive adjustments
type AIAdapter struct {
	mu            sync.RWMutex
	tvlPercent    float64
	validatorCount int
	networkTVL    uint64
	adjustmentFactor float64
	learningRate  float64
}

// New creates a new halving instance
func New() *Halving {
	return &Halving{
		config: AEHConfig{
			TotalPool:         200000000000000000000000000, // 200M ZEN
			InitialReward:     1000000000000000000000,       // 1000 ZEN per block
			HalvingFactor:     0.95,                         // 5% reduction
			HalvingInterval:   7889400,                      // ~3 months
			AdaptiveEnabled:   true,
			AdaptiveThreshold: 0.50,                         // 50% TVL
		},
		phases:        make([]HalvingPhase, 0),
		rewardPool:    200000000000000000000000000,
		distributed:   0,
		rewardHistory: make([]RewardRecord, 0),
		aiAdapter:     &AIAdapter{adjustmentFactor: 1.0, learningRate: 0.1},
	}
}

// NewWithConfig creates halving with custom configuration
func NewWithConfig(config AEHConfig) *Halving {
	return &Halving{
		config:        config,
		phases:        make([]HalvingPhase, 0),
		rewardPool:    config.TotalPool,
		distributed:   0,
		rewardHistory: make([]RewardRecord, 0),
		aiAdapter:     &AIAdapter{adjustmentFactor: 1.0, learningRate: 0.1},
	}
}

// Start initializes the halving engine
func (h *Halving) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Println("[HALVING] Initializing Adaptive Exponential Halving (AEH)")
	fmt.Printf("  - Total Pool: %d ZEN (%.0fM)\n",
		h.config.TotalPool/1e18, float64(h.config.TotalPool)/1e18)
	fmt.Printf("  - Initial Reward: %d ZEN\n", h.config.InitialReward/1e18)
	fmt.Printf("  - Halving Factor: %.2f (%.1f%% reduction)\n",
		h.config.HalvingFactor, (1-h.config.HalvingFactor)*100)
	fmt.Printf("  - Halving Interval: %d blocks (~3 months)\n", h.config.HalvingInterval)
	fmt.Printf("  - Adaptive: %v\n", h.config.AdaptiveEnabled)

	// Initialize phase 0
	phase0 := HalvingPhase{
		Phase:            0,
		StartBlock:       0,
		EndBlock:         h.config.HalvingInterval - 1,
		InitialReward:    h.config.InitialReward,
		CurrentReward:    h.config.InitialReward,
		TotalDistributed: 0,
		RemainingPool:    h.config.TotalPool,
		NextHalving:      h.config.HalvingInterval,
	}
	h.phases = append(h.phases, phase0)
	h.currentPhase = 0
	h.currentBlock = 0

	fmt.Println("✓ Halving engine initialized")
	fmt.Println("  Phase 0 active (Block 0 - " + fmt.Sprintf("%d", h.config.HalvingInterval-1) + ")")
	fmt.Printf("  Next halving at block %d\n", phase0.NextHalving)

	return nil
}

// Stop halts the halving engine
func (h *Halving) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Println("[HALVING] Stopping halving engine")
	return nil
}

// CalculateReward calculates reward for a validator at given block
func (h *Halving) CalculateReward(blockNumber int64, validator []byte) (uint64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if halving should occur
	if h.shouldHalve(blockNumber) {
		if err := h.performHalving(blockNumber); err != nil {
			return 0, fmt.Errorf("halving failed: %w", err)
		}
	}

	// Calculate base reward
	phase := h.phases[h.currentPhase]
	reward := phase.CurrentReward

	// Apply AI-based adaptive adjustment if enabled
	if h.config.AdaptiveEnabled {
		adjustment := h.aiAdapter.calculateAdjustment()
		reward = uint64(float64(reward) * adjustment)
	}

	// Check if we have enough in pool
	if reward > h.rewardPool {
		reward = h.rewardPool
		if reward == 0 {
			return 0, fmt.Errorf("reward pool exhausted")
		}
	}

	// Update pool and distributed
	h.rewardPool -= reward
	h.distributed += reward

	// Update phase
	h.phases[h.currentPhase].TotalDistributed += reward
	h.phases[h.currentPhase].RemainingPool -= reward

	// Record reward
	record := RewardRecord{
		BlockNumber: blockNumber,
		Validator:   validator,
		Amount:      reward,
		Phase:       h.currentPhase,
		Timestamp:   time.Now().Unix(),
	}
	h.rewardHistory = append(h.rewardHistory, record)

	// Keep history manageable
	if len(h.rewardHistory) > 10000 {
		h.rewardHistory = h.rewardHistory[1:]
	}

	return reward, nil
}

// shouldHalve checks if a halving should occur
func (h *Halving) shouldHalve(blockNumber int64) bool {
	// Check if we've reached halving interval
	if h.currentBlock >= int64(len(h.phases))*h.config.HalvingInterval {
		return true
	}

	// In phase > 0, check specific intervals
	if h.currentPhase > 0 {
		blocksInPhase := h.currentBlock - h.phases[h.currentPhase].StartBlock
		if blocksInPhase >= h.config.HalvingInterval {
			return true
		}
	}

	return false
}

// performHalving executes a halving event
func (h *Halving) performHalving(blockNumber int64) error {
	currentPhase := h.phases[h.currentPhase]

	// Calculate new reward using exponential decay
	// reward_n = initial_reward * (halving_factor)^n
	n := float64(h.currentPhase)
	newReward := h.config.InitialReward * math.Pow(h.config.HalvingFactor, n)

	// Create new phase
	newPhase := HalvingPhase{
		Phase:            h.currentPhase + 1,
		StartBlock:       blockNumber,
		EndBlock:         blockNumber + h.config.HalvingInterval - 1,
		InitialReward:    h.config.InitialReward,
		CurrentReward:    uint64(newReward),
		TotalDistributed: 0,
		RemainingPool:    h.rewardPool,
		NextHalving:      blockNumber + h.config.HalvingInterval,
	}

	h.phases = append(h.phases, newPhase)
	h.currentPhase++

	fmt.Printf("[HALVING] Halving event at block %d!\n", blockNumber)
	fmt.Printf("  Phase: %d → %d\n", currentPhase.Phase, newPhase.Phase)
	fmt.Printf("  Reward: %d → %d ZEN (%.1f%% reduction)\n",
		currentPhase.CurrentReward/1e18,
		newPhase.CurrentReward/1e18,
		(1-h.config.HalvingFactor)*100)
	fmt.Printf("  Remaining pool: %d ZEN\n", h.rewardPool/1e18)

	return nil
}

// GetCurrentPhase returns current halving phase
func (h *Halving) GetCurrentPhase() HalvingPhase {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.phases[h.currentPhase]
}

// GetAllPhases returns all halving phases
func (h *Halving) GetAllPhases() []HalvingPhase {
	h.mu.RLock()
	defer h.mu.RUnlock()
	phases := make([]HalvingPhase, len(h.phases))
	copy(phases, h.phases)
	return phases
}

// GetRewardPoolStatus returns current pool status
func (h *Halving) GetRewardPoolStatus() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_pool":       h.config.TotalPool / 1e18,
		"reward_pool":      h.rewardPool / 1e18,
		"distributed":      h.distributed / 1e18,
		"distributed_pct":  float64(h.distributed) / float64(h.config.TotalPool) * 100,
		"remaining_pct":    float64(h.rewardPool) / float64(h.config.TotalPool) * 100,
		"current_phase":    h.currentPhase,
		"phases_remaining": h.estimatePhasesRemaining(),
	}
}

// estimatePhasesRemaining estimates remaining phases
func (h *Halving) estimatePhasesRemaining() int {
	// Each halving reduces reward by 5%
	// Pool depletes when: sum(reward * interval) >= total_pool
	// Approximate calculation

	avgReward := float64(h.phases[h.currentPhase].CurrentReward)
	rewardPerPeriod := avgReward * float64(h.config.HalvingInterval)
	periodsRemaining := float64(h.rewardPool) / rewardPerPeriod

	return int(periodsRemaining)
}

// UpdateTVL updates total value locked (for adaptive mode)
func (h *Halving) UpdateTVL(tvl uint64, validatorCount int) {
	h.aiAdapter.mu.Lock()
	defer h.aiAdapter.mu.Unlock()

	h.aiAdapter.networkTVL = tvl
	h.aiAdapter.validatorCount = validatorCount

	// Calculate TVL percentage of total supply
	// Assuming total supply is 1B ZEN
	totalSupply := uint64(1000000000000000000000000000) // 1B ZEN
	h.aiAdapter.tvlPercent = float64(tvl) / float64(totalSupply)

	// Adjust based on TVL
	if h.config.AdaptiveEnabled {
		if h.aiAdapter.tvlPercent < h.config.AdaptiveThreshold {
			// Low TVL: increase rewards slightly to incentivize staking
			h.aiAdapter.adjustmentFactor = 1.0 + (h.config.AdaptiveThreshold - h.aiAdapter.tvlPercent) * 0.5
		} else {
			// High TVL: decrease rewards (sustainable)
			h.aiAdapter.adjustmentFactor = 1.0 - (h.aiAdapter.tvlPercent - h.config.AdaptiveThreshold) * 0.3
		}

		// Clamp adjustment
		if h.aiAdapter.adjustmentFactor > 1.5 {
			h.aiAdapter.adjustmentFactor = 1.5
		}
		if h.aiAdapter.adjustmentFactor < 0.5 {
			h.aiAdapter.adjustmentFactor = 0.5
		}
	}
}

// calculateAdjustment calculates adaptive adjustment factor
func (a *AIAdapter) calculateAdjustment() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.adjustmentFactor
}

// PredictExhaustion predicts when reward pool will be exhausted
func (h *Halving) PredictExhaustion() (int64, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.rewardPool == 0 {
		return h.currentBlock, nil
	}

	// Simple linear projection
	// In production: use AI model for better prediction
	currentReward := h.phases[h.currentPhase].CurrentReward
	blocksRemaining := h.rewardPool / currentReward
	estimatedExhaustion := h.currentBlock + int64(blocksRemaining)

	return estimatedExhaustion, nil
}

// GetRewardHistory returns recent reward history
func (h *Halving) GetRewardHistory(limit int) []RewardRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if limit <= 0 || limit > len(h.rewardHistory) {
		limit = len(h.rewardHistory)
	}

	history := make([]RewardRecord, limit)
	start := len(h.rewardHistory) - limit
	copy(history, h.rewardHistory[start:])

	return history
}

// GetStats returns halving statistics
func (h *Halving) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	exhaustionBlock, _ := h.PredictExhaustion()
	poolStatus := h.GetRewardPoolStatus()

	return map[string]interface{}{
		"current_phase":     h.currentPhase,
		"current_block":     h.currentBlock,
		"next_halving":      h.phases[h.currentPhase].NextHalving,
		"current_reward":    h.phases[h.currentPhase].CurrentReward / 1e18,
		"halving_factor":    h.config.HalvingFactor,
		"adaptive_enabled":  h.config.AdaptiveEnabled,
		"adaptive_factor":   h.aiAdapter.adjustmentFactor,
		"reward_pool":       poolStatus["reward_pool"],
		"distributed_pct":   poolStatus["distributed_pct"],
		"phases_remaining":  h.estimatePhasesRemaining(),
		"predicted_exhaustion": exhaustionBlock,
		"total_phases":      len(h.phases),
		"history_entries":   len(h.rewardHistory),
	}
}

// SetBlockNumber updates current block number
func (h *Halving) SetBlockNumber(blockNumber int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentBlock = blockNumber
}

// IsExhausted checks if reward pool is exhausted
func (h *Halving) IsExhausted() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rewardPool == 0
}
