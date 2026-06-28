package useragent

import ua "github.com/mssola/useragent"

// Info holds parsed User-Agent information.
type Info struct {
	OS      string
	Browser string
}

// Parse extracts OS and browser from a User-Agent string.
func Parse(userAgent string) Info {
	if userAgent == "" {
		return Info{}
	}
	p := ua.New(userAgent)
	browser, _ := p.Browser()
	return Info{
		OS:      p.OS(),
		Browser: browser,
	}
}
