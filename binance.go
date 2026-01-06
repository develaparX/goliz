package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Binance API base URLs
const (
	BinanceGlobalBaseURL = "https://api.binance.com"
	BinanceUSBaseURL     = "https://api.binance.us"
)

// binanceBaseURLs is the list of Binance API endpoints to try (with fallback)
var binanceBaseURLs = []string{
	BinanceGlobalBaseURL,
	BinanceUSBaseURL,
}

// Candlestick represents OHLCV data
type Candlestick struct {
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime time.Time
}

// BinanceInterval represents Binance kline intervals
type BinanceInterval string

const (
	Interval1m  BinanceInterval = "1m"
	Interval5m  BinanceInterval = "5m"
	Interval15m BinanceInterval = "15m"
	Interval30m BinanceInterval = "30m"
	Interval1h  BinanceInterval = "1h"
	Interval4h  BinanceInterval = "4h"
	Interval1d  BinanceInterval = "1d"
	Interval1w  BinanceInterval = "1w"
)

// TradingMode defines the trading style
type TradingMode string

const (
	TradingModeScalping TradingMode = "scalping"
	TradingModeSwing    TradingMode = "swing"
	TradingModeIntraday TradingMode = "intraday"
)

// getTradingModeName returns human-readable name for trading mode
func getTradingModeName(mode TradingMode) string {
	switch mode {
	case TradingModeScalping:
		return "SCALPING CAFE"
	case TradingModeSwing:
		return "SWING MASTER"
	case TradingModeIntraday:
		return "INTRADAY PRO"
	default:
		return "STANDARD"
	}
}

// GetTimeframesForMode returns the appropriate timeframes for each trading mode
// Top-Down Analysis: 5m, 15m, 1H, 4H, 1D, 1W
func GetTimeframesForMode(mode TradingMode) []BinanceInterval {
	switch mode {
	case TradingModeScalping:
		// Scalping: fokus timeframe kecil tapi tetap ada konteks besar
		// 5m, 15m, 1H, 4H, 1D (tanpa 1W karena scalping tidak perlu weekly)
		return []BinanceInterval{Interval5m, Interval15m, Interval1h, Interval4h, Interval1d}
	case TradingModeSwing:
		// Swing: full top-down analysis 5m, 15m, 1H, 4H, 1D, 1W
		return []BinanceInterval{Interval5m, Interval15m, Interval1h, Interval4h, Interval1d, Interval1w}
	case TradingModeIntraday:
		// Intraday: full top-down analysis 5m, 15m, 1H, 4H, 1D, 1W
		return []BinanceInterval{Interval5m, Interval15m, Interval1h, Interval4h, Interval1d, Interval1w}
	default:
		// Default: full top-down 5m, 15m, 1H, 4H, 1D, 1W
		return []BinanceInterval{Interval5m, Interval15m, Interval1h, Interval4h, Interval1d, Interval1w}
	}
}

// GetTimeframeName returns human-readable name for interval
func GetTimeframeName(interval BinanceInterval) string {
	switch interval {
	case Interval1m:
		return "1 Minute"
	case Interval5m:
		return "5 Minutes"
	case Interval15m:
		return "15 Minutes"
	case Interval30m:
		return "30 Minutes"
	case Interval1h:
		return "1 Hour"
	case Interval4h:
		return "4 Hours"
	case Interval1d:
		return "1 Day"
	case Interval1w:
		return "1 Week"
	default:
		return string(interval)
	}
}

// FetchCandlesticks fetches OHLCV data from Binance API with fallback to Binance US
// symbol: e.g., "BTCUSDT"
// interval: e.g., "1h"
// limit: number of candles (max 1000)
func FetchCandlesticks(symbol string, interval BinanceInterval, limit int) ([]Candlestick, error) {
	if limit > 1000 {
		limit = 1000
	}
	if limit < 1 {
		limit = 200
	}

	var lastErr error
	var body []byte

	// Try each Binance endpoint (global first, then US as fallback)
	for _, baseURL := range binanceBaseURLs {
		url := fmt.Sprintf(
			"%s/api/v3/klines?symbol=%s&interval=%s&limit=%d",
			baseURL, symbol, interval, limit,
		)

		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch from %s: %w", baseURL, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("binance API error from %s (status %d): %s", baseURL, resp.StatusCode, string(respBody))
			continue
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response from %s: %w", baseURL, err)
			continue
		}

		// Success! Break out of the loop
		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all Binance endpoints failed: %w", lastErr)
	}

	// Binance returns array of arrays
	// [OpenTime, Open, High, Low, Close, Volume, CloseTime, QuoteVolume, Trades, TakerBuyBase, TakerBuyQuote, Ignore]
	var rawData [][]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	candles := make([]Candlestick, 0, len(rawData))
	for _, item := range rawData {
		if len(item) < 7 {
			continue
		}

		openTime := int64(item[0].(float64))
		closeTime := int64(item[6].(float64))

		open, _ := strconv.ParseFloat(item[1].(string), 64)
		high, _ := strconv.ParseFloat(item[2].(string), 64)
		low, _ := strconv.ParseFloat(item[3].(string), 64)
		close, _ := strconv.ParseFloat(item[4].(string), 64)
		volume, _ := strconv.ParseFloat(item[5].(string), 64)

		candles = append(candles, Candlestick{
			OpenTime:  time.UnixMilli(openTime),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			CloseTime: time.UnixMilli(closeTime),
		})
	}

	return candles, nil
}

// ValidateSymbol checks if a symbol exists on Binance (with US fallback)
func ValidateSymbol(symbol string) (bool, error) {
	for _, baseURL := range binanceBaseURLs {
		url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", baseURL, symbol)

		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true, nil
		}
	}

	return false, nil
}

// GetCurrentPrice fetches the current price of a symbol (with US fallback)
func GetCurrentPrice(symbol string) (float64, error) {
	var lastErr error

	for _, baseURL := range binanceBaseURLs {
		url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", baseURL, symbol)

		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("symbol not found on %s", baseURL)
			continue
		}

		var result struct {
			Price string `json:"price"`
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			continue
		}

		price, _ := strconv.ParseFloat(result.Price, 64)
		return price, nil
	}

	if lastErr != nil {
		return 0, fmt.Errorf("all Binance endpoints failed: %w", lastErr)
	}
	return 0, fmt.Errorf("symbol not found")
}
