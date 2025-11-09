package network

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
)

// NetworkProtocol defines ZenNetwork P2P protocols
const (
	ProtocolID        = "/zennetwork/1.0.0"
	ConsensusProtocol = "/zennetwork/consensus/1.0.0"
	TxProtocol        = "/zennetwork/tx/1.0.0"
	SyncProtocol      = "/zennetwork/sync/1.0.0"
	StateProtocol     = "/zennetwork/state/1.0.0"
)

// Message types for P2P communication
type MessageType uint8

const (
	MsgTypeTx      MessageType = 0x01
	MsgTypeBlock   MessageType = 0x02
	MsgTypeStatus  MessageType = 0x03
	MsgTypeSync    MessageType = 0x04
	MsgTypeConsensus MessageType = 0x05
	MsgTypeState   MessageType = 0x06
)

// NetworkMessage represents a P2P message
type NetworkMessage struct {
	Type      MessageType `json:"type"`
	Data      []byte      `json:"data"`
	Timestamp int64       `json:"timestamp"`
	PeerID    peer.ID     `json:"peer_id"`
}

// PeerInfo represents network peer information
type PeerInfo struct {
	ID             peer.ID        `json:"id"`
	Addresses      []multiaddr.Multiaddr `json:"addresses"`
	ConnectionTime time.Time      `json:"connection_time"`
	Latency        time.Duration  `json:"latency"`
	BytesIn        uint64         `json:"bytes_in"`
	BytesOut       uint64         `json:"bytes_out"`
	Score          float64        `json:"score"` // Trust score
	Trusted        bool           `json:"trusted"` // Trusted validator
	Validator      bool           `json:"validator"`
	Protocols      []protocol.ID  `json:"protocols"`
}

// Network handles P2P communication
type Network struct {
	mu           sync.RWMutex
	host         host.Host
	ctx          context.Context
	cancel       context.CancelFunc
	selfID       peer.ID
	privateKey   ed25519.PrivateKey
	listener     network.Listener
	peers        map[peer.ID]*PeerInfo
	messageCh    chan NetworkMessage
	running      bool
	listeners    map[MessageType]func(NetworkMessage)
	muListeners  sync.RWMutex
}

// New creates a new Network instance
func New() *Network {
	ctx, cancel := context.WithCancel(context.Background())

	// Generate random keypair for node identity
	_, priv, _ := ed25519.GenerateKey(rand.Reader)

	n := &Network{
		ctx:         ctx,
		cancel:      cancel,
		privateKey:  priv,
		peers:       make(map[peer.ID]*PeerInfo),
		messageCh:   make(chan NetworkMessage, 1000),
		running:     false,
		listeners:   make(map[MessageType]func(NetworkMessage)),
	}

	return n
}

// Start initializes and starts the P2P network
func (n *Network) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Println("[NETWORK] Starting libp2p P2P network")

	// Create libp2p host with security
	host, err := libp2p.New(
		// Use Ed25519 for identity
		libp2p.Identity(n.privateKey),

		// Enable TLS 1.3 security
		// In production: custom libp2p security with post-quantum crypto
		libp2p.Security(tls.ID, tls.New),

		// Enable QUIC transport for high performance
		// QUIC is faster than TCP and supports multiplexing
		// In production: add custom QUIC transport

		// Enable connection manager for peer management
		// In production: configure limits

		// Enable relay for NAT traversal
		// In production: configure circuit relay

		// Custom peer scoring
		// In production: implement scoring system
	)
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}

	n.host = host
	n.selfID = host.ID()

	// Set up stream handlers
	n.setupStreamHandlers()

	// Start listening on default ports
	if err := n.startListening(); err != nil {
		return fmt.Errorf("failed to start listening: %w", err)
	}

	// Start peer management
	go n.peerManager()

	// Start message handling
	go n.messageHandler()

	// Connect to bootstrap peers
	// In production: actual bootstrap nodes
	n.bootstrap()

	n.running = true

	fmt.Printf("  - Node ID: %s\n", n.selfID.String())
	fmt.Printf("  - Listen Addrs: %d\n", len(host.Addrs()))
	fmt.Printf("  - Protocol: %s\n", ProtocolID)
	fmt.Printf("  - Security: TLS 1.3 + EdDSA\n")
	fmt.Printf("  - Transport: QUIC + TCP\n")

	return nil
}

// Stop halts the network
func (n *Network) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running {
		return nil
	}

	fmt.Println("[NETWORK] Stopping P2P network")

	// Close all peer connections
	for peerID := range n.peers {
		n.host.Network().ClosePeer(peerID)
	}

	// Close host
	if n.host != nil {
		n.host.Close()
	}

	// Cancel context
	n.cancel()

	n.running = false
	return nil
}

// startListening sets up network listeners
func (n *Network) startListening() error {
	// Listen on TCP and QUIC
	addrs := []string{
		"/ip4/0.0.0.0/tcp/26656",  // P2P port
		"/ip4/0.0.0.0/udp/26656/quic", // QUIC port
	}

	for _, addrStr := range addrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return fmt.Errorf("invalid multiaddr: %w", err)
		}

		if err := n.host.Network().Listen(addr); err != nil {
			fmt.Printf("Warning: Failed to listen on %s: %v\n", addrStr, err)
		} else {
			fmt.Printf("  - Listening on: %s\n", addrStr)
		}
	}

	return nil
}

// setupStreamHandlers configures protocol handlers
func (n *Network) setupStreamHandlers() {
	// Consensus protocol
	n.host.SetStreamHandler(ConsensusProtocol, n.handleConsensusStream)

	// Transaction protocol
	n.host.SetStreamHandler(TxProtocol, n.handleTxStream)

	// Sync protocol
	n.host.SetStreamHandler(SyncProtocol, n.handleSyncStream)

	// State protocol
	n.host.SetStreamHandler(StateProtocol, n.handleStateStream)
}

// handleConsensusStream handles consensus messages
func (n *Network) handleConsensusStream(stream network.Stream) {
	defer stream.Close()

	msg, err := n.readMessage(stream)
	if err != nil {
		return
	}

	if msg.Type == MsgTypeConsensus {
		n.dispatchMessage(msg)
	}
}

// handleTxStream handles transaction messages
func (n *Network) handleTxStream(stream network.Stream) {
	defer stream.Close()

	msg, err := n.readMessage(stream)
	if err != nil {
		return
	}

	if msg.Type == MsgTypeTx {
		n.dispatchMessage(msg)
	}
}

// handleSyncStream handles synchronization messages
func (n *Network) handleSyncStream(stream network.Stream) {
	defer stream.Close()

	msg, err := n.readMessage(stream)
	if err != nil {
		return
	}

	if msg.Type == MsgTypeSync {
		n.dispatchMessage(msg)
	}
}

// handleStateStream handles state sync messages
func (n *Network) handleStateStream(stream network.Stream) {
	defer stream.Close()

	msg, err := n.readMessage(stream)
	if err != nil {
		return
	}

	if msg.Type == MsgTypeState {
		n.dispatchMessage(msg)
	}
}

// readMessage reads a message from a stream
func (n *Network) readMessage(stream network.Stream) (NetworkMessage, error) {
	// In production: implement proper binary encoding
	// For now: simplified message reading
	buf := make([]byte, 4096)
	read, err := stream.Read(buf)
	if err != nil {
		return NetworkMessage{}, err
	}

	msg := NetworkMessage{
		Type:      MessageType(buf[0]),
		Data:      buf[1:read],
		Timestamp: time.Now().Unix(),
		PeerID:    stream.Conn().RemotePeer(),
	}

	return msg, nil
}

// writeMessage writes a message to a stream
func (n *Network) writeMessage(stream network.Stream, msg NetworkMessage) error {
	// In production: implement proper binary encoding
	data := make([]byte, len(msg.Data)+1)
	data[0] = byte(msg.Type)
	copy(data[1:], msg.Data)

	_, err := stream.Write(data)
	return err
}

// SendMessage sends a message to a specific peer
func (n *Network) SendMessage(peerID peer.ID, msg NetworkMessage) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.host.Network().Connectedness(peerID).IsConnected() {
		return fmt.Errorf("not connected to peer: %s", peerID.String())
	}

	stream, err := n.host.NewStream(context.Background(), peerID, ProtocolID)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	return n.writeMessage(stream, msg)
}

// BroadcastMessage broadcasts a message to all connected peers
func (n *Network) BroadcastMessage(msg NetworkMessage) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for peerID := range n.peers {
		if n.host.Network().Connectedness(peerID).IsConnected() {
			// Fire and forget - don't block on each peer
			go func(pid peer.ID) {
				stream, err := n.host.NewStream(context.Background(), pid, ProtocolID)
				if err != nil {
					return
				}
				defer stream.Close()
				n.writeMessage(stream, msg)
			}(peerID)
		}
	}

	return nil
}

// ConnectToPeer establishes a connection to a peer
func (n *Network) ConnectToPeer(addr multiaddr.Multiaddr) error {
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %w", err)
	}

	if err := n.host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	n.mu.Lock()
	n.peers[peerInfo.ID] = &PeerInfo{
		ID:             peerInfo.ID,
		Addresses:      peerInfo.Addrs,
		ConnectionTime: time.Now(),
		Score:          1.0,
		Trusted:        false,
		Validator:      false,
		Protocols:      make([]protocol.ID, 0),
	}
	n.mu.Unlock()

	fmt.Printf("[NETWORK] Connected to peer: %s\n", peerInfo.ID.String())
	return nil
}

// DisconnectFromPeer closes connection to a peer
func (n *Network) DisconnectFromPeer(peerID peer.ID) error {
	n.host.Network().ClosePeer(peerID)

	n.mu.Lock()
	delete(n.peers, peerID)
	n.mu.Unlock()

	fmt.Printf("[NETWORK] Disconnected from peer: %s\n", peerID.String())
	return nil
}

// GetPeers returns all connected peers
func (n *Network) GetPeers() map[peer.ID]*PeerInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make(map[peer.ID]*PeerInfo)
	for id, info := range n.peers {
		peers[id] = info
	}

	return peers
}

// GetPeerCount returns the number of connected peers
func (n *Network) GetPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.peers)
}

// RegisterListener registers a message listener
func (n *Network) RegisterListener(msgType MessageType, handler func(NetworkMessage)) {
	n.muListeners.Lock()
	defer n.muListeners.Unlock()
	n.listeners[msgType] = handler
}

// dispatchMessage dispatches a message to registered listeners
func (n *Network) dispatchMessage(msg NetworkMessage) {
	n.muListeners.RLock()
	defer n.muListeners.RUnlock()

	if handler, ok := n.listeners[msg.Type]; ok {
		handler(msg)
	}
}

// messageHandler processes incoming messages
func (n *Network) messageHandler() {
	for {
		select {
		case msg := <-n.messageCh:
			n.dispatchMessage(msg)
		case <-n.ctx.Done():
			return
		}
	}
}

// peerManager manages peer connections and health
func (n *Network) peerManager() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.mu.Lock()
			for peerID, info := range n.peers {
				// Check connection health
				if !n.host.Network().Connectedness(peerID).IsConnected() {
					delete(n.peers, peerID)
					continue
				}

				// Update latency
				conn := n.host.Network().ConnsToPeer(peerID)
				if len(conn) > 0 {
					// In production: measure actual latency
					info.Latency = 10 * time.Millisecond
				}

				// Update score based on various factors
				// In production: implement proper scoring algorithm
				info.Score = 1.0
			}
			n.mu.Unlock()
		case <-n.ctx.Done():
			return
		}
	}
}

// bootstrap connects to initial peers
func (n *Network) bootstrap() {
	// In production: actual bootstrap nodes
	// For now: no-op
	fmt.Println("[NETWORK] Bootstrap peers: (none configured)")
}

// IsRunning returns network status
func (n *Network) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}

// GetNodeID returns the node's peer ID
func (n *Network) GetNodeID() peer.ID {
	return n.selfID
}

// GetListenAddresses returns all listen addresses
func (n *Network) GetListenAddresses() []multiaddr.Multiaddr {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.host.Addrs()
}
