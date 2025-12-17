package rotate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	_, closer, err := New(Config{
		Filename: filename,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer closer.Close()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestWrite(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	writer, closer, err := New(Config{
		Filename: filename,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer closer.Close()

	message := "test log message\n"
	n, err := writer.Write([]byte(message))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	if n != len(message) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(message), n)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if string(content) != message {
		t.Errorf("Expected content %q, got %q", message, string(content))
	}
}

func TestConcurrentWrite(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	writer, closer, err := New(Config{
		Filename: filename,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer closer.Close()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			message := fmt.Sprintf("test message %d\n", n)
			for j := 0; j < 100; j++ {
				writer.Write([]byte(message))
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Log file should exist after concurrent writes")
	}
}

func TestRotationOnNew(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	// 1. 创建一个"旧"日志文件
	oldContent := "old log data"
	if err := os.WriteFile(filename, []byte(oldContent), 0o666); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}

	// 修改文件时间为昨天
	yesterday := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(filename, yesterday, yesterday); err != nil {
		t.Fatalf("Failed to modify file time: %v", err)
	}

	// 2. 创建 Writer，应该触发轮转
	writer, closer, err := New(Config{
		Filename: filename,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer closer.Close()

	// 3. 验证新文件已创建且为空（或只有新内容）
	newContent := "new log data"
	writer.Write([]byte(newContent))

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if string(content) != newContent {
		t.Errorf("Expected new content %q, got %q", newContent, string(content))
	}

	// 4. 验证旧文件已备份（新格式：test-2023-12-08.log）
	backupName := filepath.Join(tempDir, fmt.Sprintf("test-%s.log", yesterday.Format("2006-01-02")))
	backupContent, err := os.ReadFile(backupName)
	if err != nil {
		t.Fatalf("Failed to read backup file %s: %v", backupName, err)
	}
	if string(backupContent) != oldContent {
		t.Errorf("Expected backup content %q, got %q", oldContent, string(backupContent))
	}
}

func TestCleanup(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	// 1. 创建过期备份文件（新格式：test-{date}.log）
	oldDate := time.Now().AddDate(0, 0, -3)
	oldBackupName := filepath.Join(tempDir, fmt.Sprintf("test-%s.log", oldDate.Format("2006-01-02")))
	if err := os.WriteFile(oldBackupName, []byte("old backup"), 0o666); err != nil {
		t.Fatalf("Failed to create old backup: %v", err)
	}

	// 2. 创建最近备份文件（不应被删除）
	recentDate := time.Now().AddDate(0, 0, -1)
	recentBackupName := filepath.Join(tempDir, fmt.Sprintf("test-%s.log", recentDate.Format("2006-01-02")))
	if err := os.WriteFile(recentBackupName, []byte("recent backup"), 0o666); err != nil {
		t.Fatalf("Failed to create recent backup: %v", err)
	}

	// 3. 创建当前日志文件（旧的，触发轮转）
	if err := os.WriteFile(filename, []byte("current old"), 0o666); err != nil {
		t.Fatalf("Failed to create current log: %v", err)
	}
	yesterday := time.Now().Add(-24 * time.Hour)
	os.Chtimes(filename, yesterday, yesterday)

	writer, closer, err := New(Config{
		Filename: filename,
		MaxAge:   1,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer closer.Close()
	writer.Write([]byte("trigger"))

	// 等待异步清理
	time.Sleep(200 * time.Millisecond)

	// 4. 验证
	if _, err := os.Stat(oldBackupName); !os.IsNotExist(err) {
		t.Error("Old backup file should have been deleted")
	}

	if _, err := os.Stat(recentBackupName); os.IsNotExist(err) {
		t.Error("Recent backup file should NOT have been deleted")
	}
}

func TestMustNew(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	writer, closer := MustNew(Config{
		Filename: filename,
	})
	defer closer.Close()

	if writer == nil {
		t.Error("MustNew() returned nil writer")
	}
}

func TestMustNewPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew() should panic with empty filename")
		}
	}()

	MustNew(Config{})
}
