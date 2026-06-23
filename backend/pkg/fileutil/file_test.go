package fileutil

import "testing"

func TestIsMP4_RecognisesAnyFtypBrand(t *testing.T) {
	brands := []string{"isom", "iso2", "mp41", "mp42", "avc1", "dash", "qt  ", "3gp4", "3g2a", "heic", "MSNV"}
	for _, brand := range brands {
		head := make([]byte, 16)
		head[0], head[1], head[2], head[3] = 0, 0, 0, 0x20
		copy(head[4:8], []byte("ftyp"))
		copy(head[8:12], []byte(brand))
		if !IsMP4(head) {
			t.Errorf("brand %q should be recognised as ISO base media", brand)
		}
	}
}

func TestIsMP4_RejectsNonMP4(t *testing.T) {
	cases := map[string][]byte{
		"too-short": []byte("xx"),
		"no-ftyp":   {0, 0, 0, 0x20, 'm', 'd', 'a', 't', 'i', 's', 'o', 'm'},
		"empty":     {},
	}
	for name, head := range cases {
		if IsMP4(head) {
			t.Errorf("%s should not be recognised as MP4", name)
		}
	}
}

func TestIsWebM(t *testing.T) {
	good := []byte{0x1A, 0x45, 0xDF, 0xA3, 0x9F, 0x42, 0x86}
	if !IsWebM(good) {
		t.Errorf("EBML header should be recognised as WebM")
	}
	bad := map[string][]byte{
		"too-short":   {0x1A, 0x45},
		"wrong-magic": {0x1A, 0x45, 0xDE, 0xA3, 0x00, 0x00},
		"all-zero":    make([]byte, 8),
	}
	for name, head := range bad {
		if IsWebM(head) {
			t.Errorf("%s should not be recognised as WebM", name)
		}
	}
}

func TestDetectContainer(t *testing.T) {
	cases := map[string]struct {
		head []byte
		want Container
	}{
		"mp4 isom": ftypHead("isom"),
		"mp4 mov":  ftypHead("qt  "),
		"mp4 3gp":  ftypHead("3gp4"),
		"webm":     payload([]byte{0x1A, 0x45, 0xDF, 0xA3, 0x9F, 0x42}, ContainerWebM),
		"avi":      payload(append(append([]byte("RIFF"), 0x10, 0x00, 0x00, 0x00), []byte("AVI ")...), ContainerAVI),
		"flv":      payload([]byte{'F', 'L', 'V', 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}, ContainerFLV),
		"asf":      payload(append([]byte{0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C}), ContainerASF),
		"ogg":      payload([]byte{'O', 'g', 'g', 'S', 0x00, 0x02}, ContainerOgg),
		"mpeg-ps":  payload([]byte{0x00, 0x00, 0x01, 0xBA, 0x44, 0x00, 0x04, 0x00}, ContainerMPEGPS),
	}
	for name, tc := range cases {
		if got := DetectContainer(tc.head); got != tc.want {
			t.Errorf("%s: got %q want %q", name, got, tc.want)
		}
	}

	rejects := map[string][]byte{
		"empty":   {},
		"text":    []byte("hello world hello world!"),
		"png":     {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		"riff-wave": append(append([]byte("RIFF"), 0, 0, 0, 0), []byte("WAVE")...), // RIFF but not AVI
	}
	for name, head := range rejects {
		if got := DetectContainer(head); got != "" {
			t.Errorf("%s should not match any container, got %q", name, got)
		}
	}
}

func ftypHead(brand string) struct {
	head []byte
	want Container
} {
	head := make([]byte, 16)
	head[0], head[1], head[2], head[3] = 0, 0, 0, 0x20
	copy(head[4:8], []byte("ftyp"))
	copy(head[8:12], []byte(brand))
	return struct {
		head []byte
		want Container
	}{head: head, want: ContainerMP4}
}

func payload(head []byte, want Container) struct {
	head []byte
	want Container
} {
	return struct {
		head []byte
		want Container
	}{head: head, want: want}
}

func TestContainerExt(t *testing.T) {
	cases := map[Container]string{
		ContainerMP4:  ".mp4",
		ContainerWebM: ".webm",
		Container(""): ".bin",
	}
	for c, want := range cases {
		if got := c.Ext(); got != want {
			t.Errorf("%q.Ext() = %q, want %q", c, got, want)
		}
	}
}

func TestDetectImage(t *testing.T) {
	cases := map[string]struct {
		head []byte
		want ImageFormat
	}{
		"jpeg": {[]byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0}, ImageJPEG},
		"png":  {[]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0, 0}, ImagePNG},
		"webp": {append(append([]byte("RIFF"), 0, 0, 0, 0), []byte("WEBPVP8")...), ImageWebP},
		"empty":     {[]byte{}, ""},
		"text":      {[]byte("hello world"), ""},
		"gif":       {[]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, ImageGIF},
		"short-jpg": {[]byte{0xFF, 0xD8}, ""},
		"avif":      {avifHead("avif"), ImageAVIF},
		"avis":      {avifHead("avis"), ImageAVIF},
		"avif-short": {[]byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p'}, ""},
	}
	for name, tc := range cases {
		if got := DetectImage(tc.head); got != tc.want {
			t.Errorf("%s: got %q want %q", name, got, tc.want)
		}
	}
}

func avifHead(brand string) []byte {
	head := make([]byte, 16)
	head[0], head[1], head[2], head[3] = 0x00, 0x00, 0x00, 0x20
	copy(head[4:8], []byte("ftyp"))
	copy(head[8:12], []byte(brand))
	return head
}
