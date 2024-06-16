// Package keygen contains public/private RSA 4096 keys generator
package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"fmt"
	"io"
	"os"
)

func Create(publicKeyFile io.Writer, privateKeyFile io.Writer) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("geneate key err %w", err)
	}
	pubKey := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)

	if _, err := publicKeyFile.Write(pubKey); err != nil {
		return fmt.Errorf("write public key err %w", err)
	}

	privKey := x509.MarshalPKCS1PrivateKey(privateKey)

	if _, err := privateKeyFile.Write(privKey); err != nil {
		return fmt.Errorf("write private key err %w", err)
	}
	return nil
}

func ReadPubKey(publicKeyFileName string) (*rsa.PublicKey, error) {
	pubKeyFile, err := os.Open(publicKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("open %s err %w", publicKeyFileName, err)
	}
	defer pubKeyFile.Close()

	pubKeyBytes, err := io.ReadAll(pubKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read from %s err %w", publicKeyFileName, err)
	}

	res, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("can't parse pkcs1 public key from %s err %w", publicKeyFileName, err)
	}

	return res, err
}

func ReadPrivKey(privateKeyFileName string) (*rsa.PrivateKey, error) {
	privKeyFile, err := os.Open(privateKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("open %s err %w", privateKeyFileName, err)
	}
	defer privKeyFile.Close()

	privKeyBytes, err := io.ReadAll(privKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read from %s err %w", privateKeyFileName, err)
	}

	res, err := x509.ParsePKCS1PrivateKey(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("can't parse pkcs1 public key from %s err %w", privateKeyFileName, err)
	}

	return res, err
}

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()
	return rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	return rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
}
