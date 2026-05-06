package user_applications

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/babbage88/infra-core/appmanifest"
)

var supportedManifestFilenames = map[string]struct{}{
	"package.json": {},
	"project.json": {},
	"go.mod":       {},
	"cargo.toml":   {},
}

var skippedDirectories = map[string]struct{}{
	".git":         {},
	".next":        {},
	"bin":          {},
	"build":        {},
	"dist":         {},
	"node_modules": {},
	"obj":          {},
	"target":       {},
	"vendor":       {},
}

func (svc *UserApplicationsService) DiscoverUserApplication(req DiscoverUserApplicationRequest) (*DiscoverUserApplicationResponse, error) {
	repoDir, err := cloneRepositoryForDiscovery(context.Background(), req.RepositoryUrl, req.Branch, req.Tag)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(repoDir)

	candidates, err := discoverManifestCandidates(repoDir, req.RepositoryUrl)
	if err != nil {
		return nil, err
	}
	return &DiscoverUserApplicationResponse{Candidates: candidates}, nil
}

func cloneRepositoryForDiscovery(ctx context.Context, repositoryURL, branch, tag string) (string, error) {
	tempDir, err := os.MkdirTemp("", "infractl-repo-discovery-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	args := []string{"clone", "--depth", "1"}
	if tag != "" {
		args = append(args, "--branch", tag)
	} else if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repositoryURL, tempDir)

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("clone repository %q: %w (%s)", repositoryURL, err, strings.TrimSpace(string(output)))
	}

	return tempDir, nil
}

func discoverManifestCandidates(repoDir, repositoryURL string) ([]DiscoveredUserApplicationCandidate, error) {
	candidates := make([]DiscoveredUserApplicationCandidate, 0)

	err := filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if _, skip := skippedDirectories[strings.ToLower(d.Name())]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if !isSupportedManifestPath(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read manifest %s: %w", path, err)
		}
		relPath, err := filepath.Rel(repoDir, path)
		if err != nil {
			return fmt.Errorf("derive relative path for %s: %w", path, err)
		}

		manifest, err := appmanifest.ParseManifestFile(relPath, content)
		if err != nil {
			return nil
		}
		if manifest.RepositoryURL == "" {
			manifest.RepositoryURL = repositoryURL
		}
		candidates = append(candidates, toDiscoveredCandidate(manifest))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan manifests: %w", err)
	}

	slices.SortFunc(candidates, func(a, b DiscoveredUserApplicationCandidate) int {
		return compareCandidates(a, b)
	})

	return candidates, nil
}

func isSupportedManifestPath(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if _, ok := supportedManifestFilenames[base]; ok {
		return true
	}
	return strings.HasSuffix(base, ".csproj")
}

func toDiscoveredCandidate(manifest appmanifest.Manifest) DiscoveredUserApplicationCandidate {
	return DiscoveredUserApplicationCandidate{
		Name:              manifest.Name,
		Description:       manifest.Description,
		RepositoryUrl:     manifest.RepositoryURL,
		ManifestPath:      manifest.ManifestPath,
		SourceKind:        manifest.SourceKind,
		ModuleName:        manifest.ModuleName,
		PackageName:       manifest.PackageName,
		PackageManager:    manifest.PackageManager,
		DeployKind:        manifest.DeployKind,
		ApplicationKind:   inferApplicationKind(manifest),
		Registerable:      manifest.Registerable,
		DeployConfig:      manifest.DeployConfig,
		BuildConfig:       manifest.BuildConfig,
		InfraDependencies: toDiscoveredDependencies(manifest.InfraDependencies),
	}
}

func toDiscoveredDependencies(deps []appmanifest.InfraDependency) []DiscoveredInfraDependency {
	converted := make([]DiscoveredInfraDependency, 0, len(deps))
	for _, dep := range deps {
		converted = append(converted, DiscoveredInfraDependency{
			DependencyType: dep.DependencyType,
			DependencyName: dep.DependencyName,
			Config:         dep.Config,
		})
	}
	return converted
}

func inferApplicationKind(manifest appmanifest.Manifest) string {
	switch manifest.DeployKind {
	case "nginx_static_site":
		return "frontend_spa"
	case "systemd_service":
		return "go_service"
	default:
		switch manifest.SourceKind {
		case "package.json", "project.json":
			return "frontend_spa"
		default:
			return "go_service"
		}
	}
}

func compareCandidates(a, b DiscoveredUserApplicationCandidate) int {
	scoreA := candidateScore(a)
	scoreB := candidateScore(b)
	if scoreA != scoreB {
		return scoreB - scoreA
	}
	return strings.Compare(a.ManifestPath, b.ManifestPath)
}

func candidateScore(candidate DiscoveredUserApplicationCandidate) int {
	score := 0
	if candidate.ManifestPath == "package.json" || candidate.ManifestPath == "go.mod" || candidate.ManifestPath == "Cargo.toml" {
		score += 50
	}
	if !strings.Contains(candidate.ManifestPath, "/") && !strings.Contains(candidate.ManifestPath, `\`) {
		score += 25
	}
	if candidate.DeployKind != "" {
		score += 15
	}
	if candidate.ApplicationKind == "frontend_spa" {
		score += 10
	}
	return score
}
