package main

import (
	"context"
	"encoding/base64"
	"os"
	"testing"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMillionStoredSecretsHaveUniqueEncryption(t *testing.T) {
	const (
		samplePlaintext = "this is a sample user token"
		totalSecrets    = 1_000_000
	)

	// Ensure the encryption key is set and valid
	key := "12345678901234567890123456789012" // Must be 32 bytes
	if err := os.Setenv("USER_SEC_KEY", key); err != nil {
		t.Fatalf("failed to set USER_SEC_KEY: %v", err)
	}

	// Connect to the test database (adjust DSN as needed)
	dbURL := os.Getenv("PG_TEST_URL")
	if dbURL == "" {
		t.Fatal("PG_TEST_URL environment variable not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	qry := infra_db_pg.New(pool)
	provider := &user_secrets.PgUserSecretStore{DbConn: pool}

	userId := uuid.MustParse(os.Getenv("DEV_USER_UUID"))
	appId := uuid.MustParse("f69a0abc-d82c-4013-9b25-b8abf4e4a896")

	uniqueCiphertexts := make(map[string]struct{}, totalSecrets)
	secretIDs := make([]uuid.UUID, 0, totalSecrets)

	t.Logf("Storing %d secrets...", totalSecrets)
	for i := 0; i < totalSecrets; i++ {
		err := provider.StoreSecret(samplePlaintext, userId, appId)
		if err != nil {
			t.Fatalf("failed to store secret at iteration %d: %v", i, err)
		}

		// Retrieve latest ID by user & app (this assumes latest ID wins, otherwise tweak schema/indexing)
		row, err := qry.GetLatestExternalAuthToken(context.Background(), infra_db_pg.GetLatestExternalAuthTokenParams{
			UserID:        userId,
			ExternalAppID: appId,
		})
		if err != nil {
			t.Fatalf("failed to fetch stored secret id at iteration %d: %v", i, err)
		}

		secretIDs = append(secretIDs, row.ID)
	}

	t.Log("Retrieving and checking uniqueness of encrypted secrets...")

	for i, id := range secretIDs {
		secret, err := provider.RetrieveSecret(id)
		if err != nil {
			t.Fatalf("failed to retrieve secret ID %s: %v", id.String(), err)
		}

		// base64 the raw encrypted bytes to compare
		encoded := base64.StdEncoding.EncodeToString(secret.Metadata.Token)

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
