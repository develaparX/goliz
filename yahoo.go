package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Yahoo Finance API base URL
const YahooFinanceBaseURL = "https://query1.finance.yahoo.com/v8/finance/chart"

// YahooInterval represents Yahoo Finance chart intervals
type YahooInterval string

const (
	YahooInterval1m  YahooInterval = "1m"
	YahooInterval2m  YahooInterval = "2m"
	YahooInterval5m  YahooInterval = "5m"
	YahooInterval15m YahooInterval = "15m"
	YahooInterval30m YahooInterval = "30m"
	YahooInterval90m YahooInterval = "90m"
	YahooInterval1h  YahooInterval = "1h"
	YahooInterval1d  YahooInterval = "1d"
	YahooInterval1wk YahooInterval = "1wk"
	YahooInterval1mo YahooInterval = "1mo"
	YahooInterval3mo YahooInterval = "3mo"
)

// YahooRange represents the data range for Yahoo Finance
type YahooRange string

const (
	YahooRange1d  YahooRange = "1d"
	YahooRange5d  YahooRange = "5d"
	YahooRange1mo YahooRange = "1mo"
	YahooRange3mo YahooRange = "3mo"
	YahooRange6mo YahooRange = "6mo"
	YahooRange1y  YahooRange = "1y"
	YahooRange2y  YahooRange = "2y"
	YahooRange5y  YahooRange = "5y"
	YahooRangeMax YahooRange = "max"
)

// YahooChartResponse represents the API response from Yahoo Finance
type YahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string  `json:"currency"`
				Symbol             string  `json:"symbol"`
				ExchangeName       string  `json:"exchangeName"`
				InstrumentType     string  `json:"instrumentType"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
				Timezone           string  `json:"timezone"`
				DataGranularity    string  `json:"dataGranularity"`
				Range              string  `json:"range"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// GetRangeForInterval returns the appropriate range for a given interval
// to ensure we get enough candles
func GetRangeForInterval(interval YahooInterval) YahooRange {
	switch interval {
	case YahooInterval1m:
		return YahooRange1d // 1m data only available for 1 day
	case YahooInterval2m:
		return YahooRange1d // 2m data for 1 day
	case YahooInterval5m:
		return YahooRange5d // 5m data for 5 days
	case YahooInterval15m:
		return YahooRange1mo // 15m data for 1 month
	case YahooInterval30m:
		return YahooRange1mo
	case YahooInterval90m:
		return YahooRange1mo // 90m data (max 60 days allowed, safe with 1mo)
	case YahooInterval1h:
		return YahooRange3mo // 1h data for 3 months
	case YahooInterval1d:
		return YahooRange1y // 1d data for 1 year
	case YahooInterval1wk:
		return YahooRange5y // 1wk data for 5 years
	case YahooInterval1mo:
		return YahooRangeMax // 1mo data max range
	case YahooInterval3mo:
		return YahooRangeMax // 3mo data max range
	default:
		return YahooRange1mo
	}
}

// FetchYahooCandlesticks fetches OHLCV data from Yahoo Finance API
// symbol: e.g., "EURUSD=X" for forex, "AAPL" for stocks
// interval: e.g., "5m", "1h", "1d"
// limit: maximum number of candles to return
func FetchYahooCandlesticks(symbol string, interval YahooInterval, limit int) ([]Candlestick, error) {
	if limit < 1 {
		limit = 200
	}

	// Get appropriate range for the interval
	yahooRange := GetRangeForInterval(interval)

	// Build URL
	url := fmt.Sprintf("%s/%s?interval=%s&range=%s",
		YahooFinanceBaseURL, symbol, interval, yahooRange)

	// Create HTTP request with headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to avoid 403/429 errors
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Yahoo Finance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Yahoo Finance API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var yahooResp YahooChartResponse
	if err := json.Unmarshal(body, &yahooResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check for API errors
	if yahooResp.Chart.Error != nil {
		return nil, fmt.Errorf("Yahoo API error: %s - %s",
			yahooResp.Chart.Error.Code, yahooResp.Chart.Error.Description)
	}

	// Check if we have results
	if len(yahooResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for symbol: %s", symbol)
	}

	result := yahooResp.Chart.Result[0]

	// Check if we have quote data
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data returned for symbol: %s", symbol)
	}

	quote := result.Indicators.Quote[0]
	timestamps := result.Timestamp

	// Build candlestick array
	candles := make([]Candlestick, 0, len(timestamps))
	for i := 0; i < len(timestamps); i++ {
		// Skip if any OHLC value is missing (null in JSON becomes 0)
		if i >= len(quote.Open) || i >= len(quote.High) ||
			i >= len(quote.Low) || i >= len(quote.Close) {
			continue
		}

		// Skip candles with zero/null values
		if quote.Open[i] == 0 && quote.High[i] == 0 && quote.Low[i] == 0 && quote.Close[i] == 0 {
			continue
		}

		// Handle potential nil values in arrays (JSON null becomes 0 in Go)
		openPrice := quote.Open[i]
		highPrice := quote.High[i]
		lowPrice := quote.Low[i]
		closePrice := quote.Close[i]

		// Use previous close for missing open
		if openPrice == 0 && len(candles) > 0 {
			openPrice = candles[len(candles)-1].Close
		}
		if closePrice == 0 && openPrice > 0 {
			closePrice = openPrice
		}
		if highPrice == 0 {
			highPrice = openPrice
		}
		if lowPrice == 0 {
			lowPrice = openPrice
		}

		// Handle volume (forex often has 0 volume)
		volume := 0.0
		if i < len(quote.Volume) {
			volume = float64(quote.Volume[i])
		}

		// Calculate close time based on interval
		closeTimeUnix := timestamps[i]
		switch interval {
		case YahooInterval1m:
			closeTimeUnix += 60
		case YahooInterval2m:
			closeTimeUnix += 120
		case YahooInterval5m:
			closeTimeUnix += 300
		case YahooInterval15m:
			closeTimeUnix += 900
		case YahooInterval30m:
			closeTimeUnix += 1800
		case YahooInterval90m:
			closeTimeUnix += 5400
		case YahooInterval1h:
			closeTimeUnix += 3600
		case YahooInterval1d:
			closeTimeUnix += 86400
		case YahooInterval1wk:
			closeTimeUnix += 604800
		case YahooInterval1mo:
			closeTimeUnix += 2592000 // Approx 30 days
		case YahooInterval3mo:
			closeTimeUnix += 7776000 // Approx 90 days
		}

		candles = append(candles, Candlestick{
			OpenTime:  time.Unix(timestamps[i], 0),
			Open:      openPrice,
			High:      highPrice,
			Low:       lowPrice,
			Close:     closePrice,
			Volume:    volume,
			CloseTime: time.Unix(closeTimeUnix, 0),
		})
	}

	// Limit the number of candles returned
	if len(candles) > limit {
		candles = candles[len(candles)-limit:]
	}

	return candles, nil
}

// ValidateYahooSymbol checks if a symbol exists on Yahoo Finance
func ValidateYahooSymbol(symbol string) (bool, error) {
	url := fmt.Sprintf("%s/%s?interval=1d&range=1d", YahooFinanceBaseURL, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

// GetYahooCurrentPrice fetches the current price of a forex symbol
func GetYahooCurrentPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("%s/%s?interval=1m&range=1d", YahooFinanceBaseURL, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("symbol not found: %s", symbol)
	}

	body, _ := io.ReadAll(resp.Body)

	var yahooResp YahooChartResponse
	if err := json.Unmarshal(body, &yahooResp); err != nil {
		return 0, err
	}

	if len(yahooResp.Chart.Result) > 0 {
		return yahooResp.Chart.Result[0].Meta.RegularMarketPrice, nil
	}

	return 0, fmt.Errorf("no price data for symbol: %s", symbol)
}

// ConvertBinanceToYahooInterval converts Binance interval to Yahoo interval
func ConvertBinanceToYahooInterval(bi BinanceInterval) YahooInterval {
	switch bi {
	case Interval1m:
		return YahooInterval1m
	case Interval5m:
		return YahooInterval5m
	case Interval15m:
		return YahooInterval15m
	case Interval30m:
		return YahooInterval30m
	case Interval1h:
		return YahooInterval1h
	case Interval4h:
		return YahooInterval1h // Yahoo doesn't have 4h, use 1h
	case Interval1d:
		return YahooInterval1d
	case Interval1w:
		return YahooInterval1wk
	default:
		return YahooInterval1h
	}
}

// GetYahooTimeframeName returns human-readable name for Yahoo interval
func GetYahooTimeframeName(interval YahooInterval) string {
	switch interval {
	case YahooInterval1m:
		return "1 Minute"
	case YahooInterval2m:
		return "2 Minutes"
	case YahooInterval5m:
		return "5 Minutes"
	case YahooInterval15m:
		return "15 Minutes"
	case YahooInterval30m:
		return "30 Minutes"
	case YahooInterval90m:
		return "90 Minutes"
	case YahooInterval1h:
		return "1 Hour"
	case YahooInterval1d:
		return "1 Day"
	case YahooInterval1wk:
		return "1 Week"
	case YahooInterval1mo:
		return "1 Month"
	case YahooInterval3mo:
		return "3 Months"
	default:
		return string(interval)
	}
}
