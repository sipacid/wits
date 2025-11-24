package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"os/exec"

	"github.com/golang/freetype"
	"github.com/oschwald/geoip2-golang/v2"
)

const (
	templateVideoPath = "/assets/video.mp4"
	templateImgPath   = "/assets/image.png"
	fontFile          = "/assets/font.ttf"
)

func generateVideo(filePath string, data *geoip2.City) error {
	imgPath, err := generateImage(data.Traits.IPAddress.String(), data)
	if err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg",
		"-i", templateVideoPath,
		"-i", imgPath,
		"-filter_complex", "[0:v][1:v] overlay=0:0:enable='between(t,6,8)'",
		"-c:v", "libx264",
		"-c:a", "copy",
		"-preset", "ultrafast",
		filePath)
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error occurred when trying to generate video (%v): %v", err, stdErr.String())
	}

	if err := os.Remove(imgPath); err != nil {
		return fmt.Errorf("error occurred when trying to delete image file: %v", err)
	}

	return nil
}

func generateImage(ipStr string, data *geoip2.City) (string, error) {
	filePath := fmt.Sprintf("/tmp/%v.png", generateFilename(ipStr))
	lines := []string{
		ipStr,
		fmt.Sprintf("%v, %v [%v, %v]", data.City.Names.English, data.Country.Names.English, *data.Location.Latitude, *data.Location.Longitude),
	}
	fontSize := 36.0
	fontColour := image.White

	imgFile, err := os.Open(templateImgPath)
	if err != nil {
		return "", err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return "", err
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, img.Bounds(), img, image.Point{}, draw.Src)

	fontData, err := os.ReadFile(fontFile)
	if err != nil {
		return "", err
	}

	font, err := freetype.ParseFont(fontData)
	if err != nil {
		return "", err
	}

	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetDst(rgba)
	ctx.SetClip(rgba.Bounds())
	ctx.SetFont(font)
	ctx.SetFontSize(fontSize)
	ctx.SetSrc(fontColour)

	xPos := 70
	for i, line := range lines {
		displayText := line
		yPos := img.Bounds().Max.Y/2 + (i * int(ctx.PointToFixed(fontSize)>>6))
		pt := freetype.Pt(xPos, yPos)
		if _, err := ctx.DrawString(displayText, pt); err != nil {
			return "", err
		}

		ctx.SetFontSize(fontSize / 2)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	err = png.Encode(outFile, rgba)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
