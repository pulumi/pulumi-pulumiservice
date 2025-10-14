package examples

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"
)

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}
	return cwd
}

func generateRandomFiveDigits() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(100000))
	return fmt.Sprintf("%05d", n.Int64())
}
