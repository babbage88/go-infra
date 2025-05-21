package main

import (
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/database/bootstrap"
	"github.com/babbage88/go-infra/internal/bumper"
	"github.com/babbage88/go-infra/internal/pretty"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	isLocalDevelopment bool
	srvport            string
	envFile            string
	bumpVersion        bool
	bootstrapNewDb     bool
	minor              bool
	major              bool
	testEncryption     bool
	initDevUser        bool
	version            bool
)

func testPgSecretStore() {

	usrSecret, err := user_secrets.Encrypt("Secret to be encrypted")
	if err != nil {
		slog.Error("Error testing Encryption", slog.String("Error", err.Error()))
		os.Exit(1)
	}
	usrSecret.PrintSecretInfo()

	plaintext, err := usrSecret.Decrypt()
	if err != nil {
		slog.Error("Error testing Decryption", slog.String("Error", err.Error()))

	}
	slog.Info("Decrytion", slog.String("Decrypted Value", string(plaintext)))
	connPool := initPgConnPool()
	pgSecretStore := user_secrets.PgUserSecretStore{DbConn: connPool}
	devuserUUID := uuid.MustParse(os.Getenv("DEV_USER_UUID"))
	appUUID := uuid.MustParse("f69a0abc-d82c-4013-9b25-b8abf4e4a896")
	secretUUID := uuid.MustParse("00d4fd03-d306-4436-9512-f5a0996d3be8")
	err = pgSecretStore.StoreSecret("TestPgSecretStoreExample", devuserUUID, appUUID)
	if err != nil {
		slog.Error("Error storing secret", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Secret successfully storeed in Pg Database")

	retrievedSecret, err := pgSecretStore.RetrieveSecret(secretUUID)
	if err != nil {
		slog.Error("Error retrieving secret from database", "error", err.Error())
		os.Exit(1)
	}

	secretBytes, err := io.ReadAll(retrievedSecret.Reader)
	if err != nil {
		slog.Error("Error reading secret from reader", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Retrieved secret", slog.String("secret", string(secretBytes)))
	os.Exit(0)

}

func startInLocalDevelopmentMode(envFile string) {
	slog.Info("Local Development mode configure, loading envars from env-file", slog.String("env-file", envFile))
	err := godotenv.Load(envFile)
	if err != nil {
		slog.Error("error loading .env file", slog.String("error", err.Error()))
	}
}

func bumpVersionNumber(major, minor bool) {
	var bumpErr error
	switch {
	case minor:
		bumpErr = versionInfo.FetchTagsAndBumpVersion(bumper.Minor)
	case major:
		bumpErr = versionInfo.FetchTagsAndBumpVersion(bumper.Major)
	default:
		bumpErr = versionInfo.FetchTagsAndBumpVersion(bumper.Patch)
	}
	if bumpErr != nil {
		slog.Error("Error bumping version", slog.String("error", bumpErr.Error()))
		os.Exit(1)
	}
	slog.Info("Bumped verion number", "NewVersion", versionInfo.Version)
	os.Exit(0)
}

func bootstrapDb() {
	bootstrap.NewDb()
	pretty.Print("test")
	err := bootstrap.CreateInfradbUser(os.Getenv("DB_USER"))
	if err != nil {
		pretty.PrintErrorf("Error configuring db user %s", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func parseFlags() {
	flag.BoolVar(&isLocalDevelopment, "local-development", false, "Flag to configure running local developement mode, envars set froma .env file")
	flag.StringVar(&envFile, "env-file", ".env", "Path to .env file to load Environment Variables.")
	flag.StringVar(&srvport, "srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	flag.BoolVar(&bootstrapNewDb, "db-bootstrap", false, "Create new dev database.")
	flag.BoolVar(&initDevUser, "devuser", false, "Update the devuser password")
	flag.BoolVar(&version, "version", false, "Show the current version.")
	flag.BoolVar(&bumpVersion, "bump-version", false, "Bumps version tag, push to remote repo and update version.yaml")
	flag.BoolVar(&minor, "minor", false, "Bumps Minor version number")
	flag.BoolVar(&major, "major", false, "Bumps Major version number")
	flag.BoolVar(&testEncryption, "test-enc", false, "testing/debugging encrytion package")
	flag.Parse()

}

func configureStartupOptions() {
	if isLocalDevelopment {
		startInLocalDevelopmentMode(envFile)
	}

	if testEncryption {
		testPgSecretStore()
	}

	if bumpVersion {
		bumpVersionNumber(major, minor)
	}

	if bootstrapNewDb {
		bootstrapDb()
	}

	if version {
		versionInfo.PrintVersion()
		return
	}
}
