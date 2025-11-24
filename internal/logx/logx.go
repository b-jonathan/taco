package logx

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/b-jonathan/taco/internal/prompt"
)

func Init() error {
	log.SetPrefix("[taco] ")
	log.SetFlags(log.Lshortfile)
	return nil
}

func caller(skip int) (file string, line int) {
	_, f, l, ok := runtime.Caller(skip)
	if !ok {
		return "???", 0
	}
	return filepath.Base(f), l
}
func Time(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	dur := time.Since(start)
	prompt.TermLock.Lock()
	defer prompt.TermLock.Unlock()
	if err != nil {
		Infof("%s failed in %s: %v", name, dur, err)
		return err
	}
	Infof("%s finished in %s", name, dur)
	return nil
}

func Infof(format string, args ...any) {
	file, line := caller(5)
	fullFormat := "[%s:%d] [INFO] " + format + "\n"
	fullArgs := append([]any{file, line}, args...)
	fmt.Printf(fullFormat, fullArgs...)

}

func Warnf(format string, args ...any) {
	file, line := caller(5)
	fullFormat := "[%s:%d] [WARN] " + format + "\n"
	fullArgs := append([]any{file, line}, args...)
	fmt.Printf(fullFormat, fullArgs...)
}

func Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[Error] "+format+"\n", args...)
}
