package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/internal/bumper"
	"github.com/goccy/go-yaml"
)

const (
	pushCmd               string = "push"
	pushTagsCmdFlag       string = "--tags"
	gitForceFlag          string = "--force"
	gitTagCmd             string = "tag"
	gitTagCmdAddFlag      string = "-a"
	gitTagCmdMessageFlag  string = "-m"
	gitFetchCmd           string = "fetch"
	gitPullCmd            string = "pull"
	gitFetchPruneFlag     string = "--prune"
	gitFetchPruneTagsFlag string = "--prune-tags"
	gitFetchTagsFlag      string = "--tags"
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

func (v *VersionInfo) PullAndCompare(bumpType string) error {
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
		output, err := bumper.RunGitCommand(gitTagCmd, gitTagCmdAddFlag, v.Version, gitTagCmdMessageFlag, v.Version)
		if err != nil {
			return fmt.Errorf("error creating new tag %w", err)
		}
		fmt.Println(output)
		pushOutput, err := bumper.RunGitCommand(pushCmd, pushTagsCmdFlag)
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

func (v *VersionInfo) pushNewTag(newTag string) error {
	output, err := bumper.RunGitCommand(gitTagCmd, gitTagCmdAddFlag, newTag, gitTagCmdMessageFlag, newTag)
	if err != nil {
		fmt.Println(output)
		return fmt.Errorf("error pushing new tags %w", err)
	}
	fmt.Println(output)
	pushTagsOutput, err := bumper.RunGitCommand(pushCmd, pushTagsCmdFlag, gitForceFlag)
	if err != nil {
		return fmt.Errorf("error pushing new tags %w", err)
	}
	fmt.Println(pushTagsOutput)
	v.Version = newTag
	v.writeToYAML(bumper.VersionYamlFile)
	return err
}

func (v *VersionInfo) FetchTagsAndBumpVersion(bumpType string) error {
	err := bumper.FetchTags()
	if err != nil {
		slog.Error("error fetching remote tag", "error", err.Error())
		return fmt.Errorf("error fetching remote tags from origin %w", err)
	}
	pruneFlagsOutput, err := bumper.RunGitCommand(gitFetchCmd, gitFetchPruneFlag, gitFetchPruneTagsFlag, gitFetchTagsFlag)
	if err != nil {
		fmt.Println(pruneFlagsOutput)
		return fmt.Errorf("error pruning tags %w", err)
	}
	fmt.Println(pruneFlagsOutput)
	latestTag, err := bumper.GetLatestSemverTag()
	if err != nil {
		slog.Error("error getting latest git tags from remote", slog.String("error", err.Error()))
		return fmt.Errorf("error getting latest tags")
	}

	return v.BumpVersionAndCreateTag(latestTag, bumpType)
}

func (v *VersionInfo) BumpVersionAndCreateTag(currentVersion string, bumpType string) error {
	switch {
	case bumpType == bumper.Major || bumpType == bumper.Minor || bumpType == bumper.Patch:
		newTag, err := bumper.BumpVersion(currentVersion, bumpType)
		if err != nil {
			return fmt.Errorf("error bumping version err: %w", err)
		}
		return v.pushNewTag(newTag)
	default:
		slog.Warn("Invalid bumpType specifified, using patch", slog.String("bumpType", bumpType))
		newTag, err := bumper.BumpVersion(currentVersion, bumper.Patch)
		if err != nil {
			return fmt.Errorf("error bumping version err: %w", err)
		}
		return v.pushNewTag(newTag)
	}
}

// writeToYAML writes a key-value pair to a YAML file.
func (v *VersionInfo) writeToYAML(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		slog.Error("Error creating YAML file", "error", err.Error())
		return err
	}
	defer file.Close()

	verInfoBytes, err := yaml.Marshal(v)
	if err != nil {
		slog.Error("Error marshalling yaml")
		return fmt.Errorf("error marshaling yaml %w", err)
	}

	bytesWritten, err := file.Write(verInfoBytes)
	if err != nil {
		slog.Error("Error writing bytes to new version fileyaml")
		return fmt.Errorf("error marshawriting file %s %w", filename, err)
	}
	slog.Info("New version file created", slog.String("filename", filename), slog.Int("BytesWritten", bytesWritten))
	return nil
}
