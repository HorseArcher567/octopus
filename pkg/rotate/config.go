package rotate

// Config 日志轮转配置
type Config struct {
	// Filename 日志文件路径（必填）
	Filename string

	// MaxAge 保留旧日志文件的最大天数，0 表示不删除
	MaxAge int
}
