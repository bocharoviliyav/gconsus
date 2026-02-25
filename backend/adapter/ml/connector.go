// Package ml provides an inactive connector to an external ML prediction service.
// This connector is designed for future integration with predictive analytics models
// (time series, LSTM, Samba, KAN architectures) for forecasting developer activity.
//
// The connector is intentionally disabled by default. To activate it, set
// ML_ENABLED=true and ML_ENDPOINT to the model serving URL.
package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PredictionRequest describes the input payload sent to the ML service.
type PredictionRequest struct {
	UserID      string    `json:"user_id"`
	MetricName  string    `json:"metric_name"` // commits, prs, reviews, issues
	HistoryDays int       `json:"history_days"`
	ForecastDays int      `json:"forecast_days"`
	AsOf        time.Time `json:"as_of"`
}

// PredictionResponse describes the output from the ML service.
type PredictionResponse struct {
	UserID       string           `json:"user_id"`
	MetricName   string           `json:"metric_name"`
	Predictions  []PredictionPoint `json:"predictions"`
	ModelName    string           `json:"model_name"`
	Confidence   float64          `json:"confidence"`
	GeneratedAt  time.Time        `json:"generated_at"`
}

// PredictionPoint is a single forecast data point.
type PredictionPoint struct {
	Date  string  `json:"date"`  // YYYY-MM-DD
	Value float64 `json:"value"`
	Lower float64 `json:"lower"` // confidence interval lower bound
	Upper float64 `json:"upper"` // confidence interval upper bound
}

// Client is the interface that any ML prediction backend must satisfy.
type Client interface {
	// Predict sends a prediction request and returns the forecast.
	Predict(ctx context.Context, req PredictionRequest) (*PredictionResponse, error)

	// HealthCheck verifies the ML service is reachable.
	HealthCheck(ctx context.Context) error

	// IsEnabled reports whether the connector is active.
	IsEnabled() bool
}

// Config holds configuration for the ML connector.
type Config struct {
	Enabled  bool   `env:"ML_ENABLED" envDefault:"false"`
	Endpoint string `env:"ML_ENDPOINT" envDefault:"http://localhost:8501"`
	Timeout  int    `env:"ML_TIMEOUT" envDefault:"30"` // seconds
}

// ---------------------------------------------------------------------------
// Stub implementation (inactive by default)
// ---------------------------------------------------------------------------

// StubClient is a no-op implementation returned when ML_ENABLED=false.
type StubClient struct{}

func NewStub() *StubClient { return &StubClient{} }

func (s *StubClient) Predict(_ context.Context, _ PredictionRequest) (*PredictionResponse, error) {
	return nil, fmt.Errorf("ml connector is disabled")
}

func (s *StubClient) HealthCheck(_ context.Context) error {
	return fmt.Errorf("ml connector is disabled")
}

func (s *StubClient) IsEnabled() bool { return false }

// ---------------------------------------------------------------------------
// HTTP implementation (used when ML_ENABLED=true)
// ---------------------------------------------------------------------------

// HTTPClient communicates with an external ML model serving endpoint.
type HTTPClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewHTTPClient creates an active ML connector.
func NewHTTPClient(cfg Config) *HTTPClient {
	return &HTTPClient{
		endpoint: cfg.Endpoint,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

func (c *HTTPClient) IsEnabled() bool { return true }

func (c *HTTPClient) Predict(ctx context.Context, req PredictionRequest) (*PredictionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ml: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/predict", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Body = io.NopCloser(jsonReader(body))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ml: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml: status %d", resp.StatusCode)
	}

	var result PredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ml: decode response: %w", err)
	}
	return &result, nil
}

func (c *HTTPClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ml: health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ml: health check status %d", resp.StatusCode)
	}
	return nil
}

// NewClient returns an active or stub client based on the configuration.
func NewClient(cfg Config) Client {
	if !cfg.Enabled {
		return NewStub()
	}
	return NewHTTPClient(cfg)
}

// jsonReader wraps a byte slice into an io.Reader.
func jsonReader(data []byte) io.Reader {
	return &byteReader{data: data}
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
