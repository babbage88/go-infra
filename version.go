package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/internal/bumper"
	"github.com/goccy/go-yaml"
)

type VersionInfo struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Author  string `json:"author" yaml:"author"`
}

func marshalVersionInfo() VersionInfo {
	f, err := os.ReadFile(bumper.VersionYamlFile)
	if err != nil {
		slog.Error("error reading the version.yaml file", slog.String("filename", bumper.VersionYamlFile), slog.String("error", err.Error()))
	}

	var versionInfo VersionInfo
	err = yaml.Unmarshal(f, &versionInfo)
	if err != nil {
		slog.Error("Error unmarshalling the version.yaml file")
	}
	return versionInfo
}

func (v *VersionInfo) PrintVersion() string {
	versionString := fmt.Sprintf("%s %s\n", v.Name, v.Version)
	fmt.Println(versionString)
	return versionString

}

func (v *VersionInfo) LogVersionInfo() {
	slog.Info("Verion Info", slog.String("Name", v.Name), slog.String("Version", v.Version), slog.String("Author", v.Author))
}

func (v *VersionInfo) CheckMatchesGitAndBump(bumpType string) error {
	err := bumper.FetchTags()
	if err != nil {
		slog.Error("error fetching remote tag", "error", err.Error())
		return fmt.Errorf("error fetching remote tags from origin %w", err)
	}
	latestTag, err := bumper.GetLatestSemverTag()
	if err != nil {
		slog.Error("error getting latest git tags from remote", slog.String("error", err.Error()))
		return fmt.Errorf("error getting latest tags")
	}

	result := bumper.CompareSemver(latestTag, v.Version)
	switch {
	case result == -1:
		slog.Warn("The latest tag git tag is lower than the version defined in version.yaml", slog.String("version.yaml", v.Version), slog.String("latest-tag", latestTag))
		output, err := bumper.RunGitCommand("tag", "-a", v.Version, "-m", v.Version)
		if err != nil {
			return fmt.Errorf("error creating new tag %w", err)
		}
		fmt.Println(output)
		pushOutput, err := bumper.RunGitCommand("push", "--tags")
		if err != nil {
			return fmt.Errorf("error pusing new tag %w", err)
		}
		fmt.Println(pushOutput)
		return err
	case result == 0:
		// implement git describe --exact-match --tags HEAD
		//git describe --exact-match --tags HEAD
		bumpedVer, err := bumper.BumpVersion(latestTag, bumper.Patch)
		bumper.RunGitCommand()
		fmt.Println(bumpedVer)
		return err
	case result == 1:
		//update version.yaml if HEAD matches tag
		return err
	default:
		bumpedVer, err := bumper.BumpVersion(latestTag, bumper.Patch)
		bumper.RunGitCommand()
		fmt.Println(bumpedVer)
		return err
		//
	}

}
