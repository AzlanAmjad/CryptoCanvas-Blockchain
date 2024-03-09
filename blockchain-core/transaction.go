package core

import (
	"fmt"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
)

type Transaction struct {
	Data []byte

	From      crypto.PublicKey
	Signature *crypto.Signature
}

func (t *Transaction) Sign(privateKey *crypto.PrivateKey) error {
	signature, err := privateKey.Sign(t.Data)
	if err != nil {
		return err
	}

	t.Signature = signature
	t.From = privateKey.GetPublicKey()

	return nil
}

func (t *Transaction) VerifySignature() (bool, error) {
	if t.Signature == nil {
		return false, fmt.Errorf("no signature to verify")
	}

	return t.From.Verify(t.Data, t.Signature), nil
}
