package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"

	"fyne.io/fyne/v2"
)

// getPanoIcon creates a modern clipboard icon for system tray and exe
func getPanoIcon() fyne.Resource {
	size := 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Colors
	primaryBlue := color.RGBA{R: 59, G: 130, B: 246, A: 255}    // Modern blue
	darkBlue := color.RGBA{R: 37, G: 99, B: 235, A: 255}        // Darker blue for depth
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	lightGray := color.RGBA{R: 243, G: 244, B: 246, A: 255}
	clipColor := color.RGBA{R: 75, G: 85, B: 99, A: 255}        // Dark gray for clip

	// Clear background (transparent)
	transparent := color.RGBA{R: 0, G: 0, B: 0, A: 0}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, transparent)
		}
	}

	// Draw clipboard board (rounded rectangle)
	boardLeft := 8
	boardRight := 56
	boardTop := 10
	boardBottom := 58
	cornerRadius := 6.0

	for y := boardTop; y < boardBottom; y++ {
		for x := boardLeft; x < boardRight; x++ {
			if isInsideRoundedRect(x, y, boardLeft, boardTop, boardRight, boardBottom, cornerRadius) {
				// Gradient effect - darker at bottom
				t := float64(y-boardTop) / float64(boardBottom-boardTop)
				c := lerpColor(primaryBlue, darkBlue, t*0.3)
				img.Set(x, y, c)
			}
		}
	}

	// Draw paper area (inner white rectangle)
	paperLeft := 12
	paperRight := 52
	paperTop := 18
	paperBottom := 54
	paperRadius := 3.0

	for y := paperTop; y < paperBottom; y++ {
		for x := paperLeft; x < paperRight; x++ {
			if isInsideRoundedRect(x, y, paperLeft, paperTop, paperRight, paperBottom, paperRadius) {
				img.Set(x, y, white)
			}
		}
	}

	// Draw text lines on paper
	lineColor := lightGray
	lines := []struct{ y, x1, x2 int }{
		{24, 16, 42},
		{30, 16, 48},
		{36, 16, 38},
		{42, 16, 45},
		{48, 16, 35},
	}
	for _, line := range lines {
		for x := line.x1; x < line.x2; x++ {
			img.Set(x, line.y, lineColor)
			img.Set(x, line.y+1, lineColor)
		}
	}

	// Draw clip at top
	clipWidth := 20
	clipHeight := 10
	clipLeft := (size - clipWidth) / 2
	clipRight := clipLeft + clipWidth
	clipTop := 4
	clipBottom := clipTop + clipHeight

	// Clip outer
	for y := clipTop; y < clipBottom; y++ {
		for x := clipLeft; x < clipRight; x++ {
			if isInsideRoundedRect(x, y, clipLeft, clipTop, clipRight, clipBottom, 3.0) {
				img.Set(x, y, clipColor)
			}
		}
	}

	// Clip inner hole
	holeLeft := clipLeft + 4
	holeRight := clipRight - 4
	holeTop := clipTop + 3
	holeBottom := clipBottom - 1

	for y := holeTop; y < holeBottom; y++ {
		for x := holeLeft; x < holeRight; x++ {
			if isInsideRoundedRect(x, y, holeLeft, holeTop, holeRight, holeBottom, 2.0) {
				img.Set(x, y, primaryBlue)
			}
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}

	return fyne.NewStaticResource("pano-icon.png", buf.Bytes())
}

// isInsideRoundedRect checks if a point is inside a rounded rectangle
func isInsideRoundedRect(x, y, left, top, right, bottom int, radius float64) bool {
	// Check corners
	corners := []struct{ cx, cy int }{
		{left + int(radius), top + int(radius)},         // top-left
		{right - int(radius) - 1, top + int(radius)},    // top-right
		{left + int(radius), bottom - int(radius) - 1},  // bottom-left
		{right - int(radius) - 1, bottom - int(radius) - 1}, // bottom-right
	}

	// Check if in corner regions
	for i, corner := range corners {
		var inCornerRegion bool
		switch i {
		case 0: // top-left
			inCornerRegion = x < corner.cx && y < corner.cy
		case 1: // top-right
			inCornerRegion = x > corner.cx && y < corner.cy
		case 2: // bottom-left
			inCornerRegion = x < corner.cx && y > corner.cy
		case 3: // bottom-right
			inCornerRegion = x > corner.cx && y > corner.cy
		}

		if inCornerRegion {
			dx := float64(x - corner.cx)
			dy := float64(y - corner.cy)
			if math.Sqrt(dx*dx+dy*dy) > radius {
				return false
			}
		}
	}

	// Check if inside bounding box
	return x >= left && x < right && y >= top && y < bottom
}

// lerpColor linearly interpolates between two colors
func lerpColor(c1, c2 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c1.R) + t*(float64(c2.R)-float64(c1.R))),
		G: uint8(float64(c1.G) + t*(float64(c2.G)-float64(c1.G))),
		B: uint8(float64(c1.B) + t*(float64(c2.B)-float64(c1.B))),
		A: uint8(float64(c1.A) + t*(float64(c2.A)-float64(c1.A))),
	}
}
