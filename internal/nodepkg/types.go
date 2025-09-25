package nodepkg

type InitPackageParams struct {
	Name    string
	Main    string
	Scripts map[string]string // merged into existing
}
