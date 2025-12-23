package interpreter

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type fixtureManifest struct {
	Description string   `json:"description"`
	Entry       string   `json:"entry"`
	Setup       []string `json:"setup"`
	SkipTargets []string `json:"skipTargets"`
	Expect      struct {
		Result *struct {
			Kind  string      `json:"kind"`
			Value interface{} `json:"value"`
		} `json:"result"`
		Stdout               []string `json:"stdout"`
		Stderr               []string `json:"stderr"`
		Exit                 *int     `json:"exit"`
		Errors               []string `json:"errors"`
		TypecheckDiagnostics []string `json:"typecheckDiagnostics"`
	} `json:"expect"`
}

// FixtureManifest exposes the manifest schema for external consumers.
type FixtureManifest = fixtureManifest

func readManifest(t testingT, dir string) fixtureManifest {
	t.Helper()
	manifest, err := LoadFixtureManifest(dir)
	if err != nil {
		t.Fatalf("read manifest %s: %v", filepath.Join(dir, "manifest.json"), err)
	}
	return manifest
}

func readModule(t testingT, path string) (*ast.Module, string) {
	t.Helper()
	module, origin, err := LoadFixtureModule(path)
	if err != nil {
		t.Fatalf("read module %s: %v", path, err)
	}
	return module, origin
}

// LoadFixtureManifest reads a fixture manifest from disk.
func LoadFixtureManifest(dir string) (fixtureManifest, error) {
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fixtureManifest{}, nil
		}
		return fixtureManifest{}, err
	}
	var manifest fixtureManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fixtureManifest{}, err
	}
	return manifest, nil
}

// LoadFixtureModule loads a fixture module (JSON or Able source) and returns the module plus origin.
func LoadFixtureModule(path string) (*ast.Module, string, error) {
	if strings.HasSuffix(path, ".able") {
		mod, err := parseSourceModule(path)
		if err != nil {
			return nil, "", err
		}
		return mod, path, nil
	}

	if strings.HasSuffix(path, ".json") {
		if mod, err := parseModuleJSON(path); err == nil {
			origin := path
			if sibling := sourceSibling(path); sibling != "" {
				if spanModule, err := parseSourceModule(sibling); err == nil && spanModule != nil {
					ast.CopySpans(mod, spanModule)
				}
				origin = sibling
			}
			return mod, origin, nil
		}

		dir := filepath.Dir(path)
		base := strings.TrimSuffix(filepath.Base(path), ".json")
		candidates := []string{filepath.Join(dir, base+".able")}
		if base == "module" {
			candidates = append(candidates, filepath.Join(dir, "source.able"))
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				if mod, err := parseSourceModule(candidate); err == nil {
					return mod, candidate, nil
				}
			}
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, "", err
	}
	node, err := decodeNode(raw)
	if err != nil {
		return nil, "", err
	}
	mod, ok := node.(*ast.Module)
	if !ok {
		return nil, "", fs.ErrInvalid
	}
	return mod, path, nil
}

func parseModuleJSON(path string) (*ast.Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	node, err := decodeNode(raw)
	if err != nil {
		return nil, err
	}
	mod, ok := node.(*ast.Module)
	if !ok {
		return nil, fs.ErrInvalid
	}
	return mod, nil
}

func sourceSibling(jsonPath string) string {
	dir := filepath.Dir(jsonPath)
	base := strings.TrimSuffix(filepath.Base(jsonPath), ".json")
	candidates := []string{filepath.Join(dir, base+".able")}
	if base == "module" {
		candidates = append(candidates, filepath.Join(dir, "source.able"))
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}
