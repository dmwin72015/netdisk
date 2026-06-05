package storage

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"

	_ "image/gif"
	_ "image/png"
	_ "golang.org/x/image/webp"
)

const (
	ThumbDir    = "thumbnails"
	ThumbWidth  = 400
	ThumbHeight = 400
	ThumbQuality = 85
)

func ThumbnailAbsPath(root, fileHash string) string {
	return filepath.Join(root, ThumbDir, fileHash[0:2], fileHash[2:4], fileHash+".jpg")
}

func EnsureThumbDir(root, fileHash string) error {
	dir := filepath.Join(root, ThumbDir, fileHash[0:2], fileHash[2:4])
	return os.MkdirAll(dir, 0o755)
}

func GenerateThumbnail(inputPath, outputPath string) error {
	src, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	newW, newH := w, h
	if w > h {
		if w > ThumbWidth {
			newW = ThumbWidth
			newH = h * ThumbWidth / w
		}
	} else {
		if h > ThumbHeight {
			newH = ThumbHeight
			newW = w * ThumbHeight / h
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create thumbnail: %w", err)
	}
	defer out.Close()

	return jpeg.Encode(out, dst, &jpeg.Options{Quality: ThumbQuality})
}

func ServeThumbnail(root, filesDir, fileHash string) (string, error) {
	thumbPath := ThumbnailAbsPath(root, fileHash)
	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	inputPath := AbsPath(root, fileHash, filesDir)
	if _, err := os.Stat(inputPath); err != nil {
		return "", fmt.Errorf("original file not found: %w", err)
	}

	if err := GenerateThumbnail(inputPath, thumbPath); err != nil {
		return "", fmt.Errorf("generate thumbnail: %w", err)
	}

	return thumbPath, nil
}
