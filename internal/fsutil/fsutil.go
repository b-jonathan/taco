package fsutil

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

// TODO: Some of these were completely vibe coded, just need to refactor a bit to make more consistent
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

func WriteFile(file FileInfo) error {
	filename := filepath.Base(file.Path)
	// log.Printf("Ensuring file: %s", path)
	if err := EnsureFile(file.Path); err != nil {
		return fmt.Errorf("ensure %s file: %w", filename, err)
	}
	// log.Println("Ensuring file complete")
	// log.Printf("Writing File: %s", path)
	if err := os.WriteFile(file.Path, file.Content, 0o644); err != nil {
		return fmt.Errorf("write %s file: %w", filename, err)
	}
	// log.Println("Writing file complete")
	return nil
}

func WriteMultipleFiles(files []FileInfo) error {
	for _, file := range files {
		if err := WriteFile(file); err != nil {
			return fmt.Errorf("write file %s: %w", file.Path, err)
		}
	}
	return nil
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

func RenderTemplate(tmplPath string) ([]byte, error) {
	tmplPath = filepath.Join("internal", "stacks", "templates", tmplPath)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", tmplPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", tmplPath, err)
	}
	return buf.Bytes(), nil
}
