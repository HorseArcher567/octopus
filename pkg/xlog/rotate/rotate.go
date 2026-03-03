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

// dateFormat is used in rotated backup filenames.
const dateFormat = "2006-01-02"
const defaultExt = ".log"

// Writer serializes all writes behind a mutex and rotates files at day boundary.
//
// Rotation strategy:
//   - active file:   <basename><ext>
//   - backup file:   <basename>-YYYY-MM-DD<ext>
//   - if backup exists, content is appended
type Writer struct {
	config Config
	file   *os.File
	mu     sync.Mutex

	// Current active date.
	curYear  int
	curMonth time.Month
	curDay   int

	// Parsed from Config.Filename.
	basename string // Full path without extension.
	ext      string // Extension including the dot. Default is ".log".
}

// New validates config and returns a writer/closer pair.
func New(config Config) (io.Writer, io.Closer, error) {
	if config.Filename == "" {
		return nil, nil, fmt.Errorf("filename is required")
	}

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

// MustNew is like New but panics on error.
func MustNew(config Config) (io.Writer, io.Closer) {
	writer, closer, err := New(config)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize rotate writer: %v", err))
	}
	return writer, closer
}

// Write writes p to the active file and rotates when the local date changes.
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openFile(); err != nil {
			return 0, err
		}
	}

	now := time.Now()
	if now.Day() != w.curDay || now.Month() != w.curMonth || now.Year() != w.curYear {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

// Close closes the active file.
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

// init opens or rotates the target file based on its modification date.
func (w *Writer) init() error {
	if err := os.MkdirAll(filepath.Dir(w.config.Filename), 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	info, err := os.Stat(w.config.Filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat log file: %w", err)
		}
		return w.openFile()
	}

	modTime := info.ModTime()
	now := time.Now()

	if modTime.Year() != now.Year() || modTime.Month() != now.Month() || modTime.Day() != now.Day() {
		w.curYear, w.curMonth, w.curDay = modTime.Date()
		return w.rotate()
	}

	return w.openFile()
}

// openFile opens the active file in append mode and updates current date.
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

// rotate moves the active file to a dated backup and opens a fresh active file.
func (w *Writer) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}

	backupName := w.backupName(w.curYear, w.curMonth, w.curDay)

	if _, err := os.Stat(backupName); err == nil {
		if err := appendFile(w.config.Filename, backupName); err != nil {
			return fmt.Errorf("failed to append log file: %w", err)
		}
		if err := os.Remove(w.config.Filename); err != nil {
			return fmt.Errorf("failed to remove rotated file: %w", err)
		}
	} else {
		if err := os.Rename(w.config.Filename, backupName); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to rename log file: %w", err)
			}
		}
	}

	if err := w.openFile(); err != nil {
		return err
	}

	if w.config.MaxAge > 0 {
		go w.cleanup()
	}

	return nil
}

// backupName builds the rotated backup name for a specific date.
func (w *Writer) backupName(year int, month time.Month, day int) string {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	return fmt.Sprintf("%s-%s%s", w.basename, date.Format(dateFormat), w.ext)
}

// cleanup removes expired backup files best-effort.
func (w *Writer) cleanup() {
	dir := filepath.Dir(w.config.Filename)
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	now := time.Now()
	cutoff := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, -w.config.MaxAge)
	baseNameOnly := filepath.Base(w.basename)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		name := f.Name()

		fileDate, ok := w.parseBackupDate(name, baseNameOnly)
		if !ok {
			continue
		}

		if fileDate.Before(cutoff) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

// parseBackupDate parses backup date from filename.
// Expected format: {basename}-{date}{ext}.
func (w *Writer) parseBackupDate(filename, baseName string) (time.Time, bool) {
	if !strings.HasPrefix(filename, baseName+"-") {
		return time.Time{}, false
	}

	if !strings.HasSuffix(filename, w.ext) {
		return time.Time{}, false
	}

	datePart := filename[len(baseName)+1 : len(filename)-len(w.ext)]

	date, err := time.Parse(dateFormat, datePart)
	if err != nil {
		return time.Time{}, false
	}

	return date, true
}

// normalizeFilename appends the default extension when missing.
func normalizeFilename(filename string) string {
	if filepath.Ext(filename) == "" {
		return filename + defaultExt
	}
	return filename
}

// splitFilename returns basename and extension parts.
func splitFilename(filename string) (basename, ext string) {
	ext = filepath.Ext(filename)
	basename = filename[:len(filename)-len(ext)]
	return basename, ext
}

// appendFile appends src file content into dst.
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
