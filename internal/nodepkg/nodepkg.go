package nodepkg

import (
	"encoding/json"
	"path/filepath"

	"github.com/spf13/afero"
)

func InitPackage(fsys afero.Fs, dir string, params InitPackageParams) error {
	path := filepath.Join(dir, "package.json")
	b, err := afero.ReadFile(fsys, path)
	if err != nil {
		return err
	}

	var pkg map[string]any
	if err := json.Unmarshal(b, &pkg); err != nil {
		return err
	}

	// scripts merge
	scripts, _ := pkg["scripts"].(map[string]any)
	if scripts == nil {
		scripts = map[string]any{}
	}
	for k, v := range params.Scripts {
		if _, exists := scripts[k]; !exists {
			scripts[k] = v
		}
	}
	pkg["scripts"] = scripts

	// only set when provided
	if params.Name != "" {
		pkg["name"] = params.Name
	}
	if params.Main != "" {
		pkg["main"] = params.Main
	}

	out, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}
	return afero.WriteFile(fsys, path, out, 0o644)
}
