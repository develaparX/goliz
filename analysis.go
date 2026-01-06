package main

import (
	"fmt"
	"strings"
	"time"
)

// CandleDataSummary represents a summarized view of candlestick data
type CandleDataSummary struct {
	Interval     BinanceInterval
	CandleCount  int
	StartTime    time.Time
	EndTime      time.Time
	Open         float64 // First candle open
	High         float64 // Highest high
	Low          float64 // Lowest low
	Close        float64 // Last candle close
	TotalVolume  float64
	AvgVolume    float64
	PriceChange  float64 // Percentage change
	Trend        string  // BULLISH, BEARISH, SIDEWAYS
	MA20         float64
	MA50         float64
	RSI          float64
	Volatility   string // LOW, MEDIUM, HIGH
	LastCandles  []CandleSimple // Last 10 candles for pattern recognition
}

// CandleSimple is a simplified candle for the prompt
type CandleSimple struct {
	Time   string
	O      float64
	H      float64
	L      float64
	C      float64
	Vol    float64
	Change float64 // vs previous close
	Type   string  // BULL, BEAR, DOJI
}

// AnalyzeCandlestickData creates a summary from raw candlestick data
func AnalyzeCandlestickData(candles []Candlestick, interval BinanceInterval) CandleDataSummary {
	if len(candles) == 0 {
		return CandleDataSummary{}
	}

	summary := CandleDataSummary{
		Interval:    interval,
		CandleCount: len(candles),
		StartTime:   candles[0].OpenTime,
		EndTime:     candles[len(candles)-1].CloseTime,
		Open:        candles[0].Open,
		Close:       candles[len(candles)-1].Close,
		High:        candles[0].High,
		Low:         candles[0].Low,
	}

	// Calculate High, Low, Volume
	for _, c := range candles {
		if c.High > summary.High {
			summary.High = c.High
		}
		if c.Low < summary.Low {
			summary.Low = c.Low
		}
		summary.TotalVolume += c.Volume
	}
	summary.AvgVolume = summary.TotalVolume / float64(len(candles))

	// Price change percentage
	if summary.Open > 0 {
		summary.PriceChange = ((summary.Close - summary.Open) / summary.Open) * 100
	}

	// Calculate MAs
	if len(candles) >= 20 {
		sum20 := 0.0
		for i := len(candles) - 20; i < len(candles); i++ {
			sum20 += candles[i].Close
		}
		summary.MA20 = sum20 / 20
	}

	if len(candles) >= 50 {
		sum50 := 0.0
		for i := len(candles) - 50; i < len(candles); i++ {
			sum50 += candles[i].Close
		}
		summary.MA50 = sum50 / 50
	}

	// Calculate RSI (14 period)
	if len(candles) >= 15 {
		gains, losses := 0.0, 0.0
		for i := len(candles) - 14; i < len(candles); i++ {
			change := candles[i].Close - candles[i-1].Close
			if change > 0 {
				gains += change
			} else {
				losses -= change
			}
		}
		avgGain := gains / 14
		avgLoss := losses / 14
		if avgLoss > 0 {
			rs := avgGain / avgLoss
			summary.RSI = 100 - (100 / (1 + rs))
		} else {
			summary.RSI = 100
		}
	}

	// Determine Trend
	if summary.Close > summary.MA20 && summary.MA20 > summary.MA50 {
		summary.Trend = "BULLISH"
	} else if summary.Close < summary.MA20 && summary.MA20 < summary.MA50 {
		summary.Trend = "BEARISH"
	} else {
		summary.Trend = "SIDEWAYS"
	}

	// Calculate Volatility (ATR-based)
	if len(candles) >= 14 {
		atrSum := 0.0
		for i := len(candles) - 14; i < len(candles); i++ {
			atrSum += candles[i].High - candles[i].Low
		}
		atr := atrSum / 14
		atrPercent := (atr / summary.Close) * 100
		if atrPercent > 3 {
			summary.Volatility = "HIGH"
		} else if atrPercent > 1.5 {
			summary.Volatility = "MEDIUM"
		} else {
			summary.Volatility = "LOW"
		}
	}

	// Last 10 candles for pattern recognition
	startIdx := len(candles) - 10
	if startIdx < 0 {
		startIdx = 0
	}
	for i := startIdx; i < len(candles); i++ {
		c := candles[i]
		cs := CandleSimple{
			Time: c.OpenTime.Format("01-02 15:04"),
			O:    c.Open,
			H:    c.High,
			L:    c.Low,
			C:    c.Close,
			Vol:  c.Volume,
		}
		
		// Calculate change from previous
		if i > 0 {
			cs.Change = ((c.Close - candles[i-1].Close) / candles[i-1].Close) * 100
		}
		
		// Determine candle type
		bodySize := c.Close - c.Open
		totalRange := c.High - c.Low
		if totalRange > 0 {
			bodyRatio := bodySize / totalRange
			if bodyRatio > 0.1 {
				cs.Type = "BULL"
			} else if bodyRatio < -0.1 {
				cs.Type = "BEAR"
			} else {
				cs.Type = "DOJI"
			}
		}
		
		summary.LastCandles = append(summary.LastCandles, cs)
	}

	return summary
}

// FormatDataForAI formats multiple timeframe data into a structured prompt
func FormatDataForAI(symbol string, summaries []CandleDataSummary, mode TradingMode) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("=== MULTI-TIMEFRAME DATA ANALYSIS ===\n"))
	sb.WriteString(fmt.Sprintf("Symbol: %s\n", symbol))
	sb.WriteString(fmt.Sprintf("Analysis Mode: %s\n", strings.ToUpper(string(mode))))
	sb.WriteString(fmt.Sprintf("Data Generated: %s UTC\n\n", time.Now().UTC().Format("2006-01-02 15:04:05")))

	for _, s := range summaries {
		sb.WriteString(fmt.Sprintf("--- %s TIMEFRAME ---\n", GetTimeframeName(s.Interval)))
		sb.WriteString(fmt.Sprintf("Period: %s to %s\n", s.StartTime.Format("2006-01-02 15:04"), s.EndTime.Format("2006-01-02 15:04")))
		sb.WriteString(fmt.Sprintf("Candles Analyzed: %d\n", s.CandleCount))
		sb.WriteString(fmt.Sprintf("Open: %.8f | High: %.8f | Low: %.8f | Close: %.8f\n", s.Open, s.High, s.Low, s.Close))
		sb.WriteString(fmt.Sprintf("Price Change: %.2f%%\n", s.PriceChange))
		sb.WriteString(fmt.Sprintf("MA20: %.8f | MA50: %.8f\n", s.MA20, s.MA50))
		sb.WriteString(fmt.Sprintf("RSI(14): %.1f\n", s.RSI))
		sb.WriteString(fmt.Sprintf("Trend: %s | Volatility: %s\n", s.Trend, s.Volatility))
		sb.WriteString(fmt.Sprintf("Avg Volume: %.2f\n", s.AvgVolume))
		
		// Last candles
		sb.WriteString("Last 10 Candles (Time|O|H|L|C|Change|Type):\n")
		for _, c := range s.LastCandles {
			sb.WriteString(fmt.Sprintf("  %s | %.6f | %.6f | %.6f | %.6f | %+.2f%% | %s\n", 
				c.Time, c.O, c.H, c.L, c.C, c.Change, c.Type))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateDataAnalysisPrompt creates a prompt for data-based analysis
func GenerateDataAnalysisPrompt(mode TradingMode, symbol string, dataContext string) string {
	modeDescription := ""
	switch mode {
	case TradingModeScalping:
		modeDescription = `MODE: SCALPING (Entry cepat, SL ketat, target pendek)
- Fokus pada momentum dan liquidity di timeframe kecil
- Cari rejection candle, engulfing pattern, dan break of structure
- Risk:Reward minimal 1:2`
	case TradingModeSwing:
		modeDescription = `MODE: SWING TRADING (Hold beberapa hari sampai minggu)
- Fokus pada trend besar di Weekly/Daily
- Entry di pullback ke area demand/supply
- Risk:Reward minimal 1:3`
	case TradingModeIntraday:
		modeDescription = `MODE: INTRADAY (Trading dalam 1 hari)
- Identifikasi trend dari 4H/1D
- Entry di 15m/1H dengan konfirmasi
- Close posisi sebelum akhir hari`
	}

	return fmt.Sprintf(`Kamu adalah "Antigravity Quant Analyst", AI trading profesional dengan keahlian Smart Money Concept (SMC), Supply Demand, dan Multi-Timeframe Analysis.

%s

DATA MARKET (REAL-TIME dari Binance):
%s

TUGAS ANALISIS:

1. TOP-DOWN ANALYSIS
   - Analisa dari timeframe TERBESAR ke TERKECIL
   - Identifikasi: Trend utama, Key Levels (Support/Resistance), Market Structure (HH/HL atau LH/LL)

2. SMART MONEY ANALYSIS
   - Order Blocks (OB)
   - Fair Value Gaps (FVG) / Imbalance
   - Break of Structure (BOS) / Change of Character (ChoCh)
   - Liquidity zones (Equal highs/lows)

3. ENTRY ANALYSIS
   - Tentukan Entry Point yang optimal
   - Stoploss (behind structure / invalidation level)
   - Take Profit 1, 2, 3 (berdasarkan structure targets)
   - Risk:Reward Ratio

4. PROBABILITY & CONFIDENCE
   - Berikan confidence level (0-100%%)
   - Sebutkan faktor pendukung dan berlawanan

--------------------------------------------------------
OUTPUT FORMAT (STRICT HTML untuk Telegram):

<b>🛸 ANTIGRAVITY PRIME</b>
<code>%s</code> • <code>%s</code>

<blockquote>💡 <i>"[Quote insight singkat tentang setup ini]"</i></blockquote>

<b>📊 MARKET STRUCTURE</b>
Trend: [BULLISH/BEARISH/SIDEWAYS]
Key Support: [level]
Key Resistance: [level]
Market Phase: [Accumulation/Distribution/Markup/Markdown]

<b>💎 SIGNAL</b>
<pre><code class="language-diff">
[Gunakan + untuk HIJAU (positif)]
[Gunakan - untuk MERAH (negatif)]

+ ACTION:    [BUY/SELL/WAIT]
+ ENTRY:     [harga entry]
- STOPLOSS:  [harga SL]
+ TP1:       [target 1]
+ TP2:       [target 2]
+ TP3:       [target 3]
+ R:R RATIO: [rasio]
</code></pre>

<b>📈 CONFIDENCE: [XX]%%</b>
✅ Bullish Factors: [list]
⚠️ Bearish Factors: [list]

<b>📝 ANALYSIS</b>
[Jelaskan reasoning secara ringkas - max 3 paragraf]

<b>⚠️ RISK MANAGEMENT</b>
- Position Size: Max [X]%% dari portfolio
- [Tips risk management spesifik]

---
<i>Generated by Antigravity AI • Data-Based Analysis</i>
`, modeDescription, dataContext, symbol, getTradingModeName(mode), symbol, string(mode))
}

// FetchMultiTimeframeData fetches data for all timeframes without generating images
func FetchMultiTimeframeData(symbol string, mode TradingMode, candleLimit int) ([]CandleDataSummary, error) {
	timeframes := GetTimeframesForMode(mode)
	summaries := make([]CandleDataSummary, 0, len(timeframes))

	for _, tf := range timeframes {
		candles, err := FetchCandlesticks(symbol, tf, candleLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", tf, err)
		}
		
		summary := AnalyzeCandlestickData(candles, tf)
		summaries = append(summaries, summary)
	}

	return summaries, nil
}
