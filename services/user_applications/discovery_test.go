package user_applications

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverManifestCandidatesPrefersRootFrontendPackageJSON(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	rootPackageJSON := `{
		"name": "infractl-ui",
		"dependencies": { "react": "^19.0.0" },
		"devDependencies": { "vite": "^6.0.0" },
		"infractl": {
			"deployKind": "nginx_static_site"
		}
	}`
	nestedGoMod := `module github.com/example/worker`

	if err := os.WriteFile(filepath.Join(repoDir, "package.json"), []byte(rootPackageJSON), 0o644); err != nil {
		t.Fatalf("write root package.json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, "tools", "worker"), 0o755); err != nil {
		t.Fatalf("mkdir worker dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "tools", "worker", "go.mod"), []byte(nestedGoMod), 0o644); err != nil {
		t.Fatalf("write nested go.mod: %v", err)
	}

	candidates, err := discoverManifestCandidates(repoDir, "https://github.com/babbage88/infractl-ui")
	if err != nil {
		t.Fatalf("discoverManifestCandidates returned error: %v", err)
	}
	if len(candidates) < 2 {
		t.Fatalf("expected at least 2 candidates, got %d", len(candidates))
	}
	if candidates[0].ApplicationKind != "frontend_spa" {
		t.Fatalf("expected top candidate to be frontend_spa, got %q", candidates[0].ApplicationKind)
	}
	if candidates[0].ManifestPath != "package.json" {
		t.Fatalf("expected top candidate manifest path package.json, got %q", candidates[0].ManifestPath)
	}
}
