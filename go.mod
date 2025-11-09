module github.com/zennetwork/zennetwork

go 1.22

require (
	// Cosmos SDK and Tendermint
	github.com/cosmos/cosmos-sdk v0.50.6
	github.com/tendermint/tendermint v0.37.14
	github.com/cometbft/cometbft v0.38.6

	// Ethereum and EVM
	github.com/ethereum/go-ethereum v1.14.2
	github.com/ethereum-optimism/optimism v1.0.0

	// P2P Networking
	github.com/libp2p/go-libp2p v0.33.1
	github.com/libp2p/go-libp2p-kad-dht v0.25.2
	github.com/libp2p/go-libp2p-pubsub v0.10.0

	// Cryptography and Security
	golang.org/x/crypto v0.23.0
	github.com/consensys/gnark v0.15.0
	github.com/pyroscope-io/pyroscope v0.0.0-20240312175230-02b420b9d9d2

	// Merkle Tree and Data Structures
	github.com/cbergoon/merkletree v0.2.0
	github.com/coinbase/rosetta-sdk-go v0.7.11

	// Protobuf and gRPC
	github.com/golang/protobuf v1.5.4
	google.golang.org/grpc v1.63.2
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0
	github.com/regen-network/protobuf v1.3.3-alpha.regen.5

	// AI/ML Integration
	github.com/owulveryck/onnx-go v0.0.0-20240312143317-9307b2b62c15
	github.com/schu/eGon-scraper v0.0.0-20240315195929-86e43e5b33d7

	// Logging and Monitoring
	github.com/rs/zerolog v1.32.0
	github.com/prometheus/client_golang v1.19.1

	// Configuration
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/pelletier/go-toml/v2 v2.2.2
)

require (
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.48.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2982fce20baf // indirect
	github.com/chzyer/test v0.0.0-20180213030017-19d2da6d51b6 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118203151-db636b1f1cd4 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/dgraph-io/badger/v3 v3.2103.3 // indirect
	github.com/docker/docker v25.0.5+incompatible // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvyukov/go-fuzz v0.0.0-20231101144951-6a623a6bb9a0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/godbus/dbus/v5 v5.1.1 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/pprof v0.0.0-20240424215950-a892ee059fd6 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.3 // indirect
	github.com/hashicorp/golang-lru v0.6.0 // indirect
	github.com/heroku/color v0.0.6 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-b14d5feef5d1 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/kisielk/errcheck v1.6.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p-asn-util v0.4.1 // indirect
	github.com/libp2p/go-libp2p-kbucket v0.6.3 // indirect
	github.com/libp2p/go-libp2p-logs v0.1.0 // indirect
	github.com/libp2p/go-mplex v0.10.0 // indirect
	github.com/libp2p/go-msgio v0.3.0 // indirect
	github.com/libp2p/go-nat v0.2.0 // indirect
	github.com/libp2p/go-netroute v0.2.0 // indirect
	github.com/libp2p/go-opentelemetry v0.1.0 // indirect
	github.com/libp2p/go-opentelemetry-example v0.1.0 // indirect
	github.com/libp2p/go-sockaddr v0.1.1 // indirect
	github.com/libp2p/go-stream-muxer v0.1.0 // indirect
	github.com/libp2p/go-transport v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/markbates/errx v1.0.0 // indirect
	github.com/markbates/inflect/v2 v2.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/dns v1.1.58 // indirect
	github.com/minio/sha256-simd v1.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mmcloughlin/bech32 v1.2.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multiaddr v0.12.4 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.0 // indirect
	github.com/multiformats/go-multistream v0.5.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/onsi/ginkgo/v2 v2.17.1 // indirect
	github.com/onsi/gomega v1.33.0 // indirect
	github.com/opencontainers/runtime-spec v1.1.0 // indirect
	github.com/orlangure/gnocll v0.0.0-20220405075352-ac95cdad5e10 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.6 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/exporter-toolkit v0.13.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/quic-go v0.43.1 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sasha-s/go-deadlock v0.2.1-0.20190427202633-8b97210b0ad4 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spacemonkeygo/openssl v0.0.0-20181017203307-f2d9b7024b63 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/status-im/keycard-go v0.2.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6e // indirect
	github.com/tendermint/btcd v0.1.1 // indirect
	github.com/tendermint/crypto v0.0.0-20191022155603-50747014af02 // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	github.com/tendermint/iavl v0.20.6 // indirect
	github.com/tendermint/tendermint v0.37.14 // indirect
	github.com/tendermint/tm-db v0.0.5 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/ytids/zen-vdf v0.0.0-20230810210041-8c0b8b34b1a3 // indirect
	github.com/yuin/goldmark v1.7.1 // indirect
	go.etcd.io/bbolt v1.3.8 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	go.uber.org/automaxprocs v1.5.3 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/term v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8afa5bcdd // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.5.1 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
)
// Replace problematic dependencies
replace github.com/chzyer/readline => github.com/chzyer/readline v1.0.0
replace github.com/chzyer/logex => github.com/chzyer/logex v1.0.0
replace github.com/chzyer/test => github.com/chzyer/test v0.0.0
