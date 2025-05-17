package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
)

const (
	versionYamlFile string = "version.yaml"
)

type VersionInfo struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Author  string `json:"author" yaml:"author"`
}

func marshalVersionInfo() VersionInfo {
	f, err := os.ReadFile(versionYamlFile)
	if err != nil {
		slog.Error("error reading the version.yaml file", slog.String("filename", versionYamlFile), slog.String("error", err.Error()))
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
