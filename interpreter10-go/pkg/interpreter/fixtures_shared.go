package interpreter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter10-go/pkg/ast"
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
		Errors               []string `json:"errors"`
		TypecheckDiagnostics []string `json:"typecheckDiagnostics"`
	} `json:"expect"`
}

func readManifest(t testingT, dir string) fixtureManifest {
	t.Helper()
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fixtureManifest{}
		}
		t.Fatalf("read manifest %s: %v", manifestPath, err)
	}
	var manifest fixtureManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse manifest %s: %v", manifestPath, err)
	}
	return manifest
}

func readModule(t testingT, path string) (*ast.Module, string) {
	t.Helper()
	if strings.HasSuffix(path, ".able") {
		mod, err := parseSourceModule(path)
		if err != nil {
			t.Fatalf("parse source module %s: %v", path, err)
		}
		return mod, path
	}
	if strings.HasSuffix(path, ".json") {
		dir := filepath.Dir(path)
		base := strings.TrimSuffix(filepath.Base(path), ".json")
		candidates := []string{filepath.Join(dir, base+".able")}
		if base == "module" {
			candidates = append(candidates, filepath.Join(dir, "source.able"))
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				if mod, err := parseSourceModule(candidate); err == nil {
					return mod, candidate
				}
			}
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read module %s: %v", path, err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse module %s: %v", path, err)
	}
	node, err := decodeNode(raw)
	if err != nil {
		t.Fatalf("decode module %s: %v", path, err)
	}
	mod, ok := node.(*ast.Module)
	if !ok {
		t.Fatalf("decoded node is not module: %T", node)
	}
	return mod, path
}
