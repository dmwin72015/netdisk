package i18n

// Language represents a supported language
type Language string

const (
	LangZH Language = "zh"
	LangEN Language = "en"
)

// Message key type for error messages
type MessageKey string

const (
	MsgNotFound             MessageKey = "not_found"
	MsgUnauthorized         MessageKey = "unauthorized"
	MsgForbidden            MessageKey = "forbidden"
	MsgInvalidInput         MessageKey = "invalid_input"
	MsgAlreadyExists        MessageKey = "already_exists"
	MsgNameConflict         MessageKey = "name_conflict"
	MsgDuplicateFile        MessageKey = "duplicate_file"
	MsgSameFileConflict     MessageKey = "same_file_conflict"
	MsgFileTooLarge         MessageKey = "file_too_large"
	MsgUnsupportedType      MessageKey = "unsupported_file_type"
	MsgQuotaExceeded        MessageKey = "quota_exceeded"
	MsgChallengeExpired     MessageKey = "challenge_expired"
	MsgChallengeMismatch    MessageKey = "challenge_mismatch"
	MsgDirNotEmpty          MessageKey = "directory_not_empty"
	MsgFileRequired         MessageKey = "file_required"
	MsgUnsupportedImage     MessageKey = "unsupported_image"
	MsgSystemFileLocked     MessageKey = "system_file_locked"
	MsgDirectoryLocked      MessageKey = "directory_locked"
	MsgInternal             MessageKey = "internal_error"
	MsgFileTooLargeUpload   MessageKey = "file_exceeds_size_limit_upload"
	MsgTotalSizeExceeded    MessageKey = "total_size_exceeded"
	MsgInvalidPath          MessageKey = "invalid_path"
	MsgParentNotFound       MessageKey = "parent_not_found"
	MsgInvalidSortField     MessageKey = "invalid_sort_field"
	MsgInvalidPage          MessageKey = "invalid_page"
	MsgTokenExpired         MessageKey = "token_expired"
	MsgInvalidToken         MessageKey = "invalid_token"
	MsgInvalidCredentials   MessageKey = "invalid_credentials"
	MsgUserDisabled         MessageKey = "user_disabled"
	MsgEmailAlreadyExists   MessageKey = "email_already_exists"
	MsgUsernameExists       MessageKey = "username_exists"
	MsgInvalidRefreshToken  MessageKey = "invalid_refresh_token"
	MsgInvalidVerifyCode    MessageKey = "invalid_verify_code"
	MsgCodeExpired          MessageKey = "code_expired"
	MsgSendCodeTooFrequent  MessageKey = "send_code_too_frequent"
	MsgWeakPassword         MessageKey = "weak_password"
	MsgInvalidOAuthState    MessageKey = "invalid_oauth_state"
	MsgOAuthBindFailed      MessageKey = "oauth_bind_failed"
	MsgOAuthUnbindFailed    MessageKey = "oauth_unbind_failed"
	MsgStorageFull          MessageKey = "storage_full"
	MsgMediaNotReady        MessageKey = "media_not_ready"
	MsgMediaProcessing      MessageKey = "media_processing"
	MsgAlbumNotFound        MessageKey = "album_not_found"
	MsgAlbumNameRequired    MessageKey = "album_name_required"
	MsgPhotoNotFound        MessageKey = "photo_not_found"
	MsgAdminOnly            MessageKey = "admin_only"
	MsgShareNotFound        MessageKey = "share_not_found"
	MsgShareExpired         MessageKey = "share_expired"
	MsgShareCodeExists      MessageKey = "share_code_exists"
	MsgInvalidShareCode     MessageKey = "invalid_share_code"
	MsgTaskNotFound         MessageKey = "task_not_found"
)

var messages = map[MessageKey]map[Language]string{
	MsgNotFound: {
		LangZH: "资源未找到",
		LangEN: "not found",
	},
	MsgUnauthorized: {
		LangZH: "未授权，请先登录",
		LangEN: "unauthorized",
	},
	MsgForbidden: {
		LangZH: "权限不足",
		LangEN: "forbidden",
	},
	MsgInvalidInput: {
		LangZH: "输入无效",
		LangEN: "invalid input",
	},
	MsgAlreadyExists: {
		LangZH: "已存在",
		LangEN: "already exists",
	},
	MsgNameConflict: {
		LangZH: "名称冲突",
		LangEN: "name conflict",
	},
	MsgDuplicateFile: {
		LangZH: "重复文件",
		LangEN: "duplicate file",
	},
	MsgSameFileConflict: {
		LangZH: "文件冲突",
		LangEN: "same file conflict",
	},
	MsgFileTooLarge: {
		LangZH: "文件超过大小限制",
		LangEN: "file exceeds size limit",
	},
	MsgUnsupportedType: {
		LangZH: "不支持的文件类型",
		LangEN: "unsupported file type",
	},
	MsgQuotaExceeded: {
		LangZH: "存储空间不足",
		LangEN: "storage quota exceeded",
	},
	MsgChallengeExpired: {
		LangZH: "挑战已过期",
		LangEN: "challenge expired",
	},
	MsgChallengeMismatch: {
		LangZH: "挑战不匹配",
		LangEN: "challenge mismatch",
	},
	MsgDirNotEmpty: {
		LangZH: "目录不为空",
		LangEN: "directory is not empty",
	},
	MsgFileRequired: {
		LangZH: "文件不能为空",
		LangEN: "file is required",
	},
	MsgUnsupportedImage: {
		LangZH: "仅支持 JPEG、PNG 和 WebP 格式",
		LangEN: "only JPEG, PNG and WebP are supported",
	},
	MsgSystemFileLocked: {
		LangZH: "系统文件无法修改",
		LangEN: "system file cannot be modified",
	},
	MsgDirectoryLocked: {
		LangZH: "目录已锁定",
		LangEN: "directory is locked",
	},
	MsgInternal: {
		LangZH: "内部错误",
		LangEN: "internal error",
	},
	MsgFileTooLargeUpload: {
		LangZH: "文件超过上传大小限制",
		LangEN: "file exceeds upload size limit",
	},
	MsgTotalSizeExceeded: {
		LangZH: "总大小超过限制",
		LangEN: "total size exceeded",
	},
	MsgInvalidPath: {
		LangZH: "路径无效",
		LangEN: "invalid path",
	},
	MsgParentNotFound: {
		LangZH: "父目录不存在",
		LangEN: "parent not found",
	},
	MsgInvalidSortField: {
		LangZH: "无效的排序字段",
		LangEN: "invalid sort field",
	},
	MsgInvalidPage: {
		LangZH: "无效的页码",
		LangEN: "invalid page",
	},
	MsgTokenExpired: {
		LangZH: "令牌已过期",
		LangEN: "token expired",
	},
	MsgInvalidToken: {
		LangZH: "令牌无效",
		LangEN: "invalid token",
	},
	MsgInvalidCredentials: {
		LangZH: "用户名或密码错误",
		LangEN: "invalid credentials",
	},
	MsgUserDisabled: {
		LangZH: "用户已被禁用",
		LangEN: "user disabled",
	},
	MsgEmailAlreadyExists: {
		LangZH: "邮箱已被注册",
		LangEN: "email already exists",
	},
	MsgUsernameExists: {
		LangZH: "用户名已存在",
		LangEN: "username already exists",
	},
	MsgInvalidRefreshToken: {
		LangZH: "刷新令牌无效",
		LangEN: "invalid refresh token",
	},
	MsgInvalidVerifyCode: {
		LangZH: "验证码无效",
		LangEN: "invalid verify code",
	},
	MsgCodeExpired: {
		LangZH: "验证码已过期",
		LangEN: "code expired",
	},
	MsgSendCodeTooFrequent: {
		LangZH: "发送过于频繁，请稍后再试",
		LangEN: "send code too frequent, please try again later",
	},
	MsgWeakPassword: {
		LangZH: "密码强度不足",
		LangEN: "weak password",
	},
	MsgInvalidOAuthState: {
		LangZH: "OAuth 状态无效",
		LangEN: "invalid OAuth state",
	},
	MsgOAuthBindFailed: {
		LangZH: "OAuth 绑定失败",
		LangEN: "OAuth bind failed",
	},
	MsgOAuthUnbindFailed: {
		LangZH: "OAuth 解绑失败",
		LangEN: "OAuth unbind failed",
	},
	MsgStorageFull: {
		LangZH: "存储空间已满",
		LangEN: "storage full",
	},
	MsgMediaNotReady: {
		LangZH: "媒体未就绪",
		LangEN: "media not ready",
	},
	MsgMediaProcessing: {
		LangZH: "媒体正在处理中",
		LangEN: "media is processing",
	},
	MsgAlbumNotFound: {
		LangZH: "相册未找到",
		LangEN: "album not found",
	},
	MsgAlbumNameRequired: {
		LangZH: "相册名称不能为空",
		LangEN: "album name is required",
	},
	MsgPhotoNotFound: {
		LangZH: "照片未找到",
		LangEN: "photo not found",
	},
	MsgAdminOnly: {
		LangZH: "仅管理员可访问",
		LangEN: "admin only",
	},
	MsgShareNotFound: {
		LangZH: "分享未找到",
		LangEN: "share not found",
	},
	MsgShareExpired: {
		LangZH: "分享已过期",
		LangEN: "share expired",
	},
	MsgShareCodeExists: {
		LangZH: "分享码已存在",
		LangEN: "share code already exists",
	},
	MsgInvalidShareCode: {
		LangZH: "分享码无效",
		LangEN: "invalid share code",
	},
	MsgTaskNotFound: {
		LangZH: "任务未找到",
		LangEN: "task not found",
	},
}

// T returns the localized message for the given key and language.
// Falls back to English if the language or key is not found.
func T(key MessageKey, lang Language) string {
	if m, ok := messages[key]; ok {
		if msg, ok := m[lang]; ok {
			return msg
		}
		if msg, ok := m[LangEN]; ok {
			return msg
		}
	}
	return string(key)
}

// DetectLanguage reads the Accept-Language header and returns
// the best matching supported language. Defaults to English.
func DetectLanguage(acceptLanguage string) Language {
	if acceptLanguage == "" {
		return LangEN
	}
	// Simple detection: check if Chinese is preferred
	for _, part := range []string{"zh", "zh-CN", "zh-TW", "zh-HK"} {
		if contains(acceptLanguage, part) {
			return LangZH
		}
	}
	return LangEN
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
