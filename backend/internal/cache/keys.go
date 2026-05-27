package cache

import "fmt"

func PreCacheKey(fileSize int64, preHash string) string {
	return fmt.Sprintf("nd:pre:%d:%s", fileSize, preHash)
}

func ChallengeKey(userID int64, fileHash string) string {
	return fmt.Sprintf("nd:challenge:%d:%s", userID, fileHash)
}

func ChunksKey(uploadSlug string) string {
	return fmt.Sprintf("nd:chunks:%s", uploadSlug)
}

func MergeLockKey(fileHash string) string {
	return fmt.Sprintf("nd:lock:merge:%s", fileHash)
}

func RateLimitKey(ip, action string) string {
	return fmt.Sprintf("nd:rate:%s:%s", ip, action)
}

func MediaProgressKey(transcodeSlug string) string {
	return fmt.Sprintf("nd:media:transcode:%s:progress", transcodeSlug)
}
