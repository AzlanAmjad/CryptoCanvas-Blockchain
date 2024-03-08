package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignUsingPrivateKeyPublicKeyVerifyPass(t *testing.T) {
	// Generate a new private key and public key pair.
	privateKey := GeneratePrivateKey()
	publicKey := privateKey.GetPublicKey()

	// Sign the data using the private key.
	data := []byte("Hello, World!")
	signature, err := privateKey.Sign(data)
	if err != nil {
		t.Error(err)
	}

	// Verify the signature using the public key.
	assert.Equal(t, true, publicKey.Verify(data, signature))
}

func TestSignUsingPrivateKeyPublicKeyVerifyFail(t *testing.T) {
	// Generate a new private key.
	privateKey := GeneratePrivateKey()
	publicKey := privateKey.GetPublicKey()

	// Sign the data using the private key.
	data := []byte("Hello, World!")
	signature, err := privateKey.Sign(data)
	if err != nil {
		t.Error(err)
	}

	// Generate a new private key and public key pair.
	privateKey2 := GeneratePrivateKey()
	publicKey2 := privateKey2.GetPublicKey()

	// Verify the signature using the new public key.
	assert.Equal(t, false, publicKey2.Verify(data, signature))
	// Verify the signature on the wrong data using the original public key.
	assert.Equal(t, false, publicKey.Verify([]byte("Wrong Data!"), signature))
}
