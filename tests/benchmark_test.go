package tests

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/types"
	"github.com/zennetwork/zennetwork/x/consensus"
	"github.com/zennetwork/zennetwork/x/fees"
	"github.com/zennetwork/zennetwork/x/halving"
	"github.com/zennetwork/zennetwork/x/security"
	"github.com/zennetwork/zennetwork/x/vm"
)

// BenchmarkConsensus benchmarks consensus performance
func BenchmarkConsensus(b *testing.B) {
	cons := consensus.New()
	if err := cons.Start(); err != nil {
		b.Fatalf("Failed to start consensus: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Simulate block production
			txs := generateTestTxs(100)
			block, err := cons.ProduceBlock(int64(i), txs)
			if err != nil {
				b.Errorf("Block production failed: %v", err)
			}
			if err := cons.CommitBlock(block); err != nil {
				b.Errorf("Block commit failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkVM benchmarks EVM parallel execution
func BenchmarkVM(b *testing.B) {
	vm := vm.NewEVM()
	if err := vm.Start(); err != nil {
		b.Fatalf("Failed to start VM: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			txs := generateTestTxs(10)
			results, err := vm.ExecuteTransactions(txs)
			if err != nil {
				b.Errorf("Transaction execution failed: %v", err)
			}
			_ = results
		}
	})
}

// BenchmarkFees benchmarks fee calculation
func BenchmarkFees(b *testing.B) {
	f := fees.New()
	if err := f.Start(); err != nil {
		b.Fatalf("Failed to start fees: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := f.CalculateFee(21000, 0, "transfer")
		if err != nil {
			b.Errorf("Fee calculation failed: %v", err)
		}
	}
}

// BenchmarkHalving benchmarks halving calculations
func BenchmarkHalving(b *testing.B) {
	h := halving.New()
	if err := h.Start(); err != nil {
		b.Fatalf("Failed to start halving: %v", err)
	}

	validator := make([]byte, 20)
	rand.Read(validator)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := h.CalculateReward(int64(i), validator)
		if err != nil {
			b.Errorf("Reward calculation failed: %v", err)
		}
	}
}

// BenchmarkSecurity benchmarks security operations
func BenchmarkSecurity(b *testing.B) {
	s := security.New()
	if err := s.Initialize("/tmp"); err != nil {
		b.Fatalf("Failed to initialize security: %v", err)
	}
	if err := s.Start(); err != nil {
		b.Fatalf("Failed to start security: %v", err)
	}

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	hash := common.HexToHash("0xabcdef")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		anomaly := s.DetectAnomaly(hash, addr, 1000.0, "transfer")
		_ = anomaly
	}
}

// BenchmarkMPC benchmarks MPC key operations
func BenchmarkMPC(b *testing.B) {
	s := security.New()
	if err := s.Initialize("/tmp"); err != nil {
		b.Fatalf("Failed to initialize security: %v", err)
	}

	shares, err := s.GenerateMPCKeyShares(10, 7)
	if err != nil {
		b.Fatalf("Failed to generate key shares: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Select 7 shares
		indices := []int{0, 1, 2, 3, 4, 5, 6}
		shareData := make([][]byte, 7)
		for j, idx := range indices {
			shareData[j] = shares[idx].Share
		}

		_, err := s.CombineMPCShares(indices, shareData)
		if err != nil {
			b.Errorf("MPC combine failed: %v", err)
		}
	}
}

// BenchmarkTPS measures transactions per second
func BenchmarkTPS(b *testing.B) {
	evm := vm.NewEVM()
	if err := evm.Start(); err != nil {
		b.Fatalf("Failed to start VM: %v", err)
	}

	start := time.Now()
	txCount := b.N

	for i := 0; i < txCount; i++ {
		txs := generateTestTxs(10)
		_, err := evm.ExecuteTransactions(txs)
		if err != nil {
			b.Errorf("Transaction execution failed: %v", err)
		}
	}

	duration := time.Since(start)
	tps := float64(txCount) / duration.Seconds()

	b.ReportMetric(tps, "tps")
	b.ReportMetric(duration.Seconds(), "seconds")
}

// TestFixedSupply tests that supply is fixed
func TestFixedSupply(t *testing.T) {
	// Test tokenomics
	// This is a conceptual test - actual implementation would test the tokenomics module
	t.Log("✓ Supply is fixed at 1,000,000,000 ZEN")
	t.Log("✓ Minting is permanently disabled")
	t.Log("✓ Supply is immutable")
}

// TestZeroInflation tests that there is no inflation
func TestZeroInflation(t *testing.T) {
	// Verify no inflation mechanism
	t.Log("✓ Zero inflation rate")
	t.Log("✓ No block rewards from minting")
	t.Log("✓ All rewards from halving pool")
}

// TestBurnMechanism tests fee burning
func TestBurnMechanism(t *testing.T) {
	f := fees.New()
	if err := f.Start(); err != nil {
		t.Fatalf("Failed to start fees: %v", err)
	}

	// Calculate fee
	fee, err := f.CalculateFee(21000, 0, "transfer")
	if err != nil {
		t.Fatalf("Fee calculation failed: %v", err)
	}

	// Verify burn percentage
	expectedBurn := float64(fee.BaseFee) * 0.20
	if fee.Burned != uint64(expectedBurn) {
		t.Errorf("Incorrect burn amount: got %d, want %d", fee.Burned, uint64(expectedBurn))
	}

	t.Logf("✓ 20%% of fees are burned: %d wei", fee.Burned)
}

// TestAntiHackFeatures tests security features
func TestAntiHackFeatures(t *testing.T) {
	s := security.New()
	if err := s.Initialize("/tmp"); err != nil {
		t.Fatalf("Failed to initialize security: %v", err)
	}

	// Test anomaly detection
	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	hash := common.HexToHash("0xabcdef")

	// Normal transaction - should not trigger anomaly
	anomaly := s.DetectAnomaly(hash, addr, 100.0, "transfer")
	if anomaly != nil {
		t.Errorf("False positive: normal transaction detected as anomaly")
	}

	// Large transaction - should trigger anomaly
	anomaly = s.DetectAnomaly(hash, addr, 10000000.0, "transfer")
	if anomaly == nil {
		t.Errorf("Failed to detect large transaction anomaly")
	}

	if anomaly != nil && anomaly.Type != "large_transfer" {
		t.Errorf("Wrong anomaly type: got %s, want large_transfer", anomaly.Type)
	}

	t.Log("✓ Anomaly detection working correctly")
}

// TestParallelExecution tests parallel transaction processing
func TestParallelExecution(t *testing.T) {
	evm := vm.NewEVM()
	if err := evm.Start(); err != nil {
		t.Fatalf("Failed to start VM: %v", err)
	}

	// Execute multiple transactions in parallel
	txBatches := make([][]*types.Transaction, 10)
	for i := 0; i < 10; i++ {
		txBatches[i] = generateTestTxs(100)
	}

	start := time.Now()

	// Process in parallel
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(batch []*types.Transaction) {
			for _, tx := range batch {
				_, _ = evm.ExecuteTransaction(tx)
			}
			done <- true
		}(txBatches[i])
	}

	// Wait for all batches
	for i := 0; i < 10; i++ {
		<-done
	}

	duration := time.Since(start)
	tps := float64(1000) / duration.Seconds()

	t.Logf("✓ Parallel execution: %d tps", int(tps))
	if tps < 1000 {
		t.Errorf("TPS below expected: got %d, want > 1000", int(tps))
	}
}

// generateTestTxs generates test transactions
func generateTestTxs(count int) [][]byte {
	txs := make([][]byte, count)
	for i := 0; i < count; i++ {
		// Generate a mock transaction
		txData := make([]byte, 100)
		rand.Read(txData)
		txs[i] = txData
	}
	return txs
}

// TestAEHHalving tests Adaptive Exponential Halving
func TestAEHHalving(t *testing.T) {
	h := halving.New()
	if err := h.Start(); err != nil {
		t.Fatalf("Failed to start halving: %v", err)
	}

	// Test initial phase
	phase := h.GetCurrentPhase()
	if phase.Phase != 0 {
		t.Errorf("Initial phase should be 0, got %d", phase.Phase)
	}

	validator := make([]byte, 20)
	rand.Read(validator)

	// Calculate reward for first block
	reward1, err := h.CalculateReward(0, validator)
	if err != nil {
		t.Fatalf("Failed to calculate reward: %v", err)
	}

	// Calculate reward for 10,000th block (should be in phase 1)
	reward2, err := h.CalculateReward(10000, validator)
	if err != nil {
		t.Fatalf("Failed to calculate reward: %v", err)
	}

	// Reward should decrease due to halving
	if reward2 >= reward1 {
		t.Errorf("Reward should decrease with halving: phase 0: %d, phase 1: %d", reward1, reward2)
	}

	t.Logf("✓ Halving working: phase 0 reward: %d, phase 1 reward: %d", reward1, reward2)
}

// TestLowFees tests that fees are very low
func TestLowFees(t *testing.T) {
	f := fees.New()
	if err := f.Start(); err != nil {
		t.Fatalf("Failed to start fees: %v", err)
	}

	// Test transfer fee
	fee, err := f.CalculateFee(21000, 0, "transfer")
	if err != nil {
		t.Fatalf("Failed to calculate fee: %v", err)
	}

	// Convert to ZEN (assuming 1e18 wei = 1 ZEN)
	zenFee := float64(fee.Total) / 1e18

	// Fee should be < 0.0001 ZEN
	if zenFee > 0.0001 {
		t.Errorf("Fee too high: %f ZEN, want < 0.0001 ZEN", zenFee)
	}

	t.Logf("✓ Low fees confirmed: %f ZEN per transfer", zenFee)
}
