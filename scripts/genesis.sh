#!/bin/bash

# ZenNetwork Genesis Generator
# Creates initial genesis.json with fixed 1B ZEN supply

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== ZenNetwork Genesis Generator ===${NC}\n"

# Configuration
CHAIN_ID="zennetwork-mainnet-1"
TOTAL_SUPPLY="1000000000000000000000000000"  # 1B ZEN in wei
GENESIS_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "Configuration:"
echo "  Chain ID: $CHAIN_ID"
echo "  Total Supply: 1,000,000,000 ZEN (Fixed & Immutable)"
echo "  Genesis Time: $GENESIS_TIME"
echo ""

# Create genesis directory
GENESIS_DIR="./config/genesis"
mkdir -p "$GENESIS_DIR"

# Generate genesis.json
cat > "$GENESIS_DIR/genesis.json" <<EOF
{
  "genesis_time": "$GENESIS_TIME",
  "chain_id": "$CHAIN_ID",
  "consensus_params": {
    "block": {
      "max_bytes": 104857600,
      "max_gas": 100000000,
      "time_iota_ms": 3000
    },
    "evidence": {
      "max_age_num_blocks": 100000,
      "max_age_duration": "172800000000000"
    },
    "validator": {
      "pub_key_types": ["bls", "ed25519"]
    }
  },
  "app_hash": "",
  "app_state": {
    "tokenomics": {
      "total_supply": "$TOTAL_SUPPLY",
      "is_fixed": true,
      "minting_disabled": true,
      "burn_enabled": true,
      "community_allocation": "400000000000000000000000000",
      "team_allocation": "200000000000000000000000000",
      "ecosystem_allocation": "200000000000000000000000000",
      "liquidity_allocation": "100000000000000000000000000",
      "foundation_allocation": "100000000000000000000000000"
    },
    "halving": {
      "total_pool": "200000000000000000000000000",
      "initial_reward": "1000000000000000000000",
      "halving_factor": 0.95,
      "halving_interval": 7889400,
      "adaptive_enabled": true
    },
    "fees": {
      "base_fee": "100000000000000",
      "burn_percent": 20,
      "min_tip": "0",
      "max_tip": "1000000000000"
    },
    "consensus": {
      "consensus_type": "pos_poh_hybrid",
      "block_time": 3000,
      "finality_time": 1800,
      "shards": 64
    },
    "security": {
      "post_quantum_enabled": true,
      "mpc_enabled": true,
      "anomaly_detection": true,
      "min_validator_stake": "1000000000000000000000"
    },
    "oracle": {
      "ai_enabled": true,
      "ml_predictions": true,
      "update_interval": 300,
      "data_sources": ["coinbase", "coingecko", "polygon"]
    },
    "accounts": []
  }
}
EOF

echo -e "${GREEN}✓ Genesis file created: $GENESIS_DIR/genesis.json${NC}\n"

# Validate genesis
echo "Validating genesis..."
if command -v jq &> /dev/null; then
    if jq empty "$GENESIS_DIR/genesis.json" 2>/dev/null; then
        echo -e "${GREEN}✓ Genesis is valid JSON${NC}"
    else
        echo -e "${RED}✗ Genesis is invalid JSON${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ jq not found, skipping JSON validation${NC}"
fi

# Display summary
echo ""
echo -e "${GREEN}=== Genesis Summary ===${NC}"
echo "Total Supply: 1,000,000,000 ZEN (Fixed, Immutable)"
echo ""
echo "Distribution:"
echo "  Community:     40% (400M ZEN)"
echo "  Team:          20% (200M ZEN, 4y vest)"
echo "  Ecosystem:     20% (200M ZEN, via rewards)"
echo "  Liquidity:     10% (100M ZEN)"
echo "  Foundation:    10% (100M ZEN, 2y vest)"
echo ""
echo "System Configuration:"
echo "  Consensus:     PoS + PoH Hybrid (3s blocks, <2s finality)"
echo "  Fees:          <0.0001 ZEN (20% burn)"
echo "  Halving:       AEH from 200M ZEN pool"
echo "  Security:      Post-quantum + MPC"
echo "  Shards:        64"
echo "  AI Oracles:    Enabled (ML predictions)"
echo ""
echo -e "${GREEN}Genesis generation complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review genesis file: $GENESIS_DIR/genesis.json"
echo "  2. Add initial validators to accounts array"
echo "  3. Initialize node: ./zennetworkd init mynode --chain-id $CHAIN_ID"
echo "  4. Start node: ./zennetworkd start --genesis $GENESIS_DIR/genesis.json"
echo ""
