package keygen_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/keygen"
	"github.com/stretchr/testify/require"
)

func TestKeyGen(t *testing.T) {
	tempDir := os.TempDir()

	pubKeyFile, err := os.CreateTemp(tempDir, "*")
	require.NoError(t, err)
	defer os.Remove(pubKeyFile.Name())

	privKeyFile, err := os.CreateTemp(tempDir, "*")
	require.NoError(t, err)
	defer os.Remove(privKeyFile.Name())

	defer os.RemoveAll(tempDir)

	err = keygen.Create(pubKeyFile, privKeyFile)
	require.NoError(t, err)

	pubKeyRestored, err := keygen.ReadPubKey(pubKeyFile.Name())
	require.NoError(t, err)

	privKeyRestored, err := keygen.ReadPrivKey(privKeyFile.Name())
	require.NoError(t, err)

	testBytes := []byte("test string to check encription/decryption")

	ciphertext, err := keygen.EncryptWithPublicKey(testBytes, pubKeyRestored)
	require.NoError(t, err)

	res, err := keygen.DecryptWithPrivateKey(ciphertext, privKeyRestored)
	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(testBytes, res))
}
