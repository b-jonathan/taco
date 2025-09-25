package nodepkg

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func InitPackage(dir string, params InitPackageParams) error {
	path := filepath.Join(dir, "package.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pkg map[string]any
	if err := json.Unmarshal(b, &pkg); err != nil {
		return err
	}
	scripts, _ := pkg["scripts"].(map[string]any)
	if scripts == nil {
		scripts = map[string]any{}
	}
	for k, v := range params.Scripts {
		scripts[k] = v
	}
	pkg["scripts"] = scripts

	pkg["name"] = params.Name
	pkg["main"] = params.Main

	out, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}
