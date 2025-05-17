package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/internal/bumper"
	"gopkg.in/yaml.v3"
)

func pushNewTag(newTag string) error {
	output, err := bumper.RunGitCommand("tag", "-a", newTag, "-m", newTag)
	fmt.Println(output)
	return err
}

func BumpVersionAndCreateTag(currentVersion string, bumpType string) error {
	switch {
	case bumpType == bumper.Major || bumpType == bumper.Minor || bumpType == bumper.Patch:
		newTag, err := bumper.BumpVersion(currentVersion, bumpType)
		if err != nil {
			return fmt.Errorf("error bumping version err: %w", err)
		}
		return pushNewTag(newTag)
	default:
		slog.Warn("Invalid bumpType specifified, using patch", slog.String("bumpType", bumpType))
		newTag, err := bumper.BumpVersion(currentVersion, bumper.Patch)
		if err != nil {
			return fmt.Errorf("error bumping version err: %w", err)
		}
		return pushNewTag(newTag)

	}
}

// writeToYAML writes a key-value pair to a YAML file.
func writeToYAML(filename, key, value string) {
	data := make(map[string]string)

	// Read existing file if it exists
	if file, err := os.ReadFile(filename); err == nil {
		yaml.Unmarshal(file, &data)
	}

	// Update or add the key-value pair
	data[key] = value

	// Write back to file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error writing to YAML file:", err)
		return
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(data); err != nil {
		fmt.Println("Error encoding YAML:", err)
	}
}
