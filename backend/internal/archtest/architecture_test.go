package archtest

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDomainHasNoForbiddenImports ensures that all domain layers strictly adhere to Clean Architecture:
// 1. Domain must not import outer layers (delivery, infrastructure).
// 2. Domain must not import database drivers, ORMs, HTTP frameworks, or CLI libraries.
func TestDomainHasNoForbiddenImports(t *testing.T) {
	backendRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to resolve backend root: %v", err)
	}

	internalDir := filepath.Join(backendRoot, "internal")

	forbiddenImportPrefixes := []string{
		"database/sql",
		"net/http",
		"github.com/jackc/pgx",
		"github.com/go-chi",
		"github.com/gin-gonic",
		"github.com/spf13/cobra",
		"github.com/docker/docker",
	}

	fset := token.NewFileSet()

	err = filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Check if the file is inside a domain directory
		parts := strings.Split(filepath.ToSlash(path), "/")
		isDomain := false
		for _, part := range parts {
			if part == "domain" {
				isDomain = true
				break
			}
		}

		if !isDomain {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Errorf("Failed to parse file %s: %v", path, parseErr)
			return nil
		}

		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Rule 1: No infrastructure or delivery packages in domain
			if strings.Contains(importPath, "/infrastructure") || strings.Contains(importPath, "/delivery") {
				t.Errorf("Clean Architecture Violation: domain file %s imports outer layer %s", path, importPath)
			}

			// Rule 2: No forbidden database/HTTP/CLI packages in domain
			for _, forbidden := range forbiddenImportPrefixes {
				if strings.HasPrefix(importPath, forbidden) {
					t.Errorf("Clean Architecture Violation: domain file %s imports forbidden package %s", path, importPath)
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Walk error in internalDir: %v", err)
	}
}

// TestInstallerHasNoBackendImports verifies that installer does not import backend packages.
func TestInstallerHasNoBackendImports(t *testing.T) {
	backendRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to resolve backend root: %v", err)
	}

	repoRoot := filepath.Dir(backendRoot)
	installerDir := filepath.Join(repoRoot, "installer")

	fset := token.NewFileSet()

	err = filepath.Walk(installerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Errorf("Failed to parse file %s: %v", path, parseErr)
			return nil
		}

		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if strings.HasPrefix(importPath, "github.com/vulnex/vulnex/backend") {
				t.Errorf("Module Boundary Violation: installer file %s directly imports backend package %s", path, importPath)
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Walk error in installerDir: %v", err)
	}
}
