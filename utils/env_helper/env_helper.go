package env_helper

import (
	"os"

	"log/slog"

	"github.com/babbage88/go-infra/internal/type_helper"
	"github.com/joho/godotenv"
)

type EnvVars struct {
	DotFileName string            `json:"dotFileName"`
	VarName     string            `json:"envVarName"`
	SetValue    bool              `json:"setValue"`
	VarMap      map[string]string `json:"varMap"`
}

type VarMapperProvider interface {
	ParseEnvVariables() error
}

func (m *EnvVars) ParseEnvVariables() {
	envreader, err := os.Open(m.DotFileName)
	if err != nil {
		slog.Error("Error readinf .env file", slog.String("Error", err.Error()))
	}

	defer func() {
		envreader.Close()
	}()

	m.VarMap, err = godotenv.Parse(envreader)
	if err != nil {
		slog.Error("Error parsing .env variables into struct.", slog.String("Error", err.Error()))
	}
}

type EnvVarMapValueProvier interface {
	GetVarMapValue(s string) string
}

func (m *EnvVars) GetVarMapValue(key string) string {
	value, exists := m.VarMap[key]
	if exists {
		return value
	} else {
		return ""
	}
}

type VarValueProvider interface {
	GetEnvVarValue() string
}

type int64ValueProvider interface {
	ParseEnvVarInt64(key string) (int64, error)
}

type int32ValueProvider interface {
	ParseEnvVarInt32(key string) (int64, error)
}

type EnvVarsOptions func(*EnvVars)

func (dotvar EnvVars) GetEnvVarValue() string {
	return getDotEnvVariable(dotvar.VarName, dotvar.DotFileName)
}

func (dotvar *EnvVars) ParseEnvVarInt64(key string) (int64, error) {
	s := dotvar.GetVarMapValue(key)
	return type_helper.ParseInt64(s)
}

func (dotvar *EnvVars) ParseEnvVarInt32(key string) (int32, error) {
	s := dotvar.GetVarMapValue(key)
	return type_helper.ParseInt32(s)
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

func LoadEnvFile(env_file string) error {
	// load .env file
	err := godotenv.Load(env_file)

	if err != nil {
		slog.Error("Error loading .env file", slog.String("Error", err.Error()))
	}

	return err
}
