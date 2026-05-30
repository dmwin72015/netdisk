package media

import (
	"bufio"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildFFmpegArgs(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		outputDir string
	}{
		{name: "basic", inputPath: "/tmp/video.mp4", outputDir: "/tmp/out"},
		{name: "nested paths", inputPath: "/data/uploads/a/b/c.mkv", outputDir: "/data/transcode/job123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := BuildFFmpegArgs(tt.inputPath, tt.outputDir)

			// Must contain -i <inputPath>
			if !containsArgPair(args, "-i", tt.inputPath) {
				t.Errorf("expected -i %q in args", tt.inputPath)
			}

			// Must contain -hls_segment_filename with seg%04d.ts
			segPattern := filepath.Join(tt.outputDir, "seg%04d.ts")
			if !containsArgPair(args, "-hls_segment_filename", segPattern) {
				t.Errorf("expected -hls_segment_filename %q in args", segPattern)
			}

			// Last arg must be the m3u8 output
			m3u8 := filepath.Join(tt.outputDir, "index.m3u8")
			if args[len(args)-1] != m3u8 {
				t.Errorf("last arg = %q, want %q", args[len(args)-1], m3u8)
			}

			// Must contain -progress pipe:1
			if !containsArgPair(args, "-progress", "pipe:1") {
				t.Errorf("expected -progress pipe:1 in args")
			}

			// Must contain key codec flags
			if !containsArgPair(args, "-c:v", "libx264") {
				t.Errorf("expected -c:v libx264 in args")
			}
			if !containsArgPair(args, "-c:a", "aac") {
				t.Errorf("expected -c:a aac in args")
			}
		})
	}
}

func containsArgPair(args []string, flag, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

func TestParseProgress(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		totalDuration int32
		wantPcts      []int
	}{
		{
			name: "normal progress with end",
			input: "out_time_us=5000000\n" +
				"progress=continue\n" +
				"out_time_us=10000000\n" +
				"progress=end\n",
			totalDuration: 20,
			// The progress check runs after every line, not just progress= lines.
			// out_time_us=5M  -> currentTime=5, pct=25, callback(25)
			// progress=continue -> currentTime still 5, pct=25, callback(25)
			// out_time_us=10M -> currentTime=10, pct=50, callback(50)
			// progress=end -> callback(100)
			wantPcts: []int{25, 25, 50, 100},
		},
		{
			name: "end immediately",
			input: "progress=end\n",
			totalDuration: 10,
			wantPcts:      []int{100},
		},
		{
			name:          "zero duration skips percentage",
			input:         "out_time_us=5000000\nprogress=end\n",
			totalDuration: 0,
			// Only the "end" line triggers 100%
			wantPcts: []int{100},
		},
		{
			name: "clamps at 100",
			input: "out_time_us=30000000\n" +
				"progress=continue\n" +
				"progress=end\n",
			totalDuration: 20,
			// out_time_us=30M -> pct=100 (clamped), callback(100)
			// progress=continue -> pct=100 (clamped), callback(100)
			// progress=end -> callback(100)
			wantPcts: []int{100, 100, 100},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			var got []int
			ParseProgress(scanner, tt.totalDuration, func(pct int) {
				got = append(got, pct)
			})
			if len(got) != len(tt.wantPcts) {
				t.Fatalf("callback called %d times, want %d; got %v", len(got), len(tt.wantPcts), got)
			}
			for i, g := range got {
				if g != tt.wantPcts[i] {
					t.Errorf("call %d: got %d, want %d", i, g, tt.wantPcts[i])
				}
			}
		})
	}
}

func TestParseProgressCallbackNotCalledOnEmptyInput(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(""))
	called := false
	ParseProgress(scanner, 10, func(pct int) {
		called = true
	})
	if called {
		t.Error("callback should not be called for empty input")
	}
}
