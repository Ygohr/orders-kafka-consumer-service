package kafka

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

type XDGSCRAMClient struct {
	HashGeneratorFcn func() hash.Hash
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	return "", nil
}

func (x *XDGSCRAMClient) Done() bool {
	return true
}

func SHA256() hash.Hash {
	return sha256.New()
}

func SHA512() hash.Hash {
	return sha512.New()
}

func HMAC(key []byte, data []byte, hashFn func() hash.Hash) []byte {
	h := hmac.New(hashFn, key)
	h.Write(data)
	return h.Sum(nil)
}
