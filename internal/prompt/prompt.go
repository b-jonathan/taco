package prompt

import (
	"fmt"
	"os"
	"sync"

	"github.com/AlecAivazis/survey/v2"
)

var TermLock sync.Mutex

// Lock and unlock helpers if needed elsewhere
func Lock()   { TermLock.Lock() }
func Unlock() { TermLock.Unlock() }

/* Helpers */
func IsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func CreateSurveyInput(message string, options AskOpts) (string, error) {
	if !IsTTY() && options.Default == nil {
		return "", fmt.Errorf("TTY required or provide a default")
	}

	prompt := &survey.Input{
		Message: message,
		Help:    options.Help,
	}

	if value, ok := options.Default.(string); ok {
		prompt.Default = value
	}

	return askOneString(prompt, options)
}

func CreateSurveySelect(message string, choices []string, options AskOpts) (string, error) {
	if !IsTTY() && options.Default == nil {
		return "", fmt.Errorf("TTY required or provide a default")
	}

	if len(choices) == 0 {
		return "", fmt.Errorf("options cannot be empty")
	}

	prompt := &survey.Select{
		Message:  message,
		Options:  choices, //choices, this is a bit weird AHHAHA
		Help:     options.Help,
		PageSize: options.PageSize,
	}

	if v, ok := options.Default.(string); ok {
		prompt.Default = v
	}
	return askOneString(prompt, options)
}

func CreateSurveyMultiSelect(message string, choices []string, options AskOpts) ([]string, error) {
	if !IsTTY() && options.Default == nil {
		return nil, fmt.Errorf("TTY required or provide a default")
	}

	if len(choices) == 0 {
		return nil, fmt.Errorf("options cannot be empty")
	}

	prompt := &survey.Select{
		Message:  message,
		Options:  choices, //choices, this is a bit weird AHHAHA
		Help:     options.Help,
		PageSize: options.PageSize,
	}

	if v, ok := options.Default.(string); ok {
		prompt.Default = v
	}
	return askManyString(prompt, options)
}

func CreateSurveyConfirm(message string, options AskOpts) (bool, error) {
	if !IsTTY() && options.Default == nil {
		return false, fmt.Errorf("TTY required or provide a default")
	}
	prompt := &survey.Confirm{
		Message: message,
	}
	if v, ok := options.Default.(bool); ok {
		prompt.Default = v
	}
	return askOneBool(prompt, options)
}

// Internal helpers
func askOneString(p survey.Prompt, opts AskOpts) (string, error) {
	TermLock.Lock()
	defer TermLock.Unlock()
	var out string
	err := survey.AskOne(p, &out, askOpts(opts)...)
	return out, err
}

func askOneBool(p survey.Prompt, opts AskOpts) (bool, error) {
	TermLock.Lock()
	defer TermLock.Unlock()
	var out bool
	err := survey.AskOne(p, &out, askOpts(opts)...)
	return out, err
}

func askManyString(p survey.Prompt, opts AskOpts) ([]string, error) {
	TermLock.Lock()
	defer TermLock.Unlock()
	var out []string
	err := survey.AskOne(p, &out, askOpts(opts)...)
	return out, err
}

func askOpts(opts AskOpts) []survey.AskOpt {
	as := []survey.AskOpt{}
	if opts.Validator != nil {
		as = append(as, survey.WithValidator(opts.Validator))
	}
	return as
}
