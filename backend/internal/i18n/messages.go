package i18n

var actionLabels = map[string]map[string]string{
	"en": {
		"user.login":           "Login",
		"user.register":        "Register",
		"user.logout":          "Logout",
		"user.oauth_login":     "OAuth Login",
		"user.password_change": "Password Change",
		"file.upload":          "File Upload",
		"file.delete":          "File Delete",
		"file.rename":          "File Rename",
		"file.move":            "File Move",
		"dir.create":           "Create Directory",
		"dir.lock":             "Lock Directory",
		"dir.unlock":           "Unlock Directory",
		"dir.unlock_failed":    "Unlock Failed",
		"share.create":         "Create Share Link",
		"share.delete":         "Delete Share Link",
		"admin.create_user":    "Create User",
		"admin.delete_user":    "Delete User",
		"admin.update_role":    "Update Role",
	},
	"zh": {
		"user.login":           "登录",
		"user.register":        "注册",
		"user.logout":          "退出登录",
		"user.oauth_login":     "OAuth 登录",
		"user.password_change": "修改密码",
		"file.upload":          "上传文件",
		"file.delete":          "删除文件",
		"file.rename":          "重命名文件",
		"file.move":            "移动文件",
		"dir.create":           "创建目录",
		"dir.lock":             "锁定目录",
		"dir.unlock":           "解锁目录",
		"dir.unlock_failed":    "解锁失败",
		"share.create":         "创建分享链接",
		"share.delete":         "删除分享链接",
		"admin.create_user":    "创建用户",
		"admin.delete_user":    "删除用户",
		"admin.update_role":    "更新角色",
	},
}

func ActionLabel(action, locale string) string {
	if msgs, ok := actionLabels[locale]; ok {
		if label, ok := msgs[action]; ok {
			return label
		}
	}
	if msgs, ok := actionLabels["en"]; ok {
		if label, ok := msgs[action]; ok {
			return label
		}
	}
	return action
}
