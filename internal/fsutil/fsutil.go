package fsutil

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/b-jonathan/taco/internal/stacks/templates"
	"github.com/spf13/afero"
)

// TODO: Some of these were completely vibe coded, just need to refactor a bit to make more consistent
func EnsureFile(fsys afero.Fs, path string) error {
	// Create parent directories if needed using the injected FS.
	if err := fsys.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Create the file if missing. O_EXCL prevents clobbering if a race happens.
	f, err := fsys.OpenFile(path, os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		// If it already exists, that’s fine.
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	return f.Close()
}

func WriteFile(fsys afero.Fs, file FileInfo) error {
	filename := filepath.Base(file.Path)
	// log.Printf("Ensuring file: %s", path)
	if err := EnsureFile(fsys, file.Path); err != nil {
		return fmt.Errorf("ensure %s file: %w", filename, err)
	}
	// log.Println("Ensuring file complete")
	// log.Printf("Writing File: %s", path)
	if err := afero.WriteFile(fsys, file.Path, file.Content, 0o644); err != nil {
		return fmt.Errorf("write %s file: %w", filename, err)
	}
	// log.Println("Writing file complete")
	return nil
}

func WriteMultipleFiles(fsys afero.Fs, files []FileInfo) error {
	for _, file := range files {
		if err := WriteFile(fsys, file); err != nil {
			return fmt.Errorf("write file %s: %w", file.Path, err)
		}
	}
	return nil
}

func AppendUniqueLines(fsys afero.Fs, path string, lines []string) error {
	buf, _ := afero.ReadFile(fsys, path)
	for _, line := range lines {
		if !bytes.Contains(buf, []byte(line+"\n")) && !bytes.Equal(bytes.TrimSpace(buf), []byte(line)) {
			if len(buf) > 0 && buf[len(buf)-1] != '\n' {
				buf = append(buf, '\n')
			}
			buf = append(buf, []byte(line+"\n")...)
		}
	}
	return afero.WriteFile(fsys, path, buf, 0o644)
}

// GenerateFiles is a declarative helper that renders and writes multiple files.
// fileMap keys are destination paths (relative to root).
// fileMap values are template paths (inside embed.FS).
func GenerateFiles(fsys afero.Fs, root string, fileMap map[string]string) error {
	for destPath, tmplPath := range fileMap {
		content, err := RenderTemplate(tmplPath)
		if err != nil {
			return fmt.Errorf("render %s: %w", tmplPath, err)
		}

		fullPath := filepath.Join(root, destPath)
		if err := WriteFile(fsys, FileInfo{Path: fullPath, Content: content}); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}
	}
	return nil
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
	data, err := templates.FS.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("read embedded template %s: %w", tmplPath, err)
	}

	// Parse template from in-memory string
	tmpl, err := template.New(filepath.Base(tmplPath)).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", tmplPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", tmplPath, err)
	}
	return buf.Bytes(), nil
}

// stack is the parent tech, check if dependency is compatible with it.
func ValidateDependency(stack, dependency string) bool {
	if stack == "none" || dependency == "none" {
		return true
	}
	// FIX: Don't join "internal/stacks...".
	// Read relative to the root of the embedded FS.
	stackPath := stack

	// Use fs.ReadDir on the EMBEDDED filesystem (templates.FS), not os.ReadDir
	entries, err := fs.ReadDir(templates.FS, stackPath)
	if err != nil {
		// fmt.Printf("path does not exist in binary %s\n", stackPath)
		return false
	}
	//check if subfolder is "src", "db" or doesn't exist -> true.
	hasFolder := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name != "src" && name != "db" {
			hasFolder = true
			break
		}
	}
	if !hasFolder {
		return true
	}
	//else, check if dependency is in subfolders of stack.
	dependencyPath := filepath.Join(stackPath, dependency)
	// Use fs.Stat on the EMBEDDED filesystem
	if info, err := fs.Stat(templates.FS, dependencyPath); err == nil && info.IsDir() {
		return true
	}
	return false

}
