package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/internal/bumper"
	"github.com/goccy/go-yaml"
)

func (v *VersionInfo) pushNewTag(newTag string) error {
	output, err := bumper.RunGitCommand("tag", "-a", newTag, "-m", newTag)
	if err != nil {
		fmt.Println(output)
		return fmt.Errorf("error pushing new tags %w", err)
	}
	fmt.Println(output)
	v.writeToYAML(bumper.VersionYamlFile)
	v.Version = newTag
	return err
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
