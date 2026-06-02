package media

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// BuildFFmpegArgs returns the FFmpeg arguments for HLS transcoding.
func BuildFFmpegArgs(inputPath, outputDir string) []string {
	return []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "22",
		"-c:a", "aac",
		"-ac", "2",
		"-ar", "44100",
		"-b:a", "128k",
		"-vf", "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease",
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(outputDir, "seg%04d.ts"),
		"-progress", "pipe:1",
		filepath.Join(outputDir, "index.m3u8"),
	}
}

// ExtractPoster extracts a single frame from the video as a JPEG poster image.
// It seeks to 10% of the duration to avoid a black frame at the start.
func ExtractPoster(ffmpegPath, inputPath, outputPath string, durationSec int32) error {
	seekSec := float64(durationSec) * 0.1
	if seekSec < 1 {
		seekSec = 1
	}

	cmd := exec.Command(ffmpegPath,
		"-ss", fmt.Sprintf("%.1f", seekSec),
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extract poster: %w: %s", err, string(out))
	}
	return nil
}

// ProbeDuration returns the duration in seconds of a media file using ffprobe.
func ProbeDuration(ffprobePath, inputPath string) (int32, error) {
	cmd := exec.Command(ffprobePath,
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		inputPath,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}

	durationStr := strings.TrimSpace(string(out))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}

	return int32(duration), nil
}

// ParseProgress reads FFmpeg progress output and returns the percentage (0-100).
// It reads from the progress pipe and calls onProgress for each update.
func ParseProgress(scanner *bufio.Scanner, totalDuration int32, onProgress func(pct int)) {
	var currentTime float64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "out_time_us=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				microseconds, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err == nil {
					currentTime = microseconds / 1_000_000.0
				}
			}
		}

		if strings.HasPrefix(line, "progress=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == "end" {
				onProgress(100)
				return
			}
		}

		if totalDuration > 0 && currentTime > 0 {
			pct := int((currentTime / float64(totalDuration)) * 100)
			if pct > 100 {
				pct = 100
			}
			if pct > 0 {
				onProgress(pct)
			}
		}
	}
}
