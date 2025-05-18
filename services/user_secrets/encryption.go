package user_secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
)

type EncryptedUserSecretsAES256GCM struct {
	Ciphertext []byte `json:"ciphertext"`
}

func (s *EncryptedUserSecretsAES256GCM) PrintSecretInfo() {
	nonce, ciphertext := s.Ciphertext[:12], s.Ciphertext[12:]
	slog.Info("EncryptedUserSecrets Info", slog.String("ciphertext", string(ciphertext)), slog.String("nonce", string(nonce)))
	fmt.Println("Raw Bytes:", s.Ciphertext)
	fmt.Printf("Raw Full String: %s", s.String())
}

func (s *EncryptedUserSecretsAES256GCM) String() string {
	return string(s.Ciphertext)
}

func (s *EncryptedUserSecretsAES256GCM) NonceString() string {
	nonce, _ := s.Ciphertext[:12], s.Ciphertext[12:]
	return string(nonce)
}

func Encrypt(plaintext string) (EncryptedUserSecretsAES256GCM, error) {
	var encryptedSecret EncryptedUserSecretsAES256GCM

	secretKey := os.Getenv("USER_SEC_KEY")

	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		slog.Error("Error creating new cipher", "error", err.Error())
		return encryptedSecret, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		slog.Error("Error while attempting encrytion", "error", err.Error())
		return encryptedSecret, err
	}

	// We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
	// A nonce should always be randomly generated for every encryption.
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		panic(err)
	}

	// ciphertext here is actually nonce+ciphertext
	// So that when we decrypt, just knowing the nonce size
	// is enough to separate it from the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	encryptedSecret.Ciphertext = ciphertext

	return encryptedSecret, nil
}

func (s *EncryptedUserSecretsAES256GCM) Decrypt() (string, error) {
	plaintext, err := decrypt(s.String())
	if err != nil {
		slog.Error("Error decrypting secret", slog.String("error", err.Error()))
		return plaintext, err
	}
	return plaintext, err
}

func decrypt(ciphertext string) (string, error) {
	secretKey := os.Getenv("USER_SEC_KEY")

	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		slog.Error("Error creating new cipher", "error", err.Error())
		return "", err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		slog.Error("Error while attempting encrytion", "error", err.Error())
		return "", err
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		panic(err)
	}

	return string(plaintext), err
}
