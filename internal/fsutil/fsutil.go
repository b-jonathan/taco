package fsutil

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/b-jonathan/taco/internal/stacks/templates"
	"github.com/spf13/afero"
)

var Fs = afero.NewOsFs()

// TODO: Some of these were completely vibe coded, just need to refactor a bit to make more consistent
func EnsureFile(path string) error {
	// Create parent directories if needed.
	if err := Fs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Create the file if missing. O_EXCL prevents clobbering if a race happens.
	f, err := Fs.OpenFile(path, os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		// If it already exists, that's fine.
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
	if err := afero.WriteFile(Fs, file.Path, file.Content, 0o644); err != nil {
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
	buf, _ := afero.ReadFile(Fs, path)
	for _, line := range lines {
		if !bytes.Contains(buf, []byte(line+"\n")) && !bytes.Equal(bytes.TrimSpace(buf), []byte(line)) {
			if len(buf) > 0 && buf[len(buf)-1] != '\n' {
				buf = append(buf, '\n')
			}
			buf = append(buf, []byte(line+"\n")...)
		}
	}
	return afero.WriteFile(Fs, path, buf, 0o644)
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
	if stack == "" || dependency == "" {
		return false
	}

	info, err := templates.FS.ReadDir(stack)
	if err != nil {
		fmt.Printf("embedded path does not exist %s\n", stack)
		return false
	}
	//check if subfolder is "src", "db" or doesn't exist -> true.
	hasFolder := false
	for _, e := range info {
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
	dependencyPath := path.Join(stack, dependency)
	if _, err := templates.FS.ReadDir(dependencyPath); err == nil {
		return true
	}
	return false

}

func GenerateFromTemplateDir(templateRoot, outputRoot string) error {
	return fs.WalkDir(templates.FS, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		relPath, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}

		outputRel := strings.TrimSuffix(relPath, ".tmpl")
		finalPath := filepath.Join(outputRoot, outputRel)

		content, err := RenderTemplate(path)
		if err != nil {
			return fmt.Errorf("render template %s: %w", path, err)
		}

		if err := Fs.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
			return err
		}

		return afero.WriteFile(Fs, finalPath, content, 0644)
	})
}
