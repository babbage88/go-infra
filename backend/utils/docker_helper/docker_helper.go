package docker_helper

import (
	"log/slog"
	"os"
)

func GetSecret(secret string) string {
	secret_path := "/run/secrets/" + secret
	byte_secret, err := os.ReadFile(secret_path)
	if err != nil {
		slog.Error("Error parsing secre", slog.String("Error", err.Error()))
		return ""
	}

	ret_val := string(byte_secret)

	return ret_val
}

/*
From main

	db_pw := docker_helper.GetSecret("DB_PW")
	le_ini := docker_helper.GetSecret("trahan.dev_token")

	if le_ini == "" {
		slog.Warn("Le auth blank")
	}
*/
