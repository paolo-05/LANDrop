package qrcode

import (
	"bytes"
	"image"
	"image/png"
	"log"

	"github.com/skip2/go-qrcode"
)

func GenerateQRImage(url string) image.Image {
	pngData, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		log.Println("QR generation failed:", err)
		return nil
	}
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		log.Println("QR decode failed:", err)
		return nil
	}
	return img
}
