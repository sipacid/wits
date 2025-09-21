package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"regexp"

	"github.com/ipinfo/go/v2/ipinfo"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	templateGIFPath = "/assets/wits.gif"
	fontGIFFile     = "/assets/font.ttf"
)

func generateButton(filePath, userAgent string, data *ipinfo.Core) error {
	// Template GIF
	gifFile, err := os.Open(templateGIFPath)
	if err != nil {
		return err
	}
	defer gifFile.Close()

	templateGIF, err := gif.DecodeAll(gifFile)
	if err != nil {
		return err
	}

	// Setup font
	fontBytes, err := os.ReadFile(fontGIFFile)
	if err != nil {
		return fmt.Errorf("failed to read font file: %v", err)
	}

	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("failed to create font face: %v", err)
	}
	defer face.Close()

	re := regexp.MustCompile(`\(([^)]+)\)`)
	userAgentPart := re.FindStringSubmatch(userAgent)
	text := fmt.Sprintf("%v, %v, %v, %v", data.IP, data.City, data.Country, userAgentPart[1])

	// Measure text width
	tempImg := image.NewRGBA(image.Rect(0, 0, 1000, 100))
	drawer := &font.Drawer{
		Dst:  tempImg,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
	}
	textWidth := drawer.MeasureString(text).Round()

	// Animation parameters
	buttonWidth := 88
	buttonHeight := 31
	speed := 2
	totalDistance := buttonWidth + textWidth
	textFramesNeeded := (totalDistance / speed) + 1

	newGIF := &gif.GIF{
		Image:     make([]*image.Paletted, 0),
		Delay:     make([]int, 0),
		LoopCount: 0,
	}

	for i, originalFrame := range templateGIF.Image {
		newGIF.Image = append(newGIF.Image, originalFrame)

		if i < len(templateGIF.Delay) {
			newGIF.Delay = append(newGIF.Delay, templateGIF.Delay[i])
		} else {
			newGIF.Delay = append(newGIF.Delay, 10) // 100ms default
		}
	}

	lastFrameIndex := len(templateGIF.Image) - 1
	baseFrame := templateGIF.Image[lastFrameIndex]

	for frame := range textFramesNeeded {
		newImage := image.NewRGBA(image.Rect(0, 0, buttonWidth, buttonHeight))
		draw.Draw(newImage, newImage.Bounds(), baseFrame, image.Point{}, draw.Src)

		textX := buttonWidth - (frame * speed)
		textY := buttonHeight - 10

		if textX > -textWidth && textX < buttonWidth {
			drawTextWithShadow(newImage, face, text, textX, textY)
		}

		palettedImage := convertToPaletted(newImage, baseFrame.Palette)
		newGIF.Image = append(newGIF.Image, palettedImage)

		newGIF.Delay = append(newGIF.Delay, 8) // 80ms per frame for text animation
	}

	outputFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return gif.EncodeAll(outputFile, newGIF)
}

func drawTextWithShadow(img *image.RGBA, face font.Face, text string, x, y int) {
	shadowDrawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{0, 0, 0, 180}),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(x + 1),
			Y: fixed.I(y + 1),
		},
	}
	shadowDrawer.DrawString(text)

	textDrawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		},
	}
	textDrawer.DrawString(text)
}

func convertToPaletted(rgba *image.RGBA, templatePalette color.Palette) *image.Paletted {
	bounds := rgba.Bounds()

	extendedPalette := make(color.Palette, len(templatePalette))
	copy(extendedPalette, templatePalette)

	textColors := []color.RGBA{
		{0, 0, 0, 255},       // black (shadow)
		{0, 0, 0, 180},       // semi-transparent black
		{50, 50, 50, 255},    // dark gray
		{100, 100, 100, 255}, // medium gray
		{200, 200, 200, 255}, // light gray
		{220, 220, 220, 255}, // lighter gray
		{240, 240, 240, 255}, // very light gray
		{255, 255, 255, 255}, // white (text)
	}

	for _, newColor := range textColors {
		hasColor := false
		for _, existingColor := range extendedPalette {
			if colorsEqual(existingColor, newColor) {
				hasColor = true
				break
			}
		}
		if !hasColor && len(extendedPalette) < 256 {
			extendedPalette = append(extendedPalette, newColor)
		}
	}

	paletted := image.NewPaletted(bounds, extendedPalette)
	draw.Draw(paletted, bounds, rgba, bounds.Min, draw.Src)

	return paletted
}

func colorsEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
