package cli

type InitParams struct {
	Name         string
	Description  string
	Remote       string
	Private      bool
	Database_URI string
	UseGitHub 	 bool
}

type Step struct {
	Name string
	Fn   func() error
}
