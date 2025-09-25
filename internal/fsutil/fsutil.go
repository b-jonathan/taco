package fsutil

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
)

func EnsureFile(path string) error {
	// Create parent directories if needed.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Create the file if missing. O_EXCL prevents clobbering if a race happens.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		// If it already exists, that’s fine.
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	return f.Close()
}

func AppendUniqueLines(path string, lines []string) error {
	buf, _ := os.ReadFile(path)
	for _, line := range lines {
		if !bytes.Contains(buf, []byte(line+"\n")) && !bytes.Equal(bytes.TrimSpace(buf), []byte(line)) {
			if len(buf) > 0 && buf[len(buf)-1] != '\n' {
				buf = append(buf, '\n')
			}
			buf = append(buf, []byte(line+"\n")...)
		}
	}
	return os.WriteFile(path, buf, 0o644)
}

// in a shared package or file
var fileLocks sync.Map

func WithFileLock(path string, fn func() error) error {
	abs, _ := filepath.Abs(path)
	v, _ := fileLocks.LoadOrStore(abs, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	return fn()
}
