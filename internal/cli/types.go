package cli

import "github.com/AlecAivazis/survey/v2"

type AskOpts struct {
	Default   any // string for Input and Select, []string for MultiSelect
	Help      string
	PageSize  int              // MultiSelect and Select only
	Validator survey.Validator // survey.Required or ComposeValidators(...)
}

type InitParams struct {
	Name        string
	Description string
	Remote      string
	Private     bool
}
