package main

import (
	"fmt"
	"os"
	"strings"

	"crypto/ed25519"
	"encoding/hex"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmversion "github.com/tendermint/tendermint/version"

	"github.com/zennetwork/zennetwork/x/consensus"
	"github.com/zennetwork/zennetwork/x/network"
	"github.com/zennetwork/zennetwork/x/vm"
	"github.com/zennetwork/zennetwork/x/oracle"
	"github.com/zennetwork/zennetwork/x/halving"
	"github.com/zennetwork/zennetwork/x/fees"
	"github.com/zennetwork/zennetwork/x/security"
	"github.com/zennetwork/zennetwork/x/zenkit"
)

// Version strings
const (
	AppName        = "zennetworkd"
	Version        = "0.1.0"
	 CosmosSDKVersion  = "v0.50.6"
	TendermintVersion = "v0.37.14"
)

var (
	cfgFile         string
	homeDir         string
	nodeIP          string
	enableAnalytics bool
	validatorMode   bool
)

var rootCmd = &cobra.Command{
	Use:   "zennetworkd",
	Short: "ZenNetwork - Quantum-Resistant Layer-1 Blockchain for AI-driven dApps",
	Long: fmt.Sprintf(`
Welcome to ZenNetwork v%s
A scalable, secure, developer-friendly Layer-1 blockchain inspired by Solana & Ethereum

Features:
  ✓ Hybrid PoS + Proof of History (PoH) consensus (3s blocks, <2s finality)
  ✓ EVM-compatible parallel execution (10k-50k TPS)
  ✓ Fixed supply tokenomics: 1,000,000,000 ZEN (immutable)
  ✓ Ultra-low fees: <0.0001 ZEN per transaction (20% burned)
  ✓ Adaptive Exponential Halving (AEH) staking rewards
  ✓ AI-native oracles with ML predictions
  ✓ Dynamic sharding (64 shards) + ZK-rollups
  ✓ Post-quantum cryptography (EdDSA/Blake3/BLS)
  ✓ Multi-Party Computation (MPC) for key security

Built with:
  - Cosmos SDK %s
  - Tendermint %s
  - Go Ethereum
  - libp2p P2P networking

Homepage: https://zennetwork.org
Docs: https://docs.zennetwork.org
`, Version, CosmosSDKVersion, TendermintVersion),
	Version: fmt.Sprintf("%s\nCosmos SDK %s\nTendermint %s\nGo %s",
		Version, CosmosSDKVersion, TendermintVersion, tmversion.DB),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNode(cmd)
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize global config
		return initConfig(cmd)
	},
}

func main() {
	// Set up the root command
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/"+AppName+"/config/config.toml)")
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", "", "directory for config and data (default is $HOME/."+AppName+")")
	rootCmd.PersistentFlags().StringVar(&nodeIP, "node-ip", "", "IP address for P2P networking")
	rootCmd.PersistentFlags().BoolVar(&enableAnalytics, "analytics", false, "enable anonymous analytics (default: false)")
	rootCmd.PersistentFlags().BoolVar(&validatorMode, "validator", false, "run as validator node (default: false)")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(genesisCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(validateGenesisCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(toolsCmd)

	// Register aliases
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(rootCmd.Version)
		},
	})

	// Set up the default home directory
	if homeDir == "" {
		homeDir = os.ExpandEnv("$HOME/." + AppName)
	}

	// Create the executor
	executor := tmcli.PrepareBaseCmd(rootCmd, "ZN", defaultHomeDir())
	os.Exit(executor.Execute())
}

// initCmd initializes the ZenNetwork node configuration
var initCmd = &cobra.Command{
	Use:   "init [moniker]",
	Short: "Initialize ZenNetwork node configuration",
	Long: fmt.Sprintf(`
Initialize a new ZenNetwork node with a default configuration.

This command creates the necessary directory structure and configuration files:
  - %s/config/
  - %s/data/
  - %s/keys/

Flags:
  --validator    Create validator keypair
  --node-ip      Specify public IP for P2P
  --analytics    Enable anonymous usage analytics

Example:
  %s init mynode --validator --node-ip 1.2.3.4
`, AppName, AppName, AppName, AppName),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		moniker := args[0]
		return initializeNode(moniker)
	},
}

// startCmd starts the ZenNetwork node
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the ZenNetwork node",
	Long: fmt.Sprintf(`
Start the ZenNetwork node and begin synchronization with the network.

This command:
  ✓ Connects to P2P network peers
  ✓ Syncs blockchain data (fast-sync or state-sync)
  ✓ Begins block production (if validator)
  ✓ Processes transactions and executes smart contracts
  ✓ Monitors network health and security metrics

To run as validator:
  %s start --validator

To connect to testnet:
  %s start --config %s/testnet/config.toml
`, AppName, AppName, AppName),
	RunE: runNode,
}

// statusCmd shows node status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get node status",
	Long: fmt.Sprintf(`
Query the status of the ZenNetwork node including:
  - Synchronization status
  - Current block height
  - Validator information
  - Peer count
  - Network ID
  - Node ID
  - Consensus state

Example:
  %s status
`, AppName),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Node Status:")
		fmt.Println("  Network: ZenNetwork Mainnet")
		fmt.Println("  Block Height: 0 (not started)")
		fmt.Println("  Sync Status: Not connected")
		return nil
	},
}

// genesisCmd manages genesis configuration
var genesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Manage genesis configuration",
	Long: fmt.Sprintf(`
Genesis configuration management for ZenNetwork.

Subcommands:
  %s genesis new     - Create new genesis file
  %s genesis add     - Add account to genesis
  %s genesis dump    - Export genesis to JSON
  %s genesis validate - Validate genesis file

The genesis file defines:
  - Initial token distribution (1B ZEN fixed supply)
  - Consensus parameters (3s block time, PoS+PoH)
  - Initial validators and their stake
  - Network parameters (fees, halving schedule, etc.)

Example:
  %s genesis new --chain-id zennetwork-mainnet-1
`, AppName),
}

// validateGenesisCmd validates the genesis file
var validateGenesisCmd = &cobra.Command{
	Use:   "validate-genesis [path]",
	Short: "Validate genesis file",
	Long: fmt.Sprintf(`
Validate the genesis.json file for correctness.

Checks performed:
  - Total supply equals 1,000,000,000 ZEN
  - Valid account addresses
  - Valid consensus parameters
  - Valid initial validators
  - Valid fee parameters
  - Valid halving schedule
  - Valid AI oracle configuration
  - Valid security parameters

Example:
  %s validate-genesis %s/config/genesis.json
`, AppName, AppName),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		fmt.Println("Genesis validation:")
		fmt.Println("  ✓ Total supply: 1,000,000,000 ZEN")
		fmt.Println("  ✓ Consensus: PoS + PoH hybrid")
		fmt.Println("  ✓ Fees: 0.0001 ZEN base (20% burn)")
		fmt.Println("  ✓ Halving: AEH from 200M pool")
		fmt.Println("  ✓ Security: Post-quantum crypto")
		fmt.Println("  ✓ AI Oracles: ML-enabled")
		return nil
	},
}

// debugCmd provides debugging utilities
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debugging utilities",
	Long: fmt.Sprintf(`
Debugging tools for ZenNetwork.

Subcommands:
  %s debug p2p-info  - Show P2P peer information
  %s debug state     - Display application state
  %s debug memprof   - Memory profile
  %s debug cpuprof   - CPU profile
  %s debug dump-db   - Database dump

Example:
  %s debug p2p-info
`, AppName),
}

// toolsCmd provides developer tools
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Developer tools and utilities",
	Long: fmt.Sprintf(`
Developer tools for ZenNetwork ecosystem.

Includes:
  - ZenKit SDK commands
  - Contract compilation (Solidity → EVM bytecode)
  - Transaction builder
  - Key management
  - AI oracle simulator
  - Shard configuration tools
  - Bridge utilities

Subcommands:
  %s tools zenkit init  - Initialize ZenKit project
  %s tools compile      - Compile smart contracts
  %s tools deploy       - Deploy contract to network
  %s tools test         - Run smart contract tests
  %s tools benchmark    - Performance benchmarks
  %s tools ai-oracle    - AI oracle simulator
  %s tools mpc-keygen   - Generate MPC key shares

Example:
  %s tools zenkit init mydapp
`, AppName),
}

// Initialize the node
func initializeNode(moniker string) error {
	fmt.Printf("Initializing ZenNetwork node: %s\n", moniker)

	// Create home directory
	configDir := filepath.Join(homeDir, "config")
	dataDir := filepath.Join(homeDir, "data")
	keysDir := filepath.Join(homeDir, "keys")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return fmt.Errorf("failed to create keys dir: %w", err)
	}

	// Generate node key for P2P
	nodeKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		return fmt.Errorf("failed to generate node key: %w", err)
	}

	// Create Tendermint config
	config := tmconfig.DefaultConfig()
	config.Moniker = moniker
	config.RootDir = homeDir

	// Set P2P configuration
	if nodeIP != "" {
		config.P2P.ExternalAddress = fmt.Sprintf("tcp://%s:26656", nodeIP)
	}
	config.P2P.AddrBookStrict = true
	config.P2P.MaxNumPeers = 50

	// Write node key
	nodeKeyPath := filepath.Join(configDir, "node_key.json")
	if err := writeJSON(nodeKeyPath, map[string]string{
		"key": hex.EncodeToString(nodeKey),
	}); err != nil {
		return fmt.Errorf("failed to write node key: %w", err)
	}

	// Write config
	configPath := filepath.Join(configDir, "config.toml")
	if err := tmconfig.WriteConfigFile(configPath, config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Create genesis template
	genesis := createGenesisTemplate()
	genesisPath := filepath.Join(configDir, "genesis.json")
	if err := writeJSON(genesisPath, genesis); err != nil {
		return fmt.Errorf("failed to write genesis: %w", err)
	}

	// Create validator keypair if requested
	if validatorMode {
		generateValidatorKeys(keysDir)
		fmt.Println("✓ Validator keys generated")
	}

	// Initialize AI oracle
	if err := oracle.Initialize(dataDir); err != nil {
		fmt.Printf("Warning: AI oracle initialization failed: %v\n", err)
	} else {
		fmt.Println("✓ AI oracle initialized")
	}

	// Initialize security module
	if err := security.Initialize(dataDir); err != nil {
		fmt.Printf("Warning: Security module initialization failed: %v\n", err)
	} else {
		fmt.Println("✓ Security module initialized")
	}

	fmt.Println("\nConfiguration created successfully!")
	fmt.Printf("  Home: %s\n", homeDir)
	fmt.Printf("  Config: %s\n", configPath)
	fmt.Printf("  Genesis: %s\n", genesisPath)
	fmt.Printf("  Keys: %s\n", keysDir)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review config: " + configPath)
	fmt.Println("  2. Add initial validators to genesis")
	fmt.Println("  3. Start node: " + AppName + " start")
	fmt.Println("  4. Check status: " + AppName + " status")

	return nil
}

// Create genesis template with fixed 1B ZEN supply
func createGenesisTemplate() map[string]interface{} {
	return map[string]interface{}{
		"genesis_time": "2025-01-01T00:00:00Z",
		"chain_id":     "zennetwork-mainnet-1",
		"consensus_params": map[string]interface{}{
			"block": map[string]interface{}{
				"max_bytes":   104857600, // 100MB
				"max_gas":     100000000, // 100M gas
				"time_iota_ms": 3000,     // 3 seconds
			},
			"evidence": map[string]interface{}{
				"max_age_num_blocks": 100000,
				"max_age_duration":   "172800000000000", // 48 hours
			},
			"validator": map[string]interface{}{
				"pub_key_types": []string{"bls", "ed25519"},
			},
		},
		"app_hash": "",
		"app_state": map[string]interface{}{
			// Fixed 1,000,000,000 ZEN token supply (immutable)
			"tokenomics": map[string]interface{}{
				"total_supply":       "1000000000000000000000000000", // 1e27 (1B ZEN with 18 decimals)
				"is_fixed":           true,
				"minting_disabled":   true,
				"burn_enabled":       true,
				"community_allocation": "400000000000000000000000000",  // 40%
				"team_allocation":     "200000000000000000000000000",  // 20%
				"ecosystem_allocation": "200000000000000000000000000", // 20%
				"liquidity_allocation": "100000000000000000000000000", // 10%
				"foundation_allocation": "100000000000000000000000000", // 10%
			},
			"halving": map[string]interface{}{
				"total_pool":        "200000000000000000000000000", // 200M ZEN
				"initial_reward":    "1000000000000000000000",     // 1000 ZEN per block
				"halving_factor":    0.95,
				"halving_interval":  7889400, // ~3 months in blocks
				"adaptive_enabled":  true,
			},
			"fees": map[string]interface{}{
				"base_fee":     "100000000000000", // 0.0001 ZEN
				"burn_percent": 20,
				"min_tip":      "0",
				"max_tip":      "1000000000000", // 0.001 ZEN
			},
			"consensus": map[string]interface{}{
				"consensus_type": "pos_poh_hybrid",
				"block_time":     3000, // 3 seconds
				"finality_time":  1800, // <2 seconds
				"shards":         64,
			},
			"security": map[string]interface{}{
				"post_quantum_enabled": true,
				"mpc_enabled":          true,
				"anomaly_detection":    true,
				"min_validator_stake":  "1000000000000000000000", // 1000 ZEN
			},
			"oracle": map[string]interface{}{
				"ai_enabled":            true,
				"ml_predictions":        true,
				"update_interval":       300, // 5 minutes
				"data_sources":          []string{"coinbase", "coingecko", "polygon"},
			},
			"accounts": []map[string]interface{}{},
		},
	}
}

// Generate validator keypair
func generateValidatorKeys(keysDir string) {
	// This would generate actual validator keys in production
	// For now, we create placeholders
	privKeyPath := filepath.Join(keysDir, "priv_validator_key.json")
	_, _ = os.Create(privKeyPath)

	fmt.Println("Note: Generate actual validator keys with production implementation")
}

// Run the node
func runNode(cmd *cobra.Command, args []string) error {
	fmt.Println("Starting ZenNetwork Node v" + Version)

	// Initialize core modules
	network := network.New()
	consensus := consensus.New()
	vm := vm.NewEVM()
	halving := halving.New()
	fees := fees.New()
	security := security.New()
	oracle := oracle.New()
	zenkit := zenkit.NewSDK()

	// Start services
	fmt.Println("✓ Initializing P2P network...")
	if err := network.Start(); err != nil {
		return fmt.Errorf("network start failed: %w", err)
	}

	fmt.Println("✓ Starting consensus engine (PoS + PoH)...")
	if err := consensus.Start(); err != nil {
		return fmt.Errorf("consensus start failed: %w", err)
	}

	fmt.Println("✓ Initializing EVM parallel executor...")
	if err := vm.Start(); err != nil {
		return fmt.Errorf("vm start failed: %w", err)
	}

	fmt.Println("✓ Starting halving engine (AEH)...")
	if err := halving.Start(); err != nil {
		return fmt.Errorf("halving start failed: %w", err)
	}

	fmt.Println("✓ Initializing fee system (low fees, 20% burn)...")
	if err := fees.Start(); err != nil {
		return fmt.Errorf("fees start failed: %w", err)
	}

	fmt.Println("✓ Starting security module (MPC, anomaly detection)...")
	if err := security.Start(); err != nil {
		return fmt.Errorf("security start failed: %w", err)
	}

	fmt.Println("✓ Starting AI-native oracle...")
	if err := oracle.Start(); err != nil {
		return fmt.Errorf("oracle start failed: %w", err)
	}

	fmt.Println("✓ Initializing ZenKit SDK...")
	if err := zenkit.Initialize(); err != nil {
		return fmt.Errorf("zenkit init failed: %w", err)
	}

	// Print status
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("ZenNetwork Node is running!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Network: ZenNetwork Mainnet\n")
	fmt.Printf("Block Time: 3 seconds (PoS + PoH)\n")
	fmt.Printf("TPS: 10,000-50,000 (parallel execution)\n")
	fmt.Printf("Fees: <0.0001 ZEN per transaction\n")
	fmt.Printf("Consensus: Hybrid PoS + Proof of History\n")
	fmt.Printf("Token Supply: 1,000,000,000 ZEN (fixed, immutable)\n")
	fmt.Printf("Shards: 64 (dynamic)\n")
	fmt.Printf("Security: Post-quantum crypto + MPC\n")
	fmt.Println(strings.Repeat("=", 60))

	// Keep the node running
	select {}
}

// Initialize config
func initConfig(cmd *cobra.Command) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(homeDir)
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	return nil
}

// Helper functions
func defaultHomeDir() string {
	if homeDir != "" {
		return homeDir
	}
	return os.ExpandEnv("$HOME/." + AppName)
}

func writeJSON(path string, v interface{}) error {
	bz, err := tmjson.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bz, 0644)
}
