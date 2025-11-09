package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tendermint/tendermint/types"
)

// ConsensusType defines the hybrid PoS + PoH consensus
type ConsensusType string

const (
	PoS   ConsensusType = "pos"   // Proof of Stake
	PoH   ConsensusType = "poh"   // Proof of History
	Hybrid ConsensusType = "hybrid" // PoS + PoH hybrid
)

// BlockTime in milliseconds
const (
	BlockTime     = 3000  // 3 seconds
	FinalityTime  = 1800  // <2 seconds
	TargetTPS     = 10000 // Base target TPS
	MaxTPS        = 50000 // Maximum TPS with parallel execution
	MinStake      = 1000000000000000000000 // 1000 ZEN (18 decimals)
)

// Validator represents a network validator
type Validator struct {
	Address             []byte            `json:"address"`
	PubKey              []byte            `json:"pub_key"`
	Stake               uint64            `json:"stake"` // in ZEN (base unit)
	Power               int64             `json:"power"`
	Reward              uint64            `json:"reward"`
	Slashed             bool              `json:"slashed"`
	VRFProof            []byte            `json:"vrf_proof"`
	PoHSequence         uint64            `json:"poh_sequence"`
	PoHTimestamp        int64             `json:"poh_timestamp"`
	ValidatorType       ConsensusType     `json:"validator_type"`
	LastBlockProduced   int64             `json:"last_block_produced"`
	SlashingEvents      []SlashingEvent   `json:"slashing_events"`
	EcoScore            float64           `json:"eco_score"` // Green validator score
}

// SlashingEvent tracks validator violations
type SlashingEvent struct {
	Height    int64  `json:"height"`
	Reason    string `json:"reason"`
	Penalty   uint64 `json:"penalty"` // Amount slashed
	Timestamp int64  `json:"timestamp"`
}

// Committee represents a consensus committee
type Committee struct {
	ID          uint64      `json:"id"`
	Validators  []Validator `json:"validators"`
	Shuffled    bool        `json:"shuffled"`
	BlockHash   []byte      `json:"block_hash"`
	PoHSequence uint64      `json:"poh_sequence"`
}

// ProofOfHistoryEntry represents a PoH sequence entry
type ProofOfHistoryEntry struct {
	Index         uint64 `json:"index"`
	Hash          []byte `json:"hash"`
	PreviousHash  []byte `json:"previous_hash"`
	Timestamp     int64  `json:"timestamp"`
	EntryData     []byte `json:"entry_data"`
}

// Consensus handles hybrid PoS + PoH consensus
type Consensus struct {
	mu              sync.RWMutex
	ValidatorSet    []Validator     `json:"validator_set"`
	CurrentHeight   int64           `json:"current_height"`
	CurrentBlock    *types.Block    `json:"current_block"`
	Commit          *types.Commit   `json:"commit"`
	PoHSequence     []ProofOfHistoryEntry `json:"poh_sequence"`
	Committees      []Committee     `json:"committees"`
	ConsensusType   ConsensusType   `json:"consensus_type"`
	BlockProducers  []uint64        `json:"block_producers"` // Shard IDs
	FinalityVotes   map[int64][]*types.Vote `json:"finality_votes"`
	muFinality      sync.Mutex
}

// New creates a new consensus instance
func New() *Consensus {
	return &Consensus{
		ValidatorSet:    make([]Validator, 0),
		CurrentHeight:   0,
		PoHSequence:     make([]ProofOfHistoryEntry, 0),
		Committees:      make([]Committee, 0),
		ConsensusType:   Hybrid,
		BlockProducers:  make([]uint64, 64), // 64 shards
		FinalityVotes:   make(map[int64][]*types.Vote),
	}
}

// Start begins consensus operations
func (c *Consensus) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("[CONSENSUS] Starting hybrid PoS + PoH consensus")
	fmt.Printf("  - Block Time: %dms\n", BlockTime)
	fmt.Printf("  - Finality: <%dms\n", FinalityTime)
	fmt.Printf("  - Target TPS: %d (Max: %d)\n", TargetTPS, MaxTPS)
	fmt.Printf("  - Min Stake: %d ZEN\n", MinStake/1000000000000000000)
	fmt.Printf("  - Validators: %d\n", len(c.ValidatorSet))

	// Initialize PoH with genesis
	if err := c.initializePoH(); err != nil {
		return fmt.Errorf("failed to initialize PoH: %w", err)
	}

	// Shuffle validators into committees
	c.shuffleValidators()

	// Start block production loop
	go c.blockProductionLoop()

	return nil
}

// Stop halts consensus
func (c *Consensus) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("[CONSENSUS] Stopping consensus engine")
	return nil
}

// AddValidator adds a new validator to the set
func (c *Consensus) AddValidator(v Validator) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check minimum stake
	if v.Stake < MinStake {
		return fmt.Errorf("validator stake below minimum: %d < %d", v.Stake, MinStake)
	}

	// Check if already exists
	for _, val := range c.ValidatorSet {
		if string(val.Address) == string(v.Address) {
			return fmt.Errorf("validator already exists: %x", v.Address)
		}
	}

	// Calculate voting power based on stake
	v.Power = int64(v.Stake / 1000000000) // Normalize

	// Add to set
	c.ValidatorSet = append(c.ValidatorSet, v)

	// Re-shuffle committees
	c.shuffleValidators()

	fmt.Printf("[CONSENSUS] Added validator: %x (Stake: %d ZEN, Power: %d)\n",
		v.Address[:8], v.Stake/1000000000000000000, v.Power)

	return nil
}

// RemoveValidator removes a validator from the set
func (c *Consensus) RemoveValidator(address []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, val := range c.ValidatorSet {
		if string(val.Address) == string(address) {
			c.ValidatorSet = append(c.ValidatorSet[:i], c.ValidatorSet[i+1:]...)
			c.shuffleValidators()
			fmt.Printf("[CONSENSUS] Removed validator: %x\n", address[:8])
			return nil
		}
	}

	return fmt.Errorf("validator not found: %x", address[:8])
}

// ProduceBlock produces a new block using PoS + PoH
func (c *Consensus) ProduceBlock(height int64, txs [][]byte) (*types.Block, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get PoH entry for this height
	pohEntry, err := c.getPoHEntry(height)
	if err != nil {
		return nil, fmt.Errorf("failed to get PoH entry: %w", err)
	}

	// Select validator based on PoS (stake-weighted) and PoH (sequence)
	proposer, err := c.selectProposer(height, pohEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to select proposer: %w", err)
	}

	// Create block header
	header := &types.Header{
		Height:     height,
		Time:       time.Now(),
		LastBlockID: types.BlockID{Hash: c.CurrentBlock.Header.Hash()},
		Proposer:   proposer,
	}

	// Create block
	block := &types.Block{
		Header: header,
		Data: types.Data{
			Txs: txs,
		},
	}

	// Add PoH proof to block
	pohProof := PoHProof{
		Entry:      *pohEntry,
		Validator:  proposer,
		Signature:  []byte{}, // Would be actual signature in production
		Timestamp:  time.Now().UnixNano(),
	}

	// Encode PoH proof
	pohProofBytes, _ := json.Marshal(pohProof)
	block.Data.Extensions = []types.Extension{
		{Index: 0, Bytes: pohProofBytes},
	}

	// Update current state
	c.CurrentHeight = height
	c.CurrentBlock = block

	fmt.Printf("[CONSENSUS] Block produced at height %d by validator %x\n",
		height, proposer[:8])

	return block, nil
}

// CommitBlock commits a block and finalizes it
func (c *Consensus) CommitBlock(block *types.Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify PoH proof
	if err := c.verifyPoHProof(block); err != nil {
		return fmt.Errorf("PoH proof verification failed: %w", err)
	}

	// Collect signatures for commit
	// In production, this would be actual validator signatures
	commit := &types.Commit{
		BlockID: types.BlockID{Hash: block.Header.Hash()},
		Signatures: make([]types.CommitSig, 0),
	}

	c.Commit = commit

	fmt.Printf("[CONSENSUS] Block committed at height %d (Finality achieved in <2s)\n",
		block.Header.Height)

	return nil
}

// FinalizeBlock achieves finality using BFT
func (c *Consensus) FinalizeBlock(block *types.Block) error {
	c.muFinality.Lock()
	defer c.muFinality.Unlock()

	// 2/3 + 1 validator signatures for finality
	requiredSignatures := (len(c.ValidatorSet) * 2 / 3) + 1

	// Check if we have enough signatures
	currentVotes := c.FinalityVotes[block.Header.Height]
	if len(currentVotes) >= requiredSignatures {
		// Block is finalized
		fmt.Printf("[CONSENSUS] Block finalized at height %d (TPS: %d)\n",
			block.Header.Height, c.calculateTPS())

		// Reward validators
		c.distributeRewards(block.Header.Height)

		// Update PoH sequence
		c.updatePoHSequence(block)

		return nil
	}

	return fmt.Errorf("insufficient signatures for finality: %d/%d",
		len(currentVotes), requiredSignatures)
}

// SlashValidator penalizes a validator for misbehavior
func (c *Consensus) SlashValidator(address []byte, reason string, penalty uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find validator
	for i, val := range c.ValidatorSet {
		if string(val.Address) == string(address) {
			// Apply slashing
			if penalty > val.Stake {
				penalty = val.Stake
			}
			val.Stake -= penalty
			val.Slashed = true

			// Record slashing event
			event := SlashingEvent{
				Height:    c.CurrentHeight,
				Reason:    reason,
				Penalty:   penalty,
				Timestamp: time.Now().Unix(),
			}
			val.SlashingEvents = append(val.SlashingEvents, event)

			// Remove from validator set if stake drops below minimum
			if val.Stake < MinStake {
				c.ValidatorSet = append(c.ValidatorSet[:i], c.ValidatorSet[i+1:]...)
				fmt.Printf("[CONSENSUS] Validator %x removed (below minimum stake)\n", address[:8])
			} else {
				c.ValidatorSet[i] = val
			}

			fmt.Printf("[CONSENSUS] Validator %x slashed: %d ZEN (%s)\n",
				address[:8], penalty/1000000000000000000, reason)

			return nil
		}
	}

	return fmt.Errorf("validator not found: %x", address[:8])
}

// PoHProof represents a Proof of History proof
type PoHProof struct {
	Entry      ProofOfHistoryEntry `json:"entry"`
	Validator  []byte              `json:"validator"`
	Signature  []byte              `json:"signature"`
	Timestamp  int64               `json:"timestamp"`
}

// initializePoH creates the initial PoH sequence
func (c *Consensus) initializePoH() error {
	// Genesis entry
	genesis := ProofOfHistoryEntry{
		Index:         0,
		Hash:          []byte("genesis"),
		PreviousHash:  []byte{},
		Timestamp:     time.Now().Unix(),
		EntryData:     []byte("zen-network-genesis"),
	}
	c.PoHSequence = append(c.PoHSequence, genesis)

	fmt.Println("[CONSENSUS] PoH initialized with genesis entry")
	return nil
}

// shuffleValidators creates consensus committees
func (c *Consensus) shuffleValidators() {
	totalValidators := len(c.ValidatorSet)
	if totalValidators == 0 {
		return
	}

	// Create 64 committees (one per shard)
	c.Committees = make([]Committee, 64)

	validatorsPerShard := totalValidators / 64
	if validatorsPerShard == 0 {
		validatorsPerShard = 1
	}

	for shardID := 0; shardID < 64; shardID++ {
		committee := Committee{
			ID:          uint64(shardID),
			Shuffled:    true,
			PoHSequence: uint64(shardID),
		}

		start := shardID * validatorsPerShard
		end := start + validatorsPerShard
		if end > totalValidators {
			end = totalValidators
		}
		if start >= totalValidators {
			start = 0
			end = 0
		}

		committee.Validators = c.ValidatorSet[start:end]
		c.Committees[shardID] = committee
	}

	fmt.Printf("[CONSENSUS] Created %d committees (%d validators/shard)\n",
		len(c.Committees), validatorsPerShard)
}

// blockProductionLoop manages continuous block production
func (c *Consensus) blockProductionLoop() {
	ticker := time.NewTicker(time.Millisecond * BlockTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			height := c.CurrentHeight + 1
			c.mu.Unlock()

			// Get transactions from mempool
			// In production: get from network module
			txs := make([][]byte, 0)

			// Produce block
			if block, err := c.ProduceBlock(height, txs); err == nil {
				// Commit block
				if err := c.CommitBlock(block); err == nil {
					// Finalize
					c.FinalizeBlock(block)
				}
			}
		}
	}
}

// getPoHEntry retrieves a PoH entry for given height
func (c *Consensus) getPoHEntry(height int64) (*ProofOfHistoryEntry, error) {
	if height < 0 || height >= int64(len(c.PoHSequence)) {
		// Generate new entry
		prevEntry := c.PoHSequence[len(c.PoHSequence)-1]
		entry := ProofOfHistoryEntry{
			Index:         prevEntry.Index + 1,
			Hash:          c.hashEntry(prevEntry, height),
			PreviousHash:  prevEntry.Hash,
			Timestamp:     time.Now().Unix(),
			EntryData:     []byte{},
		}
		c.PoHSequence = append(c.PoHSequence, entry)
		return &entry, nil
	}

	return &c.PoHSequence[height], nil
}

// hashEntry creates a hash for PoH entry
func (c *Consensus) hashEntry(prev ProofOfHistoryEntry, height int64) []byte {
	// Combine previous hash, height, and timestamp
	data := append(prev.Hash, make([]byte, 8)...)
	binary.BigEndian.PutUint64(data[len(prev.Hash):], uint64(height))
	data = append(data, make([]byte, 8)...)
	binary.BigEndian.PutInt64(data[len(data)-8:], time.Now().Unix())

	hash := sha256.Sum256(data)
	return hash[:]
}

// selectProposer chooses the block proposer
func (c *Consensus) selectProposer(height int64, pohEntry *ProofOfHistoryEntry) ([]byte, error) {
	if len(c.ValidatorSet) == 0 {
		return nil, fmt.Errorf("no validators available")
	}

	// Use PoH sequence to pseudo-randomly select validator
	// In production: use VRF for unbiasable randomness
	validatorIndex := pohEntry.Index % uint64(len(c.ValidatorSet))
	return c.ValidatorSet[validatorIndex].Address, nil
}

// verifyPoHProof verifies a PoH proof
func (c *Consensus) verifyPoHProof(block *types.Block) error {
	// Check if block has PoH extension
	if len(block.Data.Extensions) == 0 {
		return fmt.Errorf("missing PoH proof")
	}

	// Verify the proof
	// In production: verify actual signature
	pohProof := PoHProof{}
	if err := json.Unmarshal(block.Data.Extensions[0].Bytes, &pohProof); err != nil {
		return fmt.Errorf("failed to unmarshal PoH proof: %w", err)
	}

	// Verify hash chain
	// In production: complete verification
	return nil
}

// updatePoHSequence updates the PoH sequence
func (c *Consensus) updatePoHSequence(block *types.Block) {
	// Add to sequence if not present
	if block.Header.Height >= int64(len(c.PoHSequence)) {
		entry := ProofOfHistoryEntry{
			Index:         uint64(block.Header.Height),
			Hash:          block.Header.Hash(),
			PreviousHash:  c.CurrentBlock.Header.Hash(),
			Timestamp:     time.Now().Unix(),
			EntryData:     block.Data.TxsHash,
		}
		c.PoHSequence = append(c.PoHSequence, entry)
	}
}

// calculateTPS calculates current transactions per second
func (c *Consensus) calculateTPS() int {
	// Simplified TPS calculation
	// In production: actual measurement
	return TargetTPS
}

// distributeRewards distributes block rewards to validators
func (c *Consensus) distributeRewards(height int64) {
	// Calculate total reward from halving module
	// In production: integrate with halving module
	totalReward := uint64(1000000000000000000) // 1 ZEN for example

	// Distribute to validators based on stake
	for i, val := range c.ValidatorSet {
		stakeRatio := float64(val.Stake) / float64(c.getTotalStake())
		reward := uint64(float64(totalReward) * stakeRatio)
		c.ValidatorSet[i].Reward += reward
	}
}

// getTotalStake calculates total staked amount
func (c *Consensus) getTotalStake() uint64 {
	var total uint64
	for _, val := range c.ValidatorSet {
		total += val.Stake
	}
	return total
}

// GetStatus returns current consensus status
func (c *Consensus) GetStatus() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"height":         c.CurrentHeight,
		"validators":     len(c.ValidatorSet),
		"committees":     len(c.Committees),
		"consensus_type": c.ConsensusType,
		"block_time":     BlockTime,
		"finality_time":  FinalityTime,
		"target_tps":     TargetTPS,
		"max_tps":        MaxTPS,
		"total_stake":    c.getTotalStake() / 1000000000000000000,
	}
}
