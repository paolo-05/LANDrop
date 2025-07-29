package qrcode

import (
	"testing"
)

func TestGenerateQRImageValid(t *testing.T) {
	testCases := []string{
		"https://example.com",
		"http://localhost:8080",
		"Hello World",
		"test@example.com",
		"A longer text string with various characters 123!@#",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			img := GenerateQRImage(tc)
			if img == nil {
				t.Errorf("GenerateQRImage returned nil for valid input: %s", tc)
				return
			}

			// Check that we got a valid image
			bounds := img.Bounds()
			if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
				t.Errorf("Generated image has invalid dimensions: %dx%d", bounds.Dx(), bounds.Dy())
			}

			// QR codes should be square
			if bounds.Dx() != bounds.Dy() {
				t.Errorf("QR code should be square, got %dx%d", bounds.Dx(), bounds.Dy())
			}

			// Check expected size (should be 256x256 based on the function)
			if bounds.Dx() != 256 || bounds.Dy() != 256 {
				t.Errorf("Expected 256x256 image, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestGenerateQRImageEmpty(t *testing.T) {
	img := GenerateQRImage("")
	// Empty string should fail gracefully and return nil
	if img != nil {
		t.Error("GenerateQRImage should return nil for empty string")
	}
}

func TestGenerateQRImageSpace(t *testing.T) {
	// Test with just a space (minimal valid content)
	img := GenerateQRImage(" ")
	if img == nil {
		t.Error("GenerateQRImage should handle space character")
		return
	}

	bounds := img.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 256 {
		t.Errorf("Expected 256x256 image for space, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateQRImageLargeInput(t *testing.T) {
	// Create a moderately large input string (QR codes have size limits)
	largeInput := ""
	for i := 0; i < 100; i++ { // Reduced from 1000 to 100
		largeInput += "This is a long string for QR testing. "
	}

	img := GenerateQRImage(largeInput)
	// For very large inputs, QR code generation might fail
	// The function should handle this gracefully and return nil
	if img != nil {
		// If it succeeds, the image should still be valid
		bounds := img.Bounds()
		if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
			t.Error("Generated image has invalid dimensions")
		}
	}
	// If img is nil, that's acceptable for oversized input
}

func TestGenerateQRImageSpecialCharacters(t *testing.T) {
	specialChars := []string{
		"🌟🚀💻",             // Emojis
		"中文测试",            // Chinese characters
		"Тест на русском", // Cyrillic
		"العربية",         // Arabic
		"ñáéíóú",          // Accented characters
	}

	for _, sc := range specialChars {
		t.Run(sc, func(t *testing.T) {
			img := GenerateQRImage(sc)
			if img == nil {
				t.Errorf("GenerateQRImage failed for special characters: %s", sc)
				return
			}

			bounds := img.Bounds()
			if bounds.Dx() != 256 || bounds.Dy() != 256 {
				t.Errorf("Expected 256x256 image, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestGenerateQRImageType(t *testing.T) {
	img := GenerateQRImage("test")
	if img == nil {
		t.Fatal("GenerateQRImage returned nil")
	}

	// Check that we can get color information (confirming it's a valid image)
	bounds := img.Bounds()
	if bounds.Dx() > 0 && bounds.Dy() > 0 {
		// Try to get color at a pixel to ensure it's a valid image
		_ = img.At(bounds.Min.X, bounds.Min.Y)
		// If we get here without panic, the image is valid
	}
}
