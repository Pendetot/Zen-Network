Zen Network - Quantum-Resistant Layer-1 Blockchain
==================================================

What is Zen Network?
--------------------

Zen Network is a next-generation Layer-1 blockchain platform built in Go, designed for AI-driven decentralized applications. It combines high throughput, quantum-resistant security, and low fees.

Key Features
------------

- High Performance: 10,000-50,000 TPS with parallel execution
- Low Fees: <0.0001 ZEN per transaction
- Quantum-Resistant Security: EdDSA, Blake3, BLS, Falcon/Dilithium
- AI-Native Oracles: Machine learning powered data feeds
- Fixed Supply: 1,000,000,000 ZEN tokens (no inflation)
- EVM Compatible: Run Ethereum smart contracts

Quick Start
-----------

Requirements:
- Go 1.22 or higher
- 8+ CPU cores
- 16GB+ RAM
- 1TB+ storage

Build from source:
```
git clone https://github.com/Pendetot/Zen-Network.git
cd Zen-Network
go build -o zennetworkd ./cmd/zennetworkd/
```

Run a node:
```
./zennetworkd init mynode
./zennetworkd start
```

Documentation
-------------

More detailed documentation can be found in:
- whitepaper.md - Technical whitepaper
- config/genesis/ - Genesis configuration
- x/ - Core blockchain modules

License
-------

MIT License - see LICENSE file for details

Contributing
------------

Contributions are welcome! Please feel free to submit a Pull Request.
