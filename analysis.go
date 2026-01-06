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

// GenerateDataAnalysisPrompt creates a prompt for data-based analysis (matching manual flow)
func GenerateDataAnalysisPrompt(mode TradingMode, symbol string, dataContext string) string {
	baseRole := ""
	strategy := ""

	switch mode {
	case TradingModeScalping:
		baseRole = `ROLE: Kamu adalah "Antigravity Scalper", trader agresif spesialis timeframe kecil (M5, M15). Kamu mencari momentum cepat, liquidity grabs, dan rejection tajam.`
		strategy = `METODE SCALPING (FAST EXECUTION):
- Fokus cari: Liquidity Sweep (Pengambilan Stoploss retail) lalu Reversal.
- Rejection Candle Wajib Jelas (Pinbar/Engulfing).
- Risk Reward Ratio minimal 1:2.
- Stoploss harus KETAT (Tight).`
	case TradingModeSwing:
		baseRole = `ROLE: Kamu adalah "Antigravity Swing Master", trader sabar yang menunggu setup sempurna di timeframe besar (1D, 1W).`
		strategy = `METODE SWING TRADING:
- Fokus pada trend besar dan hold beberapa hari sampai minggu.
- Entry di pullback ke area demand/supply yang kuat.
- Risk Reward Ratio minimal 1:3.`
	case TradingModeIntraday:
		baseRole = `ROLE: Kamu adalah "Antigravity Quant Analyst", AI trading intraday yang mencari setup High Probability (Win Rate > 80%).`
		strategy = `METODE INTRADAY:
- Gunakan Smart Money Concept (SMC) + Supply Demand.
- Validasi Market Structure (BOS/ChoCh).
- Close semua posisi sebelum akhir hari.`
	default:
		baseRole = `ROLE: Kamu adalah "Antigravity Quant Analyst", AI trading profesional dengan keahlian SMC dan Multi-Timeframe Analysis.`
		strategy = `METODE STANDARD:
- Gunakan Smart Money Concept (SMC) + Supply Demand.
- Validasi Market Structure (BOS/ChoCh).
- Cari konfirmasi Divergence atau Pola Chart Pattern.`
	}

	return fmt.Sprintf(`%s

DATA MARKET REAL-TIME (Binance):
%s

%s

TUGAS ANALISIS TOP-DOWN:

LANGKAH 1: EXTERNAL DATA VALIDATION
- Cari sentimen pasar %s hari ini menggunakan Google Search.

LANGKAH 2: MULTI-TIMEFRAME ANALYSIS
- Analisa dari timeframe TERBESAR ke TERKECIL
- Identifikasi: Trend utama di HTF (Higher Time Frame)
- Cari entry presisi di LTF (Lower Time Frame)
- Pastikan confluence antara HTF dan LTF

LANGKAH 3: SMART MONEY ANALYSIS
- Order Blocks (OB) - zona akumulasi institusional
- Fair Value Gaps (FVG) / Imbalance
- Break of Structure (BOS) / Change of Character (ChoCh)
- Liquidity zones (Equal highs/lows yang akan di-sweep)

LANGKAH 4: ENTRY SETUP
- Entry Point yang optimal (harga spesifik)
- Stoploss (behind structure / invalidation level)
- Take Profit 1, 2, 3 (berdasarkan structure targets)
- Risk:Reward Ratio

--------------------------------------------------------
CRITICAL RULE:
1. GUNAKAN FORMAT HTML (Telegram Compatible).
2. Escape karakter < > & di dalam teks biasa.
3. GUNAKAN Code Block "diff" untuk warna merah/hijau.
4. BERIKAN HARGA SPESIFIK untuk Entry, SL, TP (bukan range).
--------------------------------------------------------

OUTPUT FORMAT (STRICT HTML):

<b>🛸 ANTIGRAVITY PRIME</b>
<code>%s</code> • <code>%s</code>

<b>⚙️ STRATEGY MODE: %s</b>

<blockquote>💡 <i>"[Quote insight singkat tentang setup ini]"</i></blockquote>

<b>📊 MARKET STRUCTURE</b>
HTF Trend (1D/4H): <b>[BULLISH/BEARISH]</b>
LTF Trend (1H/15m): <b>[BULLISH/BEARISH]</b>
Key Support: [level harga]
Key Resistance: [level harga]
Volatility: [Low/Med/High]

<b>💎 SIGNAL CARD</b>
<pre><code class="language-diff">
[Gunakan tanda + untuk HIJAU (Buy/TP/Positif)]
[Gunakan tanda - untuk MERAH (Sell/SL/Negatif)]

+ ACTION:  [BUY/SELL/WAIT]
+ ENTRY:   [harga entry spesifik]
- SL:      [harga stoploss]
+ TP 1:    [target 1]
+ TP 2:    [target 2]
+ TP 3:    [target 3]
+ R:R:     [rasio risk reward]
</code></pre>

<b>📈 CONFIDENCE: [XX]%%</b>

<b>📝 ANALYSIS BRIEF</b>
[Jelaskan alasan teknikal secara padat - max 2 paragraf]

<b>⚠️ RISK NOTES</b>
- Position Size: Max [X]%% dari portfolio
- [Kondisi invalidasi setup]

---
<i>Generated by Antigravity AI • Data-Based Analysis</i>
`, baseRole, dataContext, strategy, symbol, symbol, getTradingModeName(mode), getTradingModeName(mode))
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
