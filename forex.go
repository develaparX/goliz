package main

import (
	"fmt"
	"strings"
)

// ForexPair represents a forex currency pair
type ForexPair struct {
	Symbol      string // Yahoo Finance symbol (e.g., "EURUSD=X")
	DisplayName string // Human readable name (e.g., "EUR/USD")
	BaseCurr    string // Base currency (e.g., "EUR")
	QuoteCurr   string // Quote currency (e.g., "USD")
}

// Common forex pairs with Yahoo Finance symbols
var CommonForexPairs = map[string]ForexPair{
	// Major Pairs
	"EURUSD": {Symbol: "EURUSD=X", DisplayName: "EUR/USD", BaseCurr: "EUR", QuoteCurr: "USD"},
	"GBPUSD": {Symbol: "GBPUSD=X", DisplayName: "GBP/USD", BaseCurr: "GBP", QuoteCurr: "USD"},
	"USDJPY": {Symbol: "USDJPY=X", DisplayName: "USD/JPY", BaseCurr: "USD", QuoteCurr: "JPY"},
	"USDCHF": {Symbol: "USDCHF=X", DisplayName: "USD/CHF", BaseCurr: "USD", QuoteCurr: "CHF"},
	"AUDUSD": {Symbol: "AUDUSD=X", DisplayName: "AUD/USD", BaseCurr: "AUD", QuoteCurr: "USD"},
	"USDCAD": {Symbol: "USDCAD=X", DisplayName: "USD/CAD", BaseCurr: "USD", QuoteCurr: "CAD"},
	"NZDUSD": {Symbol: "NZDUSD=X", DisplayName: "NZD/USD", BaseCurr: "NZD", QuoteCurr: "USD"},

	// Cross Pairs (EUR crosses)
	"EURGBP": {Symbol: "EURGBP=X", DisplayName: "EUR/GBP", BaseCurr: "EUR", QuoteCurr: "GBP"},
	"EURJPY": {Symbol: "EURJPY=X", DisplayName: "EUR/JPY", BaseCurr: "EUR", QuoteCurr: "JPY"},
	"EURCHF": {Symbol: "EURCHF=X", DisplayName: "EUR/CHF", BaseCurr: "EUR", QuoteCurr: "CHF"},
	"EURAUD": {Symbol: "EURAUD=X", DisplayName: "EUR/AUD", BaseCurr: "EUR", QuoteCurr: "AUD"},
	"EURCAD": {Symbol: "EURCAD=X", DisplayName: "EUR/CAD", BaseCurr: "EUR", QuoteCurr: "CAD"},
	"EURNZD": {Symbol: "EURNZD=X", DisplayName: "EUR/NZD", BaseCurr: "EUR", QuoteCurr: "NZD"},

	// Cross Pairs (GBP crosses)
	"GBPJPY": {Symbol: "GBPJPY=X", DisplayName: "GBP/JPY", BaseCurr: "GBP", QuoteCurr: "JPY"},
	"GBPCHF": {Symbol: "GBPCHF=X", DisplayName: "GBP/CHF", BaseCurr: "GBP", QuoteCurr: "CHF"},
	"GBPAUD": {Symbol: "GBPAUD=X", DisplayName: "GBP/AUD", BaseCurr: "GBP", QuoteCurr: "AUD"},
	"GBPCAD": {Symbol: "GBPCAD=X", DisplayName: "GBP/CAD", BaseCurr: "GBP", QuoteCurr: "CAD"},
	"GBPNZD": {Symbol: "GBPNZD=X", DisplayName: "GBP/NZD", BaseCurr: "GBP", QuoteCurr: "NZD"},

	// Cross Pairs (JPY crosses)
	"AUDJPY": {Symbol: "AUDJPY=X", DisplayName: "AUD/JPY", BaseCurr: "AUD", QuoteCurr: "JPY"},
	"CADJPY": {Symbol: "CADJPY=X", DisplayName: "CAD/JPY", BaseCurr: "CAD", QuoteCurr: "JPY"},
	"CHFJPY": {Symbol: "CHFJPY=X", DisplayName: "CHF/JPY", BaseCurr: "CHF", QuoteCurr: "JPY"},
	"NZDJPY": {Symbol: "NZDJPY=X", DisplayName: "NZD/JPY", BaseCurr: "NZD", QuoteCurr: "JPY"},

	// Other crosses
	"AUDCAD": {Symbol: "AUDCAD=X", DisplayName: "AUD/CAD", BaseCurr: "AUD", QuoteCurr: "CAD"},
	"AUDCHF": {Symbol: "AUDCHF=X", DisplayName: "AUD/CHF", BaseCurr: "AUD", QuoteCurr: "CHF"},
	"AUDNZD": {Symbol: "AUDNZD=X", DisplayName: "AUD/NZD", BaseCurr: "AUD", QuoteCurr: "NZD"},
	"CADCHF": {Symbol: "CADCHF=X", DisplayName: "CAD/CHF", BaseCurr: "CAD", QuoteCurr: "CHF"},
	"NZDCAD": {Symbol: "NZDCAD=X", DisplayName: "NZD/CAD", BaseCurr: "NZD", QuoteCurr: "CAD"},
	"NZDCHF": {Symbol: "NZDCHF=X", DisplayName: "NZD/CHF", BaseCurr: "NZD", QuoteCurr: "CHF"},

	// Exotic Pairs (popular ones)
	"USDSGD": {Symbol: "USDSGD=X", DisplayName: "USD/SGD", BaseCurr: "USD", QuoteCurr: "SGD"},
	"USDHKD": {Symbol: "USDHKD=X", DisplayName: "USD/HKD", BaseCurr: "USD", QuoteCurr: "HKD"},
	"USDZAR": {Symbol: "USDZAR=X", DisplayName: "USD/ZAR", BaseCurr: "USD", QuoteCurr: "ZAR"},
	"USDMXN": {Symbol: "USDMXN=X", DisplayName: "USD/MXN", BaseCurr: "USD", QuoteCurr: "MXN"},
	"USDTRY": {Symbol: "USDTRY=X", DisplayName: "USD/TRY", BaseCurr: "USD", QuoteCurr: "TRY"},
	"USDSEK": {Symbol: "USDSEK=X", DisplayName: "USD/SEK", BaseCurr: "USD", QuoteCurr: "SEK"},
	"USDNOK": {Symbol: "USDNOK=X", DisplayName: "USD/NOK", BaseCurr: "USD", QuoteCurr: "NOK"},
	"USDIDR": {Symbol: "USDIDR=X", DisplayName: "USD/IDR", BaseCurr: "USD", QuoteCurr: "IDR"},

	// Gold & Silver (Commodities traded as forex)
	"XAUUSD": {Symbol: "GC=F", DisplayName: "XAU/USD (Gold)", BaseCurr: "XAU", QuoteCurr: "USD"},
	"XAGUSD": {Symbol: "SI=F", DisplayName: "XAG/USD (Silver)", BaseCurr: "XAG", QuoteCurr: "USD"},
}

// NormalizeForexSymbol normalizes user input to Yahoo Finance symbol format
// Input can be: "EURUSD", "EUR/USD", "eurusd", "EURUSD=X"
// Output: Yahoo Finance symbol like "EURUSD=X"
func NormalizeForexSymbol(input string) (string, string, error) {
	// Clean input
	input = strings.ToUpper(strings.TrimSpace(input))
	input = strings.ReplaceAll(input, "/", "")
	input = strings.ReplaceAll(input, " ", "")

	// Remove =X suffix if present (user might input it)
	input = strings.TrimSuffix(input, "=X")
	input = strings.TrimSuffix(input, "=F")

	// Check if it's in our common pairs
	if pair, ok := CommonForexPairs[input]; ok {
		return pair.Symbol, pair.DisplayName, nil
	}

	// If not found, try to construct Yahoo symbol
	// Assume it's a valid forex pair and add =X
	if len(input) == 6 {
		return input + "=X", input[:3] + "/" + input[3:], nil
	}

	return "", "", fmt.Errorf("invalid forex symbol format: %s (use format like EURUSD or EUR/USD)", input)
}

// GetForexTimeframesForMode returns appropriate timeframes for forex trading
// Same as crypto but uses Yahoo-compatible intervals
func GetForexTimeframesForMode(mode TradingMode) []YahooInterval {
	switch mode {
	case TradingModeScalping:
		// Scalping: Using all available timeframes for maximum context
		return []YahooInterval{
			YahooInterval1m, YahooInterval2m, YahooInterval5m, YahooInterval15m, YahooInterval30m,
			YahooInterval1h, YahooInterval90m, YahooInterval1d, YahooInterval1wk,
			YahooInterval1mo, YahooInterval3mo,
		}
	case TradingModeSwing:
		// Swing: Using all available timeframes for maximum context
		return []YahooInterval{
			YahooInterval1m, YahooInterval2m, YahooInterval5m, YahooInterval15m, YahooInterval30m,
			YahooInterval1h, YahooInterval90m, YahooInterval1d, YahooInterval1wk,
			YahooInterval1mo, YahooInterval3mo,
		}
	case TradingModeIntraday:
		// Intraday: Using all available timeframes for maximum context
		return []YahooInterval{
			YahooInterval1m, YahooInterval2m, YahooInterval5m, YahooInterval15m, YahooInterval30m,
			YahooInterval1h, YahooInterval90m, YahooInterval1d, YahooInterval1wk,
			YahooInterval1mo, YahooInterval3mo,
		}
	default:
		// Default: Using all available timeframes for maximum context
		return []YahooInterval{
			YahooInterval1m, YahooInterval2m, YahooInterval5m, YahooInterval15m, YahooInterval30m,
			YahooInterval1h, YahooInterval90m, YahooInterval1d, YahooInterval1wk,
			YahooInterval1mo, YahooInterval3mo,
		}
	}
}

// FetchForexMultiTimeframeData fetches forex data for all timeframes
func FetchForexMultiTimeframeData(symbol string, mode TradingMode, candleLimit int) ([]CandleDataSummary, error) {
	timeframes := GetForexTimeframesForMode(mode)
	summaries := make([]CandleDataSummary, 0, len(timeframes))

	for _, tf := range timeframes {
		candles, err := FetchYahooCandlesticks(symbol, tf, candleLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", tf, err)
		}

		// Convert Yahoo interval to Binance interval for compatibility with existing analysis
		binanceInterval := ConvertYahooToBinanceInterval(tf)
		summary := AnalyzeCandlestickData(candles, binanceInterval)
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// ConvertYahooToBinanceInterval converts Yahoo interval to Binance interval for display
func ConvertYahooToBinanceInterval(yi YahooInterval) BinanceInterval {
	switch yi {
	case YahooInterval1m:
		return Interval1m
	case YahooInterval2m:
		return BinanceInterval("2m")
	case YahooInterval5m:
		return Interval5m
	case YahooInterval15m:
		return Interval15m
	case YahooInterval30m:
		return Interval30m
	case YahooInterval90m:
		return BinanceInterval("90m")
	case YahooInterval1h:
		return Interval1h
	case YahooInterval1d:
		return Interval1d
	case YahooInterval1wk:
		return Interval1w
	case YahooInterval1mo:
		return BinanceInterval("1M")
	case YahooInterval3mo:
		return BinanceInterval("3M")
	default:
		return Interval1h
	}
}

// FormatForexDataForAI formats forex multi-timeframe data for AI analysis
func FormatForexDataForAI(symbol, displayName string, summaries []CandleDataSummary, mode TradingMode) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== FOREX MULTI-TIMEFRAME DATA ANALYSIS ===\n"))
	sb.WriteString(fmt.Sprintf("Symbol: %s (%s)\n", displayName, symbol))
	sb.WriteString(fmt.Sprintf("Market: FOREX\n"))
	sb.WriteString(fmt.Sprintf("Analysis Mode: %s\n", strings.ToUpper(string(mode))))
	sb.WriteString(fmt.Sprintf("Data Source: Yahoo Finance\n\n"))

	for _, s := range summaries {
		sb.WriteString(fmt.Sprintf("--- %s TIMEFRAME ---\n", GetTimeframeName(s.Interval)))
		sb.WriteString(fmt.Sprintf("Period: %s to %s\n", s.StartTime.Format("2006-01-02 15:04"), s.EndTime.Format("2006-01-02 15:04")))
		sb.WriteString(fmt.Sprintf("Candles Analyzed: %d\n", s.CandleCount))
		sb.WriteString(fmt.Sprintf("Open: %.5f | High: %.5f | Low: %.5f | Close: %.5f\n", s.Open, s.High, s.Low, s.Close))
		sb.WriteString(fmt.Sprintf("Price Change: %.2f%%\n", s.PriceChange))
		sb.WriteString(fmt.Sprintf("MA20: %.5f | MA50: %.5f\n", s.MA20, s.MA50))
		sb.WriteString(fmt.Sprintf("RSI(14): %.1f\n", s.RSI))
		sb.WriteString(fmt.Sprintf("Trend: %s | Volatility: %s\n", s.Trend, s.Volatility))

		// Last candles
		sb.WriteString("Last 10 Candles (Time|O|H|L|C|Change|Type):\n")
		for _, c := range s.LastCandles {
			sb.WriteString(fmt.Sprintf("  %s | %.5f | %.5f | %.5f | %.5f | %+.2f%% | %s\n",
				c.Time, c.O, c.H, c.L, c.C, c.Change, c.Type))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateForexAnalysisPrompt creates a specialized prompt for forex analysis
func GenerateForexAnalysisPrompt(mode TradingMode, symbol, displayName, dataContext string) string {
	baseRole := ""
	strategy := ""

	switch mode {
	case TradingModeScalping:
		baseRole = `ROLE: Kamu adalah "Antigravity FX Scalper", trader forex agresif spesialis timeframe kecil (M5, M15). Kamu mencari momentum cepat di sesi London dan New York.`
		strategy = `METODE FOREX SCALPING:
- Fokus pada sesi overlap London-NY (14:00-22:00 WIB) untuk volatilitas optimal.
- Perhatikan news event ekonomi (NFP, FOMC, ECB) yang bisa menyebabkan spike.
- Entry saat liquidity sweep di level psikologis (00, 50).
- Risk Reward Ratio minimal 1:2 dengan tight stoploss.`
	case TradingModeSwing:
		baseRole = `ROLE: Kamu adalah "Antigravity FX Swing Master", trader forex sabar yang menunggu setup daily/weekly.`
		strategy = `METODE FOREX SWING:
- Analisa fundamental: interest rate differential, ekonomi makro.
- Entry di pullback ke area demand/supply pada chart daily.
- Hold posisi beberapa hari sampai minggu.
- Risk Reward Ratio minimal 1:3.`
	case TradingModeIntraday:
		baseRole = `ROLE: Kamu adalah "Antigravity FX Intraday Pro", trader forex intraday yang close semua posisi sebelum market tutup.`
		strategy = `METODE FOREX INTRADAY:
- Trade saat sesi aktif (Asia, London, New York).
- Gunakan SMC untuk identifikasi order blocks dan FVG.
- Close semua posisi sebelum swap/rollover (05:00 WIB).
- Perhatikan spread dan likuiditas.`
	default:
		baseRole = `ROLE: Kamu adalah "Antigravity FX Analyst", AI trading forex profesional dengan keahlian analisis teknikal dan fundamental.`
		strategy = `METODE FOREX STANDARD:
- Gunakan Smart Money Concept (SMC) + Supply Demand.
- Validasi dengan analisa fundamental (news, economic calendar).
- Cari confluence antara teknikal dan fundamental.`
	}

	return fmt.Sprintf(`%s

DATA MARKET REAL-TIME (Yahoo Finance):
%s

%s

CONTEXT FOREX:
- Symbol: %s (%s)
- Market Type: Foreign Exchange (FOREX)
- Trading Hours: 24/5 (Minggu 22:00 - Jumat 22:00 GMT)
- Spread: Variable tergantung sesi dan likuiditas

TUGAS ANALISIS TOP-DOWN:

LANGKAH 1: EXTERNAL DATA VALIDATION
- Cari sentimen pasar forex untuk %s hari ini menggunakan Google Search.
- Cek calendar ekonomi untuk news yang akan rilis.

LANGKAH 2: MULTI-TIMEFRAME ANALYSIS (FULL SPECTRUM)
- Lakukan analisa menyeluruh mulai dari MACRO (3M, 1M, 1W) untuk melihat Big Picture.
- Identifikasi Trend Major & Key Levels di D1, 1H, 90m.
- Cari struktur entry & momentum presisi di timeframes kecil (30m, 15m, 5m, 2m, 1m).
- Validasi sinyal hanya jika ada CONFLUENCE di 3 timeframe berbeda (misal: 1M + 1H + 5m).

LANGKAH 3: SMART MONEY ANALYSIS
- Order Blocks (OB) di level psikologis (00, 50, 20, 80)
- Fair Value Gaps (FVG) / Imbalance
- Break of Structure (BOS) / Change of Character (ChoCh)
- Liquidity zones (Equal highs/lows)

LANGKAH 4: ENTRY SETUP
- Entry Point yang optimal (harga spesifik dengan 5 desimal untuk forex)
- Stoploss (behind structure / invalidation level)
- Take Profit 1, 2, 3 (berdasarkan structure targets)
- Risk:Reward Ratio

--------------------------------------------------------
CRITICAL RULE:
1. GUNAKAN FORMAT HTML (Telegram Compatible).
2. Escape karakter < > & di dalam teks biasa.
3. GUNAKAN Code Block "diff" untuk warna merah/hijau.
4. BERIKAN HARGA SPESIFIK untuk Entry, SL, TP (dengan 5 desimal untuk forex).
5. PERHATIKAN PIP VALUE dan SPREAD dalam analisa.
6. RISK REWARD RATIO MINIMAL 1(SL) : 2(TP) ADALAH WAJIB. JIKA TIDAK TERPENUHI, ACTION = WAIT.
7. JANGAN MEMAKSAKAN SIGNAL JIKA SETUP BELUM MATANG (ACTION = WAIT).
--------------------------------------------------------

OUTPUT FORMAT (STRICT HTML):

<b>🛸 ANTIGRAVITY FX PRIME</b>
<code>%s</code> • <code>FOREX</code>

<b>⚙️ STRATEGY MODE: %s</b>

<blockquote>💡 <i>"[Quote insight singkat tentang setup ini]"</i></blockquote>

<b>📊 MARKET STRUCTURE</b>
HTF Trend (3M/1M/1W/1D): <b>[BULLISH/BEARISH]</b>
LTF Trend (1H/15m/1m): <b>[BULLISH/BEARISH]</b>
Key Support: [level harga]
Key Resistance: [level harga]
Volatility: [Low/Med/High]
Active Session: [Asia/London/NY]

<b>💎 SIGNAL CARD</b>
<pre><code class="language-diff">
[Gunakan tanda + untuk HIJAU (Buy/TP/Positif)]
[Gunakan tanda - untuk MERAH (Sell/SL/Negatif)]

+ ACTION:  [BUY/SELL/WAIT]
+ ENTRY:   [harga entry 5 desimal]
- SL:      [harga stoploss]
+ TP 1:    [target 1]
+ TP 2:    [target 2]
+ TP 3:    [target 3]
+ R:R:     [rasio risk reward]
+ PIPS:    [estimasi profit dalam pips]
</code></pre>

<b>📈 CONFIDENCE: [XX]%%</b>

<b>📝 ANALYSIS BRIEF</b>
[Jelaskan alasan teknikal secara padat - max 2 paragraf]

<b>⚠️ RISK NOTES</b>
- Position Size: Max [X]%% dari portfolio
- Spread consideration: [spread normal/wide]
- [Kondisi invalidasi setup]
- [News/Event yang perlu diwaspadai]

---
<i>Generated by Antigravity AI • FOREX Analysis • Yahoo Finance Data</i>
`, baseRole, dataContext, strategy, displayName, symbol, displayName, displayName, getTradingModeName(mode))
}
