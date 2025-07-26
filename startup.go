package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/babbage88/go-infra/database/bootstrap"
	"github.com/babbage88/go-infra/internal/bumper"
	"github.com/babbage88/go-infra/internal/embedbin"
	"github.com/babbage88/go-infra/internal/pretty"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	userHttps          bool
	certFile           string
	certKey            string
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
	initDb             bool
	migrationBin       string
)

func testAES256GCMEncryptDecrypt() {
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
}

func testPgSecretStore() {

	connPool := initPgConnPool()
	pgSecretStore := user_secrets.NewPgUserSecretStore(connPool)
	devuserUUID := uuid.MustParse(os.Getenv("DEV_USER_UUID"))
	appUUID := uuid.MustParse("f69a0abc-d82c-4013-9b25-b8abf4e4a896")
	secretUUID := uuid.MustParse("f7a62a3b-9680-441f-9fa2-6339bb419a47")
	secretId, err := pgSecretStore.StoreSecret("TestPgSecretStoreExample", devuserUUID, appUUID, time.Now().AddDate(0, 0, 1))
	if err != nil {
		slog.Error("Error storing secret", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Secret successfully stored in Pg Database", slog.String("secretId", secretId.String()))

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
	flag.BoolVar(&userHttps, "use-https", false, "Starts api with TLS/SSL with the specified cert and key")
	flag.StringVar(&certFile, "cert-file", "server.crt", "Path to certificate file to use")
	flag.StringVar(&certKey, "cert-key", "server.key", "Path to certificate key file file to use")
	flag.BoolVar(&minor, "minor", false, "Bumps Minor version number")
	flag.BoolVar(&major, "major", false, "Bumps Major version number")
	flag.BoolVar(&testEncryption, "test-enc", false, "testing/debugging encrytion package")
	flag.BoolVar(&initDb, "init-db", false, "Initialize the database with default schema and data")
	flag.StringVar(&migrationBin, "migration-bin", "", "Path to the migration binary to run (optional, uses embedded binary if not specified)")
	flag.Parse()

}

func configureStartupOptions() {
	if isLocalDevelopment || envFile != ".env" {
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

	if initDb {
		if err := runMigration(migrationBin); err != nil {
			slog.Error("Error running migration", slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("Database initialized and migrated successfully.")
		os.Exit(0)
	}

	if version {
		versionInfo.PrintVersion()
		return
	}
}

// runMigration executes a goose migration binary with the GOOSE_DBSTRING environment variable set.
//
// Usage:
//
//	go-infra --init-db --migration-bin /path/to/goose --env-file .env
//	go-infra --init-db --env-file .env  # Uses embedded binary
//
// The function will:
// 1. Check if the migration binary exists at the specified path (or extract embedded binary)
// 2. Verify that DATABASE_URL environment variable is set
// 3. Set GOOSE_DBSTRING environment variable to the value of DATABASE_URL
// 4. Execute the migration binary with stdout/stderr forwarded to the console
// 5. Exit with code 0 on success, 1 on failure
func runMigration(binPath string) error {
	var actualBinPath string
	var cleanupFunc func()

	if binPath == "" {
		// Use embedded binary
		extractedPath, cleanup, err := extractEmbeddedBinary()
		if err != nil {
			return fmt.Errorf("failed to extract embedded binary: %w", err)
		}
		actualBinPath = extractedPath
		cleanupFunc = cleanup
		defer cleanupFunc()
	} else {
		// Check if the specified binary file exists
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			return fmt.Errorf("migration binary does not exist at path: %s", binPath)
		}
		actualBinPath = binPath
	}

	// Get the DATABASE_URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	slog.Info("Running migration binary",
		slog.String("binary", actualBinPath),
		slog.String("database_url", databaseURL))

	cmd := exec.Command(actualBinPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set the GOOSE_DBSTRING environment variable
	cmd.Env = append(os.Environ(), "GOOSE_DBSTRING="+databaseURL)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run migration binary %s: %w", actualBinPath, err)
	}

	slog.Info("Migration binary completed successfully")
	return nil
}

// extractEmbeddedBinary extracts the embedded goinfra-goosey binary to a temporary file
// and returns the path to the extracted file and a cleanup function.
func extractEmbeddedBinary() (string, func(), error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "goinfra-goosey-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Write the embedded binary to the temp file
	if _, err := tmpFile.Write(embedbin.GooseyBinary); err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, fmt.Errorf("failed to write embedded binary to temp file: %w", err)
	}

	// Make the file executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, fmt.Errorf("failed to make temp file executable: %w", err)
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup, nil
}
