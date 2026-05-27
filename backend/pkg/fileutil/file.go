package fileutil

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

const (
	// MaxFileSize is the hard cap for individual uploads (2 GiB).
	MaxFileSize int64 = 2 * 1024 * 1024 * 1024
	// MaxThumbnailBytes caps user-uploaded thumbnail images.
	MaxThumbnailBytes int64 = 5 * 1024 * 1024
	// MagicReadSize is the number of bytes read from the head of an upload
	// to verify its container before accepting it.
	MagicReadSize = 32
)

var (
	ErrFileTooLarge    = errors.New("file too large")
	ErrUnsupportedType = errors.New("unsupported file type")
)

// Container identifies what kind of media wrapper an upload uses. It maps
// 1:1 to the file extension we use when persisting the upload, so that
// ffmpeg/ffprobe pick up the right demuxer.
type Container string

const (
	ContainerMP4    Container = "mp4"   // ISO base media (covers MOV/3GP — same ftyp signature)
	ContainerWebM   Container = "webm"  // EBML family (covers MKV too)
	ContainerAVI    Container = "avi"
	ContainerFLV    Container = "flv"
	ContainerASF    Container = "asf"  // Microsoft ASF (WMV)
	ContainerOgg    Container = "ogg"  // Ogg (OGV/Theora/Vorbis)
	ContainerMPEGPS Container = "mpg"  // MPEG-1/2 program stream
)

// Ext returns the canonical lowercase extension (with leading dot) for a
// given container.
func (c Container) Ext() string {
	switch c {
	case ContainerMP4:
		return ".mp4"
	case ContainerWebM:
		return ".webm"
	case ContainerAVI:
		return ".avi"
	case ContainerFLV:
		return ".flv"
	case ContainerASF:
		return ".wmv"
	case ContainerOgg:
		return ".ogv"
	case ContainerMPEGPS:
		return ".mpg"
	default:
		return ".bin"
	}
}

// DetectContainer inspects the leading bytes of an upload to identify its
// container. Returns an empty string when nothing matches.
func DetectContainer(head []byte) Container {
	switch {
	case isMP4(head):
		return ContainerMP4
	case isWebM(head):
		return ContainerWebM
	case isAVI(head):
		return ContainerAVI
	case isFLV(head):
		return ContainerFLV
	case isASF(head):
		return ContainerASF
	case isOgg(head):
		return ContainerOgg
	case isMPEGPS(head):
		return ContainerMPEGPS
	}
	return ""
}

// IsMP4 is kept for backwards compatibility with existing tests; new code
// should use DetectContainer.
func IsMP4(head []byte) bool { return isMP4(head) }

// IsWebM reports whether head starts with the EBML signature shared by WebM
// and Matroska containers.
func IsWebM(head []byte) bool { return isWebM(head) }

func isMP4(head []byte) bool {
	// Any ftyp box at offset 4 — let ffmpeg decide whether the brand is
	// actually decodable. The whitelist approach reject too many real-world
	// files (MOV with `qt  `, 3GP variants, fragmented MP4, etc.).
	return len(head) >= 8 && bytes.Equal(head[4:8], []byte("ftyp"))
}

// EBML magic: 0x1A 0x45 0xDF 0xA3. WebM and Matroska share this header; the
// DocType further inside identifies which. For our purposes any EBML container
// is acceptable since ffmpeg demuxes them uniformly.
var ebmlMagic = []byte{0x1A, 0x45, 0xDF, 0xA3}

func isWebM(head []byte) bool {
	return len(head) >= 4 && bytes.Equal(head[:4], ebmlMagic)
}

// isAVI looks for "RIFF" at offset 0 and "AVI " at offset 8 (between is the
// 4-byte little-endian file size).
func isAVI(head []byte) bool {
	return len(head) >= 12 &&
		bytes.Equal(head[:4], []byte("RIFF")) &&
		bytes.Equal(head[8:12], []byte("AVI "))
}

// isFLV checks the 3-byte "FLV" signature followed by version (typically 1).
func isFLV(head []byte) bool {
	return len(head) >= 4 && bytes.Equal(head[:3], []byte("FLV"))
}

// ASF/WMV files start with a fixed 16-byte GUID identifying the header object.
var asfHeaderGUID = []byte{
	0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11,
	0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C,
}

func isASF(head []byte) bool {
	return len(head) >= 16 && bytes.Equal(head[:16], asfHeaderGUID)
}

// Ogg pages all start with "OggS".
func isOgg(head []byte) bool {
	return len(head) >= 4 && bytes.Equal(head[:4], []byte("OggS"))
}

// MPEG-1/2 program stream begins with pack header start code 0x000001BA.
var mpegPSStart = []byte{0x00, 0x00, 0x01, 0xBA}

func isMPEGPS(head []byte) bool {
	return len(head) >= 4 && bytes.Equal(head[:4], mpegPSStart)
}

// ValidateMultipart performs a streaming validation of the upload header:
// it opens the file, reads MagicReadSize bytes to identify the container,
// and returns the opened reader (positioned at offset 0) plus the detected
// container. Caller takes ownership of the returned file and MUST close it.
func ValidateMultipart(fh *multipart.FileHeader, maxBytes int64) (multipart.File, Container, error) {
	if maxBytes <= 0 {
		maxBytes = MaxFileSize
	}
	if fh.Size > maxBytes {
		return nil, "", ErrFileTooLarge
	}
	src, err := fh.Open()
	if err != nil {
		return nil, "", err
	}
	head := make([]byte, MagicReadSize)
	n, err := io.ReadFull(src, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		_ = src.Close()
		return nil, "", err
	}
	container := DetectContainer(head[:n])
	if container == "" {
		_ = src.Close()
		return nil, "", ErrUnsupportedType
	}
	if seeker, ok := src.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			_ = src.Close()
			return nil, "", err
		}
		return src, container, nil
	}
	// Fallback: reopen for callers whose multipart.File doesn't implement Seeker.
	_ = src.Close()
	reopened, err := fh.Open()
	if err != nil {
		return nil, "", err
	}
	return reopened, container, nil
}

// ContainerFromFilename infers a container from a filename's extension. Useful
// for backfill paths where the magic bytes aren't readily available.
func ContainerFromFilename(name string) Container {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp4", ".m4v", ".mov", ".3gp", ".3g2":
		return ContainerMP4
	case ".webm", ".mkv":
		return ContainerWebM
	case ".avi":
		return ContainerAVI
	case ".flv":
		return ContainerFLV
	case ".wmv", ".asf":
		return ContainerASF
	case ".ogv", ".ogg":
		return ContainerOgg
	case ".mpg", ".mpeg":
		return ContainerMPEGPS
	}
	return ""
}

// ImageFormat identifies a supported user-uploaded thumbnail format. We accept
// JPEG, PNG, and WebP because every modern browser produces at least one of
// them from <canvas> exports.
type ImageFormat string

const (
	ImageJPEG ImageFormat = "jpeg"
	ImagePNG  ImageFormat = "png"
	ImageWebP ImageFormat = "webp"
)

func (f ImageFormat) Ext() string {
	switch f {
	case ImageJPEG:
		return ".jpg"
	case ImagePNG:
		return ".png"
	case ImageWebP:
		return ".webp"
	default:
		return ""
	}
}

// DetectImage returns the format if head's leading bytes match a supported
// image container, or empty string otherwise. Uses magic bytes only — no
// trust placed in Content-Type from the client.
func DetectImage(head []byte) ImageFormat {
	switch {
	case len(head) >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF:
		return ImageJPEG
	case len(head) >= 8 && bytes.Equal(head[:8], pngMagic):
		return ImagePNG
	case len(head) >= 12 && bytes.Equal(head[:4], []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WEBP")):
		return ImageWebP
	}
	return ""
}

var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// ValidateImageUpload checks an image upload's size and magic bytes, then
// returns the opened reader (positioned at offset 0) and the detected format.
// Caller takes ownership of the returned file and MUST close it.
func ValidateImageUpload(fh *multipart.FileHeader, maxBytes int64) (multipart.File, ImageFormat, error) {
	if maxBytes <= 0 {
		maxBytes = MaxThumbnailBytes
	}
	if fh.Size > maxBytes {
		return nil, "", ErrFileTooLarge
	}
	src, err := fh.Open()
	if err != nil {
		return nil, "", err
	}
	head := make([]byte, 16)
	n, err := io.ReadFull(src, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		_ = src.Close()
		return nil, "", err
	}
	format := DetectImage(head[:n])
	if format == "" {
		_ = src.Close()
		return nil, "", ErrUnsupportedType
	}
	if seeker, ok := src.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			_ = src.Close()
			return nil, "", err
		}
		return src, format, nil
	}
	_ = src.Close()
	reopened, err := fh.Open()
	if err != nil {
		return nil, "", err
	}
	return reopened, format, nil
}
