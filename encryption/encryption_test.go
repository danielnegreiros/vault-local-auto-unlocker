package encryption_test

import (
	"log"
	"testing"
	"vault-unlocker/encryption"

	"github.com/stretchr/testify/assert"
)

func TestEnc(t *testing.T) {

	cryp, err := encryption.NewCrypto("./")
	assert.NoError(t, err)

	val, err := cryp.Encrypt("something")
	assert.NoError(t, err)
	log.Println(val)

}
