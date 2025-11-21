package logx

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/b-jonathan/taco/internal/logx"
	"github.com/b-jonathan/taco/internal/prompt"
)

func Init() error {
	log.SetPrefix("[taco] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}

func Time(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	dur := time.Since(start)
	prompt.TermLock.Lock()
	defer prompt.TermLock.Unlock()
	if err != nil {
		logx.Infof("%s failed in %s: %v", name, dur, err)
		return err
	}
	logx.Infof("%s finished in %s", name, dur)
	return nil
}

func Infof(format string, args ...any) {
	log.Printf("[INFO] "+format+"", args...)
}

func Warnf(format string, args ...any) {
	log.Printf("[WARN] "+format, args...)
}

func Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[Error] "+format+"\n", args...)
}
