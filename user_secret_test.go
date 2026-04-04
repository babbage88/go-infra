package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMillionStoredSecretsHaveUniqueEncryption(t *testing.T) {
	const (
		samplePlaintext = "this is a sample user token"
		totalSecrets    = 1_000_000
		testDBName      = "gotestdb"
		testDBUser      = "gotestdb"
		testDBPassword  = "dothetest"
		testDBHost      = "10.2.10.248"
		testDBPort      = 5432
	)

	// Ensure the encryption key is set and valid
	key := "12345678901234567890123456789012" // Must be 32 bytes
	if err := os.Setenv("USER_SEC_KEY", key); err != nil {
		t.Fatalf("failed to set USER_SEC_KEY: %v", err)
	}

	// Connect to the test database (adjust DSN as needed)
	dbURL := os.Getenv("PG_TEST_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", testDBUser, testDBPassword, testDBHost, testDBPort, testDBName)
		if err := ensureTestDatabase(t, testDBName, testDBUser, testDBPassword, testDBHost); err != nil {
			t.Fatalf("failed to provision test database: %v", err)
		}
		if err := os.Setenv("PG_TEST_URL", dbURL); err != nil {
			t.Fatalf("failed to set PG_TEST_URL: %v", err)
		}
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	provider := user_secrets.NewPgUserSecretStore(pool)

	userId := uuid.MustParse(os.Getenv("DEV_USER_UUID"))
	appId := uuid.MustParse("f69a0abc-d82c-4013-9b25-b8abf4e4a896")

	uniqueCiphertexts := make(map[string]struct{}, totalSecrets)
	secretIDs := make([]uuid.UUID, 0, totalSecrets)

	t.Logf("Storing %d secrets...", totalSecrets)
	for i := 0; i < totalSecrets; i++ {
		secretId, err := provider.StoreSecret(samplePlaintext, userId, appId, time.Now())
		if err != nil {
			t.Fatalf("failed to store secret at iteration %d: %v", i, err)
		}

		secretIDs = append(secretIDs, secretId)
	}

	t.Log("Retrieving and checking uniqueness of encrypted secrets...")

	for i, id := range secretIDs {
		secret, err := provider.RetrieveSecret(id)
		if err != nil {
			t.Fatalf("failed to retrieve secret ID %s: %v", id.String(), err)
		}

		// base64 the raw encrypted bytes to compare
		encoded := base64.StdEncoding.EncodeToString(secret.ExternalAuthToken.Token)

		if _, exists := uniqueCiphertexts[encoded]; exists {
			t.Fatalf("duplicate ciphertext detected at index %d (id: %s)", i, id.String())
		}

		uniqueCiphertexts[encoded] = struct{}{}

		if i > 0 && i%100_000 == 0 {
			t.Logf("%d secrets retrieved and validated", i)
		}
	}

	t.Logf("Successfully validated %d unique encrypted secrets.", totalSecrets)
}

func ensureTestDatabase(t *testing.T, dbName, dbUser, dbPassword, dbHost string) error {
	t.Helper()

	args := []string{
		"database", "new-appdb",
		"--ssh-remote-host", dbHost,
		"--ssh-key", expandPath("~/.ssh/id_ed25519"),
		"--ssh-remote-user", "root",
		"--create-db",
		"--connect-ssh",
		"--drop-first",
		"--db-name", dbName,
		"--db-user", dbUser,
		"--db-password", dbPassword,
		"--goosey-path", filepath.Join("..", "infra-db", "goosey"),
	}

	cmd := exec.Command("infractl", args...)
	cmd.Dir = filepath.Join("..", "infra-cli")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Logf("provisioned test database with infractl")
		return nil
	}

	if _, lookErr := exec.LookPath("infractl"); lookErr == nil {
		return fmt.Errorf("run infractl: %w: %s", err, string(output))
	}

	fallbackArgs := append([]string{"run", filepath.Join("..", "infra-cli")}, args...)
	fallback := exec.Command("go", fallbackArgs...)
	fallback.Dir = filepath.Join("..", "infra-cli")
	fallback.Env = os.Environ()
	output, err = fallback.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run fallback go command: %w: %s", err, string(output))
	}

	t.Logf("provisioned test database with go run fallback")
	return nil
}

func expandPath(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
