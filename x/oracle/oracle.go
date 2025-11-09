package oracle

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/owulveryck/onnx-go"
)

// OracleType represents different oracle types
type OracleType string

const (
	PriceOracle    OracleType = "price"     // Token prices
	MLOracle      OracleType = "ml"        // ML predictions
	ConsensusOracle OracleType = "consensus" // Consensus data
	ESGOracle     OracleType = "esg"       // ESG metrics
)

// DataPoint represents a single data point
type DataPoint struct {
	Value     float64   `json:"value"`
	Source    string    `json:"source"`
	Timestamp int64     `json:"timestamp"`
	Signature []byte    `json:"signature"`
	Verified  bool      `json:"verified"`
}

// PriceData represents token price data
type PriceData struct {
	Symbol    string    `json:"symbol"`
	USD       float64   `json:"usd"`
	Change24h float64   `json:"change_24h"`
	Volume    float64   `json:"volume"`
	MarketCap float64   `json:"market_cap"`
	Source    string    `json:"source"`
	Timestamp int64     `json:"timestamp"`
}

// MLPrediction represents a machine learning prediction
type MLPrediction struct {
	Model        string                 `json:"model"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	Confidence   float64                `json:"confidence"`
	Accuracy     float64                `json:"accuracy"`
	Timestamp    int64                  `json:"timestamp"`
	Horizon      int                    `json:"horizon"` // Prediction horizon in hours
}

// Oracle represents an AI-native oracle
type Oracle struct {
	mu              sync.RWMutex
	oracleType      OracleType
	models          map[string]*onnx.Model
	dataPoints      map[string][]DataPoint
	priceData       map[string]*PriceData
	predictions     map[string]*MLPrediction
	anomalyDetector *AnomalyDetector
	running         bool
	updateInterval  time.Duration
}

// AnomalyDetector detects anomalous data
type AnomalyDetector struct {
	mu         sync.RWMutex
	threshold  float64
	data       []float64
	modelType  string
}

// New creates a new oracle instance
func New() *Oracle {
	return &Oracle{
		oracleType:     MLOracle,
		models:         make(map[string]*onnx.Model),
		dataPoints:     make(map[string][]DataPoint),
		priceData:      make(map[string]*PriceData),
		predictions:    make(map[string]*MLPrediction),
		anomalyDetector: &AnomalyDetector{
			threshold: 3.0, // 3-sigma rule
		},
		running:        false,
		updateInterval: 300 * time.Second, // 5 minutes
	}
}

// NewPriceOracle creates a price oracle
func NewPriceOracle() *Oracle {
	oracle := New()
	oracle.oracleType = PriceOracle
	oracle.updateInterval = 30 * time.Second // Faster updates for prices
	return oracle
}

// NewConsensusOracle creates a consensus oracle
func NewConsensusOracle() *Oracle {
	oracle := New()
	oracle.oracleType = ConsensusOracle
	oracle.updateInterval = 3 * time.Second // Sync with block time
	return oracle
}

// Initialize sets up the oracle
func (o *Oracle) Initialize(dataDir string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	fmt.Println("[ORACLE] Initializing AI-native oracle")
	fmt.Printf("  - Type: %s\n", o.oracleType)
	fmt.Printf("  - Update Interval: %v\n", o.updateInterval)
	fmt.Printf("  - Data Sources: CoinGecko, CoinMarketCap, Internal ML\n")

	// Load ML models
	if err := o.loadModels(); err != nil {
		return fmt.Errorf("failed to load models: %w", err)
	}

	// Start update loop
	o.running = true
	go o.updateLoop()

	fmt.Println("✓ Oracle initialized")

	return nil
}

// Start begins oracle operation
func (o *Oracle) Start() error {
	if o.running {
		return nil
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	fmt.Println("[ORACLE] Starting AI-native oracle")
	o.running = true
	go o.updateLoop()

	return nil
}

// Stop halts the oracle
func (o *Oracle) Stop() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.running {
		return nil
	}

	fmt.Println("[ORACLE] Stopping oracle")
	o.running = false

	return nil
}

// UpdatePriceData updates price information
func (o *Oracle) UpdatePriceData(symbol string, data *PriceData) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Validate data
	if err := o.validatePriceData(data); err != nil {
		return fmt.Errorf("invalid price data: %w", err)
	}

	// Check for anomalies
	if o.anomalyDetector != nil {
		if o.isAnomaly(data.USD) {
			fmt.Printf("[ORACLE] Anomaly detected for %s: $%.2f (outlier)\n", symbol, data.USD)
			// In production: alert or reject data
		}
	}

	o.priceData[symbol] = data
	o.addDataPoint(symbol, DataPoint{
		Value:     data.USD,
		Source:    data.Source,
		Timestamp: data.Timestamp,
		Verified:  true,
	})

	return nil
}

// GetPriceData retrieves price data
func (o *Oracle) GetPriceData(symbol string) (*PriceData, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	data, ok := o.priceData[symbol]
	if !ok {
		return nil, fmt.Errorf("no price data for %s", symbol)
	}

	return data, nil
}

// GeneratePrediction generates ML prediction
func (o *Oracle) GeneratePrediction(modelName string, input map[string]interface{}) (*MLPrediction, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// In production: use actual ONNX model
	// For now: mock prediction
	prediction := &MLPrediction{
		Model:      modelName,
		Input:      input,
		Output:     make(map[string]interface{}),
		Confidence: 0.95,
		Accuracy:   0.92,
		Timestamp:  time.Now().Unix(),
		Horizon:    24, // 24 hours
	}

	// Simple mock output (price trend)
	if price, ok := input["price"].(float64); ok {
		prediction.Output["predicted_price"] = price * 1.05
		prediction.Output["trend"] = "up"
		prediction.Output["volatility"] = 0.15
	}

	o.predictions[modelName] = prediction

	return prediction, nil
}

// GetPrediction retrieves a prediction
func (o *Oracle) GetPrediction(modelName string) (*MLPrediction, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pred, ok := o.predictions[modelName]
	if !ok {
		return nil, fmt.Errorf("no prediction for model %s", modelName)
	}

	return pred, nil
}

// UpdateConsensusData updates consensus-related oracle data
func (o *Oracle) UpdateConsensusData(data map[string]interface{}) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Store consensus data
	jsonData, _ := json.Marshal(data)
	o.addDataPoint("consensus", DataPoint{
		Value:     0, // Placeholder
		Source:    "internal",
		Timestamp: time.Now().Unix(),
		Data:      jsonData,
	})

	return nil
}

// ValidateData validates oracle data
func (o *Oracle) ValidateData(symbol string, value float64) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Check if value is within reasonable bounds
	if value < 0 {
		return false
	}

	// Compare with historical data
	points := o.dataPoints[symbol]
	if len(points) < 2 {
		return true
	}

	// Calculate moving average
	var sum float64
	for _, p := range points {
		sum += p.Value
	}
	avg := sum / float64(len(points))

	// Check if value is within 5 standard deviations
	var variance float64
	for _, p := range points {
		diff := p.Value - avg
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(points)))

	if stdDev == 0 {
		return true
	}

	zScore := math.Abs(value-avg) / stdDev
	return zScore < 5.0 // Within 5-sigma
}

// addDataPoint adds a data point
func (o *Oracle) addDataPoint(key string, point DataPoint) {
	if o.dataPoints[key] == nil {
		o.dataPoints[key] = make([]DataPoint, 0)
	}

	o.dataPoints[key] = append(o.dataPoints[key], point)

	// Keep only recent data (last 1000 points)
	if len(o.dataPoints[key]) > 1000 {
		o.dataPoints[key] = o.dataPoints[key][1:]
	}
}

// isAnomaly checks if value is anomalous
func (o *Oracle) isAnomaly(value float64) bool {
	o.anomalyDetector.mu.Lock()
	defer o.anomalyDetector.mu.Unlock()

	o.anomalyDetector.data = append(o.anomalyDetector.data, value)

	if len(o.anomalyDetector.data) < 10 {
		return false
	}

	// Calculate mean and std dev
	var sum, mean, variance float64
	for _, v := range o.anomalyDetector.data {
		sum += v
	}
	mean = sum / float64(len(o.anomalyDetector.data))

	for _, v := range o.anomalyDetector.data {
		diff := v - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(o.anomalyDetector.data)))

	if stdDev == 0 {
		return false
	}

	// Check if value is beyond threshold
	zScore := math.Abs(value-mean) / stdDev
	return zScore > o.anomalyDetector.threshold
}

// validatePriceData validates price data
func (o *Oracle) validatePriceData(data *PriceData) error {
	if data.USD <= 0 {
		return fmt.Errorf("invalid price: %f", data.USD)
	}
	if data.Timestamp > time.Now().Unix() {
		return fmt.Errorf("future timestamp")
	}
	if data.Source == "" {
		return fmt.Errorf("missing source")
	}

	return nil
}

// loadModels loads ML models
func (o *Oracle) loadModels() error {
	// In production: load actual ONNX models
	// For now: mock
	fmt.Println("[ORACLE] Loading ML models:")
	fmt.Println("  - Price prediction model (Transformer)")
	fmt.Println("  - Anomaly detection model (Isolation Forest)")
	fmt.Println("  - Volatility model (LSTM)")
	fmt.Println("  - Market sentiment model (BERT)")

	return nil
}

// updateLoop continuously updates oracle data
func (o *Oracle) updateLoop() {
	ticker := time.NewTicker(o.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !o.running {
				return
			}
			o.update()
		}
	}
}

// update performs data updates
func (o *Oracle) update() {
	// In production: fetch from actual data sources
	// For now: mock updates

	switch o.oracleType {
	case PriceOracle:
		o.mockPriceUpdate()
	case MLOracle:
		o.mockMLUpdate()
	case ConsensusOracle:
		o.mockConsensusUpdate()
	}
}

// mockPriceUpdate simulates price updates
func (o *Oracle) mockPriceUpdate() {
	coins := []string{"BTC", "ETH", "SOL", "ZEN", "USDC"}

	for _, symbol := range coins {
		// Mock price data
		price := 100.0 + float64(time.Now().Unix()%10000)/100
		change := (float64(time.Now().Unix()%200) - 100) / 10

		data := &PriceData{
			Symbol:    symbol,
			USD:       price,
			Change24h: change,
			Volume:    1000000,
			MarketCap: price * 1000000,
			Source:    "mock-oracle",
			Timestamp: time.Now().Unix(),
		}

		_ = o.UpdatePriceData(symbol, data)
	}

	// Print current prices
	o.mu.RLock()
	defer o.mu.RUnlock()
	fmt.Print("[ORACLE] Prices: ")
	for symbol, data := range o.priceData {
		fmt.Printf("%s: $%.2f (%.1f%%) ", symbol, data.USD, data.Change24h)
	}
	fmt.Println()
}

// mockMLUpdate simulates ML predictions
func (o *Oracle) mockMLUpdate() {
	input := map[string]interface{}{
		"price":   100.0,
		"volume":  1000000,
		"time":    time.Now().Unix(),
		"sentiment": 0.75,
	}

	if _, err := o.GeneratePrediction("price_prediction", input); err == nil {
		// Prediction generated
	}
}

// mockConsensusUpdate simulates consensus data
func (o *Oracle) mockConsensusUpdate() {
	data := map[string]interface{}{
		"validators":       1000,
		"total_stake":      5000000000000000000000, // 5M ZEN
		"avg_block_time":   3.0,
		"consensus_quorum": 0.67,
	}

	_ = o.UpdateConsensusData(data)
}

// GetStats returns oracle statistics
func (o *Oracle) GetStats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return map[string]interface{}{
		"type":             o.oracleType,
		"running":          o.running,
		"update_interval":  o.updateInterval.String(),
		"data_points":      len(o.dataPoints),
		"predictions":      len(o.predictions),
		"price_pairs":      len(o.priceData),
		"anomaly_threshold": o.anomalyDetector.threshold,
	}
}
