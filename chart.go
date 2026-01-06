package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// ChartConfig holds chart rendering configuration
type ChartConfig struct {
	Width       int
	Height      int
	Padding     int
	CandleWidth int
	CandleGap   int
	ShowVolume  bool
	ShowMA      bool
	MAperiods   []int
	DarkMode    bool
}

// DefaultChartConfig returns a sensible default configuration
func DefaultChartConfig() ChartConfig {
	return ChartConfig{
		Width:       1200,
		Height:      600,
		Padding:     60,
		CandleWidth: 4,
		CandleGap:   2,
		ShowVolume:  true,
		ShowMA:      true,
		MAperiods:   []int{20, 50},
		DarkMode:    true,
	}
}

// Color palette
var (
	colorBullish   = color.RGBA{R: 38, G: 166, B: 91, A: 255}  // Green
	colorBearish   = color.RGBA{R: 231, G: 76, B: 60, A: 255}  // Red
	colorBgDark    = color.RGBA{R: 21, G: 25, B: 31, A: 255}   // Dark background
	colorBgLight   = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	colorGridDark  = color.RGBA{R: 42, G: 46, B: 57, A: 255}   // Grid lines
	colorGridLight = color.RGBA{R: 230, G: 230, B: 230, A: 255}
	colorTextDark  = color.RGBA{R: 180, G: 180, B: 180, A: 255}
	colorTextLight = color.RGBA{R: 60, G: 60, B: 60, A: 255}
	colorMA20      = color.RGBA{R: 255, G: 193, B: 7, A: 255}  // Yellow
	colorMA50      = color.RGBA{R: 156, G: 39, B: 176, A: 255} // Purple
	colorVolume    = color.RGBA{R: 100, G: 149, B: 237, A: 128} // Cornflower blue with transparency
)

// GenerateCandlestickChart creates a PNG candlestick chart from OHLCV data
func GenerateCandlestickChart(candles []Candlestick, symbol string, interval BinanceInterval, config ChartConfig) ([]byte, error) {
	if len(candles) == 0 {
		return nil, fmt.Errorf("no candle data to render")
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Fill background
	bgColor := colorBgDark
	if !config.DarkMode {
		bgColor = colorBgLight
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Calculate price range
	minPrice, maxPrice := candles[0].Low, candles[0].High
	maxVolume := candles[0].Volume
	for _, c := range candles {
		if c.Low < minPrice {
			minPrice = c.Low
		}
		if c.High > maxPrice {
			maxPrice = c.High
		}
		if c.Volume > maxVolume {
			maxVolume = c.Volume
		}
	}

	// Add padding to price range
	priceRange := maxPrice - minPrice
	minPrice -= priceRange * 0.05
	maxPrice += priceRange * 0.05
	priceRange = maxPrice - minPrice

	// Calculate chart area
	chartLeft := config.Padding
	chartRight := config.Width - config.Padding
	chartTop := config.Padding
	chartBottom := config.Height - config.Padding

	if config.ShowVolume {
		chartBottom = int(float64(config.Height-config.Padding) * 0.75)
	}

	chartWidth := chartRight - chartLeft
	chartHeight := chartBottom - chartTop

	// Volume area
	volumeTop := chartBottom + 10
	volumeBottom := config.Height - config.Padding
	volumeHeight := volumeBottom - volumeTop

	// Draw grid lines
	gridColor := colorGridDark
	if !config.DarkMode {
		gridColor = colorGridLight
	}
	drawHorizontalGridLines(img, chartLeft, chartRight, chartTop, chartBottom, 5, gridColor)

	// Calculate candle positions
	totalCandleWidth := config.CandleWidth + config.CandleGap
	maxCandles := chartWidth / totalCandleWidth
	if len(candles) > maxCandles {
		candles = candles[len(candles)-maxCandles:]
	}

	// Draw candles
	for i, c := range candles {
		x := chartLeft + i*totalCandleWidth + config.CandleGap/2

		// Calculate Y positions
		openY := chartTop + int((maxPrice-c.Open)/priceRange*float64(chartHeight))
		closeY := chartTop + int((maxPrice-c.Close)/priceRange*float64(chartHeight))
		highY := chartTop + int((maxPrice-c.High)/priceRange*float64(chartHeight))
		lowY := chartTop + int((maxPrice-c.Low)/priceRange*float64(chartHeight))

		// Determine color
		candleColor := colorBullish
		if c.Close < c.Open {
			candleColor = colorBearish
		}

		// Draw wick (high-low line)
		wickX := x + config.CandleWidth/2
		drawLine(img, wickX, highY, wickX, lowY, candleColor)

		// Draw body
		bodyTop := openY
		bodyBottom := closeY
		if closeY < openY {
			bodyTop = closeY
			bodyBottom = openY
		}
		if bodyBottom-bodyTop < 1 {
			bodyBottom = bodyTop + 1
		}
		drawFilledRect(img, x, bodyTop, x+config.CandleWidth, bodyBottom, candleColor)

		// Draw volume bar
		if config.ShowVolume && maxVolume > 0 {
			volHeight := int((c.Volume / maxVolume) * float64(volumeHeight))
			volY := volumeBottom - volHeight
			volColor := colorVolume
			if c.Close < c.Open {
				volColor = color.RGBA{R: 231, G: 76, B: 60, A: 128}
			} else {
				volColor = color.RGBA{R: 38, G: 166, B: 91, A: 128}
			}
			drawFilledRect(img, x, volY, x+config.CandleWidth, volumeBottom, volColor)
		}
	}

	// Draw Moving Averages
	if config.ShowMA && len(candles) > 50 {
		ma20 := calculateMA(candles, 20)
		ma50 := calculateMA(candles, 50)

		drawMALine(img, ma20, candles, chartLeft, chartTop, totalCandleWidth, maxPrice, priceRange, float64(chartHeight), colorMA20)
		drawMALine(img, ma50, candles, chartLeft, chartTop, totalCandleWidth, maxPrice, priceRange, float64(chartHeight), colorMA50)
	}

	// Draw price scale on right side
	textColor := colorTextDark
	if !config.DarkMode {
		textColor = colorTextLight
	}
	drawPriceScale(img, chartRight+5, chartTop, chartBottom, minPrice, maxPrice, textColor)

	// Draw title
	title := fmt.Sprintf("%s - %s", symbol, GetTimeframeName(interval))
	drawText(img, chartLeft, 20, title, textColor)

	// Draw current price
	lastCandle := candles[len(candles)-1]
	priceStr := formatPrice(lastCandle.Close)
	priceColor := colorBullish
	if lastCandle.Close < lastCandle.Open {
		priceColor = colorBearish
	}
	drawText(img, chartLeft+200, 20, fmt.Sprintf("Price: %s", priceStr), priceColor)

	// Draw timestamp
	timestamp := time.Now().Format("2006-01-02 15:04 UTC")
	drawText(img, chartRight-150, 20, timestamp, textColor)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// Helper functions

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	// Simple vertical/horizontal line (Bresenham for general case)
	if x1 == x2 {
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for y := y1; y <= y2; y++ {
			img.Set(x1, y, c)
		}
	} else if y1 == y2 {
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		for x := x1; x <= x2; x++ {
			img.Set(x, y1, c)
		}
	}
}

func drawFilledRect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, c)
		}
	}
}

func drawHorizontalGridLines(img *image.RGBA, x1, x2, y1, y2, count int, c color.Color) {
	step := (y2 - y1) / count
	for i := 0; i <= count; i++ {
		y := y1 + i*step
		for x := x1; x <= x2; x += 3 { // Dashed line
			img.Set(x, y, c)
		}
	}
}

func calculateMA(candles []Candlestick, period int) []float64 {
	ma := make([]float64, len(candles))
	for i := range candles {
		if i < period-1 {
			ma[i] = 0
			continue
		}
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += candles[i-j].Close
		}
		ma[i] = sum / float64(period)
	}
	return ma
}

func drawMALine(img *image.RGBA, ma []float64, candles []Candlestick, chartLeft, chartTop, totalCandleWidth int, maxPrice, priceRange, chartHeight float64, c color.Color) {
	prevX, prevY := 0, 0
	for i, val := range ma {
		if val == 0 {
			continue
		}
		x := chartLeft + i*totalCandleWidth + totalCandleWidth/2
		y := chartTop + int((maxPrice-val)/priceRange*chartHeight)

		if prevX != 0 && prevY != 0 {
			drawLineBresenham(img, prevX, prevY, x, y, c)
		}
		prevX, prevY = x, y
	}
}

func drawLineBresenham(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawText(img *image.RGBA, x, y int, text string, c color.Color) {
	col := color.RGBAModel.Convert(c).(color.RGBA)
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)
}

func drawPriceScale(img *image.RGBA, x, top, bottom int, minPrice, maxPrice float64, c color.Color) {
	steps := 5
	priceStep := (maxPrice - minPrice) / float64(steps)
	yStep := (bottom - top) / steps

	for i := 0; i <= steps; i++ {
		price := maxPrice - float64(i)*priceStep
		y := top + i*yStep
		priceStr := formatPrice(price)
		drawText(img, x, y+4, priceStr, c)
	}
}

func formatPrice(price float64) string {
	if price >= 1000 {
		return fmt.Sprintf("%.2f", price)
	} else if price >= 1 {
		return fmt.Sprintf("%.4f", price)
	} else if price >= 0.01 {
		return fmt.Sprintf("%.6f", price)
	}
	return fmt.Sprintf("%.8f", price)
}

// GenerateMultiTimeframeCharts generates charts for all timeframes
func GenerateMultiTimeframeCharts(symbol string, mode TradingMode, candleLimit int) ([]ChartData, error) {
	timeframes := GetTimeframesForMode(mode)
	charts := make([]ChartData, 0, len(timeframes))

	config := DefaultChartConfig()

	for _, tf := range timeframes {
		candles, err := FetchCandlesticks(symbol, tf, candleLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s data: %w", tf, err)
		}

		imgData, err := GenerateCandlestickChart(candles, symbol, tf, config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s chart: %w", tf, err)
		}

		charts = append(charts, ChartData{
			Interval:  tf,
			ImageData: imgData,
			Candles:   candles,
		})
	}

	return charts, nil
}

// ChartData holds generated chart information
type ChartData struct {
	Interval  BinanceInterval
	ImageData []byte
	Candles   []Candlestick
}

// CalculateTechnicalSummary provides a brief technical summary for context
func CalculateTechnicalSummary(candles []Candlestick) string {
	if len(candles) < 50 {
		return "Insufficient data"
	}

	last := candles[len(candles)-1]
	ma20 := calculateMA(candles, 20)
	ma50 := calculateMA(candles, 50)

	lastMA20 := ma20[len(ma20)-1]
	lastMA50 := ma50[len(ma50)-1]

	// Calculate trend
	trend := "NEUTRAL"
	if last.Close > lastMA20 && lastMA20 > lastMA50 {
		trend = "BULLISH"
	} else if last.Close < lastMA20 && lastMA20 < lastMA50 {
		trend = "BEARISH"
	}

	// Calculate volatility (ATR-like)
	var atrSum float64
	for i := len(candles) - 14; i < len(candles); i++ {
		atrSum += candles[i].High - candles[i].Low
	}
	atr := atrSum / 14
	volatility := "LOW"
	avgPrice := last.Close
	if atr/avgPrice > 0.02 {
		volatility = "HIGH"
	} else if atr/avgPrice > 0.01 {
		volatility = "MEDIUM"
	}

	// Calculate momentum (simple RSI approximation)
	gains, losses := 0.0, 0.0
	for i := len(candles) - 14; i < len(candles); i++ {
		change := candles[i].Close - candles[i].Open
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}
	rs := 1.0
	if losses > 0 {
		rs = gains / losses
	}
	rsi := 100 - (100 / (1 + rs))
	
	momentum := "NEUTRAL"
	if rsi > 70 {
		momentum = "OVERBOUGHT"
	} else if rsi < 30 {
		momentum = "OVERSOLD"
	} else if rsi > 55 {
		momentum = "BULLISH"
	} else if rsi < 45 {
		momentum = "BEARISH"
	}

	return fmt.Sprintf("Trend: %s | Volatility: %s | Momentum: %s | RSI: %.1f", trend, volatility, momentum, rsi)
}

// TradeLevels holds the entry/exit levels for chart marking
type TradeLevels struct {
	Entry float64
	SL    float64
	TP1   float64
	TP2   float64
	TP3   float64
}

// Colors for level lines
var (
	colorEntry = color.RGBA{R: 33, G: 150, B: 243, A: 255}  // Blue - Entry
	colorSL    = color.RGBA{R: 244, G: 67, B: 54, A: 255}   // Red - Stoploss
	colorTP    = color.RGBA{R: 76, G: 175, B: 80, A: 255}   // Green - Take Profit
)

// GenerateChartWithLevels creates a chart with Entry/SL/TP levels marked
func GenerateChartWithLevels(candles []Candlestick, symbol string, interval BinanceInterval, levels *TradeLevels) ([]byte, error) {
	if len(candles) == 0 {
		return nil, fmt.Errorf("no candle data to render")
	}

	config := DefaultChartConfig()
	config.Width = 1400
	config.Height = 700

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Fill background
	bgColor := colorBgDark
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Calculate price range (include levels in range)
	minPrice, maxPrice := candles[0].Low, candles[0].High
	maxVolume := candles[0].Volume
	for _, c := range candles {
		if c.Low < minPrice {
			minPrice = c.Low
		}
		if c.High > maxPrice {
			maxPrice = c.High
		}
		if c.Volume > maxVolume {
			maxVolume = c.Volume
		}
	}

	// Extend range to include levels
	if levels != nil {
		if levels.SL > 0 && levels.SL < minPrice {
			minPrice = levels.SL * 0.995
		}
		if levels.TP3 > 0 && levels.TP3 > maxPrice {
			maxPrice = levels.TP3 * 1.005
		} else if levels.TP2 > 0 && levels.TP2 > maxPrice {
			maxPrice = levels.TP2 * 1.005
		} else if levels.TP1 > 0 && levels.TP1 > maxPrice {
			maxPrice = levels.TP1 * 1.005
		}
	}

	// Add padding to price range
	priceRange := maxPrice - minPrice
	minPrice -= priceRange * 0.05
	maxPrice += priceRange * 0.05
	priceRange = maxPrice - minPrice

	// Calculate chart area
	chartLeft := config.Padding
	chartRight := config.Width - config.Padding - 80 // Extra space for level labels
	chartTop := config.Padding
	chartBottom := config.Height - config.Padding

	if config.ShowVolume {
		chartBottom = int(float64(config.Height-config.Padding) * 0.75)
	}

	chartWidth := chartRight - chartLeft
	chartHeight := chartBottom - chartTop

	// Volume area
	volumeTop := chartBottom + 10
	volumeBottom := config.Height - config.Padding
	volumeHeight := volumeBottom - volumeTop

	// Draw grid lines
	drawHorizontalGridLines(img, chartLeft, chartRight, chartTop, chartBottom, 5, colorGridDark)

	// Calculate candle positions
	totalCandleWidth := config.CandleWidth + config.CandleGap
	maxCandles := chartWidth / totalCandleWidth
	if len(candles) > maxCandles {
		candles = candles[len(candles)-maxCandles:]
	}

	// Draw candles
	for i, c := range candles {
		x := chartLeft + i*totalCandleWidth + config.CandleGap/2

		// Calculate Y positions
		openY := chartTop + int((maxPrice-c.Open)/priceRange*float64(chartHeight))
		closeY := chartTop + int((maxPrice-c.Close)/priceRange*float64(chartHeight))
		highY := chartTop + int((maxPrice-c.High)/priceRange*float64(chartHeight))
		lowY := chartTop + int((maxPrice-c.Low)/priceRange*float64(chartHeight))

		// Determine color
		candleColor := colorBullish
		if c.Close < c.Open {
			candleColor = colorBearish
		}

		// Draw wick
		wickX := x + config.CandleWidth/2
		drawLine(img, wickX, highY, wickX, lowY, candleColor)

		// Draw body
		bodyTop := openY
		bodyBottom := closeY
		if closeY < openY {
			bodyTop = closeY
			bodyBottom = openY
		}
		if bodyBottom-bodyTop < 1 {
			bodyBottom = bodyTop + 1
		}
		drawFilledRect(img, x, bodyTop, x+config.CandleWidth, bodyBottom, candleColor)

		// Draw volume bar
		if config.ShowVolume && maxVolume > 0 {
			volHeight := int((c.Volume / maxVolume) * float64(volumeHeight))
			volY := volumeBottom - volHeight
			volColor := colorVolume
			if c.Close < c.Open {
				volColor = color.RGBA{R: 231, G: 76, B: 60, A: 128}
			} else {
				volColor = color.RGBA{R: 38, G: 166, B: 91, A: 128}
			}
			drawFilledRect(img, x, volY, x+config.CandleWidth, volumeBottom, volColor)
		}
	}

	// Draw Moving Averages
	if config.ShowMA && len(candles) > 50 {
		ma20 := calculateMA(candles, 20)
		ma50 := calculateMA(candles, 50)
		drawMALine(img, ma20, candles, chartLeft, chartTop, totalCandleWidth, maxPrice, priceRange, float64(chartHeight), colorMA20)
		drawMALine(img, ma50, candles, chartLeft, chartTop, totalCandleWidth, maxPrice, priceRange, float64(chartHeight), colorMA50)
	}

	// Draw Level Lines (Entry, SL, TP)
	if levels != nil {
		// Entry Level (Blue)
		if levels.Entry > 0 {
			entryY := chartTop + int((maxPrice-levels.Entry)/priceRange*float64(chartHeight))
			drawHorizontalLevelLine(img, chartLeft, chartRight, entryY, colorEntry)
			drawText(img, chartRight+5, entryY+4, fmt.Sprintf("ENTRY %.2f", levels.Entry), colorEntry)
		}

		// Stoploss Level (Red)
		if levels.SL > 0 {
			slY := chartTop + int((maxPrice-levels.SL)/priceRange*float64(chartHeight))
			drawHorizontalLevelLine(img, chartLeft, chartRight, slY, colorSL)
			drawText(img, chartRight+5, slY+4, fmt.Sprintf("SL %.2f", levels.SL), colorSL)
		}

		// TP1 Level (Green)
		if levels.TP1 > 0 {
			tp1Y := chartTop + int((maxPrice-levels.TP1)/priceRange*float64(chartHeight))
			drawHorizontalLevelLine(img, chartLeft, chartRight, tp1Y, colorTP)
			drawText(img, chartRight+5, tp1Y+4, fmt.Sprintf("TP1 %.2f", levels.TP1), colorTP)
		}

		// TP2 Level (Green)
		if levels.TP2 > 0 {
			tp2Y := chartTop + int((maxPrice-levels.TP2)/priceRange*float64(chartHeight))
			drawHorizontalLevelLine(img, chartLeft, chartRight, tp2Y, colorTP)
			drawText(img, chartRight+5, tp2Y+4, fmt.Sprintf("TP2 %.2f", levels.TP2), colorTP)
		}

		// TP3 Level (Green)
		if levels.TP3 > 0 {
			tp3Y := chartTop + int((maxPrice-levels.TP3)/priceRange*float64(chartHeight))
			drawHorizontalLevelLine(img, chartLeft, chartRight, tp3Y, colorTP)
			drawText(img, chartRight+5, tp3Y+4, fmt.Sprintf("TP3 %.2f", levels.TP3), colorTP)
		}
	}

	// Draw price scale
	drawPriceScale(img, chartRight+5, chartTop, chartBottom, minPrice, maxPrice, colorTextDark)

	// Draw title
	title := fmt.Sprintf("%s - %s | ENTRY CHART", symbol, GetTimeframeName(interval))
	drawText(img, chartLeft, 20, title, colorTextDark)

	// Draw current price
	lastCandle := candles[len(candles)-1]
	priceStr := formatPrice(lastCandle.Close)
	priceColor := colorBullish
	if lastCandle.Close < lastCandle.Open {
		priceColor = colorBearish
	}
	drawText(img, chartLeft+300, 20, fmt.Sprintf("Price: %s", priceStr), priceColor)

	// Draw timestamp
	timestamp := time.Now().Format("2006-01-02 15:04 UTC")
	drawText(img, chartRight-150, 20, timestamp, colorTextDark)

	// Draw legend
	drawText(img, chartLeft, chartBottom+50, "🔵 Entry  🔴 Stoploss  🟢 Take Profit  🟡 MA20  🟣 MA50", colorTextDark)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// drawHorizontalLevelLine draws a dashed horizontal line for levels
func drawHorizontalLevelLine(img *image.RGBA, x1, x2, y int, c color.Color) {
	for x := x1; x <= x2; x++ {
		// Solid line (or use x += 2 for dashed)
		img.Set(x, y, c)
		img.Set(x, y-1, c)
	}
}
