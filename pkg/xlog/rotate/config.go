// Package rotate provides a concurrency-safe, daily rotating file writer.
package rotate

// Config controls rotation behavior.
type Config struct {
	// Filename is the active log file path.
	Filename string

	// MaxAge keeps at most N days of backups.
	// Zero disables cleanup.
	MaxAge int
}
