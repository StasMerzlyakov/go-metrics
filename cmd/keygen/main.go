// Package main contains utility for key generation
package main

import (
	"os"

	"github.com/StasMerzlyakov/go-metrics/internal/keygen"
)

const (
	publicKeyFileName  = "./public.key"
	privateKeyFileName = "./private.key"
)

func main() {

	privKeyFile, err := os.Create(privateKeyFileName)
	if err != nil {
		panic(err)
	}
	defer privKeyFile.Close()

	pubKeyFile, err := os.Create(publicKeyFileName)
	if err != nil {
		panic(err)
	}
	defer privKeyFile.Close()

	if err := keygen.Create(pubKeyFile, privKeyFile); err != nil {
		panic(err)
	}
}
