package rotate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 日期格式（固定）
const dateFormat = "2006-01-02"
const defaultExt = ".log"

// Writer 实现了 io.WriteCloser 接口，支持按天日志轮转
type Writer struct {
	config Config
	file   *os.File
	mu     sync.Mutex

	// 缓存当前日期，避免每次 Write 都比较字符串
	curYear  int
	curMonth time.Month
	curDay   int

	// 文件名的基础部分和扩展名
	basename string // 不含扩展名的文件名（包含路径）
	ext      string // 扩展名（包含点号，默认值为".log"）
}

// New 创建一个新的按天轮转写入器
func New(config Config) (io.Writer, io.Closer, error) {
	if config.Filename == "" {
		return nil, nil, fmt.Errorf("filename is required")
	}

	// 标准化文件名（确保有扩展名）
	config.Filename = normalizeFilename(config.Filename)

	w := &Writer{
		config: config,
	}
	w.basename, w.ext = splitFilename(w.config.Filename)

	if err := w.init(); err != nil {
		return nil, nil, err
	}

	return w, w, nil
}

// MustNew 创建一个新的按天轮转写入器（失败时 panic）
func MustNew(config Config) (io.Writer, io.Closer) {
	writer, closer, err := New(config)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize rotate writer: %v", err))
	}
	return writer, closer
}

// Write 实现 io.Writer 接口
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openFile(); err != nil {
			return 0, err
		}
	}

	// 检查是否跨天，需要轮转
	now := time.Now()
	if now.Day() != w.curDay || now.Month() != w.curMonth || now.Year() != w.curYear {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

// Close 实现 io.Closer 接口
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}

// init 初始化日志文件
func (w *Writer) init() error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(w.config.Filename), 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 检查现有文件
	info, err := os.Stat(w.config.Filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat log file: %w", err)
		}
		// 文件不存在，直接创建
		return w.openFile()
	}

	// 文件存在，检查最后修改时间
	modTime := info.ModTime()
	now := time.Now()

	if modTime.Year() != now.Year() || modTime.Month() != now.Month() || modTime.Day() != now.Day() {
		// 如果是以前的文件，立即轮转
		w.curYear, w.curMonth, w.curDay = modTime.Date()
		return w.rotate()
	}

	// 是当天的文件，直接打开
	return w.openFile()
}

// openFile 打开或创建当前日志文件
func (w *Writer) openFile() error {
	now := time.Now()
	w.curYear, w.curMonth, w.curDay = now.Date()

	file, err := os.OpenFile(w.config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	w.file = file
	return nil
}

// rotate 执行轮转
func (w *Writer) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}

	// 生成备份文件名：{basename}-{date}{ext}
	backupName := w.backupName(w.curYear, w.curMonth, w.curDay)

	// 重命名当前文件
	if _, err := os.Stat(backupName); err == nil {
		// 备份文件已存在，追加内容
		if err := appendFile(w.config.Filename, backupName); err != nil {
			return fmt.Errorf("failed to append log file: %w", err)
		}
		if err := os.Remove(w.config.Filename); err != nil {
			return fmt.Errorf("failed to remove rotated file: %w", err)
		}
	} else {
		// 正常重命名
		if err := os.Rename(w.config.Filename, backupName); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to rename log file: %w", err)
			}
		}
	}

	// 打开新文件
	if err := w.openFile(); err != nil {
		return err
	}

	// 异步清理过期文件
	if w.config.MaxAge > 0 {
		go w.cleanup()
	}

	return nil
}

// backupName 生成备份文件名：{basename}-{date}{ext}
// 例如：logs/app.log -> logs/app-2023-12-08.log
func (w *Writer) backupName(year int, month time.Month, day int) string {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	return fmt.Sprintf("%s-%s%s", w.basename, date.Format(dateFormat), w.ext)
}

// cleanup 清理过期文件
func (w *Writer) cleanup() {
	dir := filepath.Dir(w.config.Filename)
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// 计算截止日期（只比较日期，忽略时分秒）
	now := time.Now()
	cutoff := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, -w.config.MaxAge)
	baseNameOnly := filepath.Base(w.basename)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		name := f.Name()

		// 解析日期
		fileDate, ok := w.parseBackupDate(name, baseNameOnly)
		if !ok {
			continue
		}

		// 检查是否过期（严格小于 cutoff 才删除）
		if fileDate.Before(cutoff) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

// parseBackupDate 从备份文件名中解析日期
// 格式：{basename}-{date}{ext}
func (w *Writer) parseBackupDate(filename, baseName string) (time.Time, bool) {
	// 检查前缀
	if !strings.HasPrefix(filename, baseName+"-") {
		return time.Time{}, false
	}

	// 检查后缀
	if !strings.HasSuffix(filename, w.ext) {
		return time.Time{}, false
	}

	// 提取日期部分
	datePart := filename[len(baseName)+1 : len(filename)-len(w.ext)]

	// 解析日期
	date, err := time.Parse(dateFormat, datePart)
	if err != nil {
		return time.Time{}, false
	}

	return date, true
}

// normalizeFilename 标准化文件名，确保有扩展名
func normalizeFilename(filename string) string {
	if filepath.Ext(filename) == "" {
		return filename + defaultExt
	}
	return filename
}

// splitFilename 将文件名分割为基础部分和扩展名
// 例如："logs/app.log" -> ("logs/app", ".log")
func splitFilename(filename string) (basename, ext string) {
	ext = filepath.Ext(filename)
	basename = filename[:len(filename)-len(ext)]
	return basename, ext
}

// appendFile 将 src 文件内容追加到 dst 文件
func appendFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}
