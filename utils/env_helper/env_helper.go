package env_helper

import (
	"os"

	"log/slog"

	"github.com/babbage88/go-infra/utils/type_helper"
	"github.com/joho/godotenv"
)

type EnvVars struct {
	DotFileName string `json:"dotFileName"`
	VarName     string `json:"envVarName"`
	SetValue    bool   `json:"setValue"`
}

type VarValueProvider interface {
	GetEnvVarValue() string
}

type int64ValueProvider interface {
	ParseEnvVarInt64() (int64, error)
}

type EnvVarsOptions func(*EnvVars)

func (dotvar EnvVars) GetEnvVarValue() string {
	return getDotEnvVariable(dotvar.VarName, dotvar.DotFileName)
}

func (dotvar EnvVars) ParseEnvVarInt64() (int64, error) {
	s := getDotEnvVariable(dotvar.VarName, dotvar.DotFileName)
	return type_helper.ParseInt64(s)
}

func NewDotEnvSource(opts ...EnvVarsOptions) *EnvVars {
	const (
		dotFileName = ".env"
		varName     = "JWT_KEY"
		setValue    = false
	)

	envars := &EnvVars{
		DotFileName: dotFileName,
		VarName:     varName,
		SetValue:    setValue,
	}

	for _, opt := range opts {
		opt(envars)
	}

	return envars
}

func WithDotEnvFileName(dotFileName string) EnvVarsOptions {
	return func(c *EnvVars) {
		c.DotFileName = dotFileName
	}
}

func WithVarName(varName string) EnvVarsOptions {
	return func(c *EnvVars) {
		c.VarName = varName
	}
}

func WithSetValue(setValue bool) EnvVarsOptions {
	return func(c *EnvVars) {
		c.SetValue = setValue
	}
}

func getDotEnvVariable(key string, dotfilename string) string {

	// load .env file
	err := godotenv.Load(dotfilename)

	if err != nil {
		slog.Error("Error loading .env file", slog.String("Error", err.Error()))
	}

	return os.Getenv(key)
}
