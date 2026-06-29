package i18n

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func ResolveLocale(c echo.Context) string {
	if lang := c.QueryParam("lang"); lang != "" {
		return normalize(lang)
	}
	if accept := c.Request().Header.Get("Accept-Language"); accept != "" {
		lang := parseAcceptLanguage(accept)
		if lang != "" {
			return normalize(lang)
		}
	}
	return "en"
}

func normalize(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch {
	case lang == "zh" || strings.HasPrefix(lang, "zh-"):
		return "zh"
	default:
		return "en"
	}
}

func parseAcceptLanguage(header string) string {
	parts := strings.Split(header, ",")
	for _, part := range parts {
		lang := strings.TrimSpace(part)
		if idx := strings.IndexByte(lang, ';'); idx >= 0 {
			lang = strings.TrimSpace(lang[:idx])
		}
		if lang != "" {
			return lang
		}
	}
	return ""
}
