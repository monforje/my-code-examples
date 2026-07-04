package avatar

import (
	"crypto/md5"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strings"
)

const defaultSize = 420

// Identicon holds the computed grid and color for a 5x5 GitHub-style avatar.
type Identicon struct {
	Grid  [5][5]bool
	Color color.RGBA
}

// Generate creates an Identicon from any input string:
//
//  1. MD5-hash the lowercased input → 16 bytes
//  2. Derive HSL color from bytes 0-2, convert to RGB
//  3. Map bytes 3-17 onto a 5x3 half-grid (even byte → filled cell)
//  4. Mirror columns 0-1 onto columns 4-3 to produce a symmetric 5x5 grid
func Generate(input string) Identicon {
	hash := md5.Sum([]byte(strings.ToLower(input)))

	h := float64(hash[0]) / 255.0 * 360.0
	s := 45.0 + (float64(hash[1])/255.0)*10.0
	l := 65.0 + (float64(hash[2])/255.0)*10.0
	r, g, b := hslToRGB(h, s/100.0, l/100.0)

	var grid [5][5]bool
	for row := 0; row < 5; row++ {
		for col := 0; col < 3; col++ {
			byteIndex := (3 + row*3 + col) % len(hash)
			filled := hash[byteIndex]%2 == 0
			grid[row][col] = filled
			mirror := 4 - col
			grid[row][mirror] = filled
		}
	}

	return Identicon{
		Grid:  grid,
		Color: color.RGBA{R: r, G: g, B: b, A: 255},
	}
}

// Render writes the identicon as PNG to w.
func (id Identicon) Render(w io.Writer, size int) error {
	if size <= 0 {
		size = defaultSize
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))

	bg := color.RGBA{246, 248, 250, 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
		}
	}

	cellSize := float64(size) / 7.0
	padding := cellSize

	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			if !id.Grid[row][col] {
				continue
			}
			x0 := int(padding + float64(col)*cellSize)
			y0 := int(padding + float64(row)*cellSize)
			x1 := int(padding + float64(col+1)*cellSize)
			y1 := int(padding + float64(row+1)*cellSize)
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					img.Set(x, y, id.Color)
				}
			}
		}
	}

	return png.Encode(w, img)
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}
