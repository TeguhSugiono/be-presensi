package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

const (
	maxImageSide         = 1280
	minImageSide         = 320
	maxImagePixels int64 = 80_000_000
	maxStoredSize        = 300 * 1024 // 300 KiB
	initialQuality       = 80
	minQuality           = 35
)

var (
	ErrPhotoRequired    = errors.New("photo is required")
	ErrInvalidPhoto     = errors.New("only valid JPEG or PNG images are allowed")
	ErrInvalidPhotoType = errors.New(
		"photo category must be check-in or check-out",
	)
	ErrPhotoCompression = errors.New(
		"unable to compress photo below 300 KB",
	)
)

// SaveAttendancePhoto validates, resizes, compresses, and stores an
// attendance photo. It returns a relative path for database storage.
func SaveAttendancePhoto(
	fileHeader *multipart.FileHeader,
	category string,
) (string, error) {
	if fileHeader == nil || fileHeader.Size <= 0 {
		return "", ErrPhotoRequired
	}

	if category != "check-in" && category != "check-out" {
		return "", ErrInvalidPhotoType
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded photo: %w", err)
	}
	defer file.Close()

	// No fixed upload-size limit. Image dimensions are checked to
	// reduce the risk of excessive memory use during decoding.
	config, format, err := image.DecodeConfig(file)
	if err != nil || (format != "jpeg" && format != "png") {
		return "", ErrInvalidPhoto
	}

	if config.Width <= 0 || config.Height <= 0 ||
		int64(config.Width) > maxImagePixels/int64(config.Height) {
		return "", ErrInvalidPhoto
	}

	if int64(config.Width)*int64(config.Height) > maxImagePixels {
		return "", ErrInvalidPhoto
	}

	// Reopen the uploaded file because DecodeConfig consumes its reader.
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("rewind uploaded photo: %w", err)
	}

	img, err := imaging.Decode(file)
	if err != nil {
		return "", ErrInvalidPhoto
	}

	// First resize the image to a practical maximum resolution.
	img = imaging.Fit(img, maxImageSide, maxImageSide, imaging.Lanczos)

	// Retry with lower quality, then progressively smaller dimensions
	// until the encoded JPEG is below the storage target.
	var encoded bytes.Buffer
	currentQuality := initialQuality
	currentSide := maxImageSide

	for {
		encoded.Reset()

		err = imaging.Encode(
			&encoded,
			img,
			imaging.JPEG,
			imaging.JPEGQuality(currentQuality),
		)
		if err != nil {
			return "", fmt.Errorf("encode attendance photo: %w", err)
		}

		if encoded.Len() < maxStoredSize {
			break
		}

		if currentQuality > minQuality {
			currentQuality -= 10
			if currentQuality < minQuality {
				currentQuality = minQuality
			}
			continue
		}

		if currentSide <= minImageSide {
			return "", ErrPhotoCompression
		}

		currentSide -= 160
		if currentSide < minImageSide {
			currentSide = minImageSide
		}

		img = imaging.Fit(img, currentSide, currentSide, imaging.Lanczos)
		currentQuality = initialQuality
	}

	dir := filepath.Join("assets", "attendance", category)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("create photo directory: %w", err)
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate photo filename: %w", err)
	}

	filename := hex.EncodeToString(randomBytes) + ".jpg"
	finalPath := filepath.Join(dir, filename)

	tempFile, err := os.CreateTemp(dir, ".attendance-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temporary photo: %w", err)
	}

	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(encoded.Bytes()); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("write attendance photo: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("sync attendance photo: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close attendance photo: %w", err)
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", fmt.Errorf("save attendance photo: %w", err)
	}

	return filepath.ToSlash(finalPath), nil
}
