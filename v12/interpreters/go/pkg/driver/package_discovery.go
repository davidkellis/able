package driver

import (
	"fmt"
	"path/filepath"
	"sort"
)

// PackageSource describes a package discovered beneath a source root.
// Files are absolute, sorted Able source paths.
type PackageSource struct {
	Name  string
	Files []string
}

// DiscoverPackages returns all packages visible in the supplied roots without
// parsing or evaluating their source. It uses the same package-name resolution
// as Loader, so callers can safely pass the returned names to LoadOptions.
func DiscoverPackages(searchPaths []SearchPath, includeTests bool) ([]PackageSource, error) {
	byName := make(map[string]PackageSource)
	for _, searchPath := range searchPaths {
		if searchPath.Path == "" {
			continue
		}
		rootDir, rootName, err := discoverRootForPath(searchPath.Path)
		if err != nil {
			return nil, err
		}
		kind, _ := determineSearchRootMetadata(searchPath, rootDir, rootName)
		packages, _, err := indexSourceFiles(rootDir, rootName, kind, includeTests)
		if err != nil {
			return nil, err
		}
		for name, files := range packages {
			if existing, ok := byName[name]; ok {
				return nil, fmt.Errorf(
					"loader: package %s found in multiple roots (%s, %s)",
					name,
					filepath.Dir(existing.Files[0]),
					rootDir,
				)
			}
			byName[name] = PackageSource{
				Name:  name,
				Files: append([]string(nil), files...),
			}
		}
	}

	packages := make([]PackageSource, 0, len(byName))
	for _, source := range byName {
		packages = append(packages, source)
	}
	sort.Slice(packages, func(i, j int) bool { return packages[i].Name < packages[j].Name })
	return packages, nil
}
