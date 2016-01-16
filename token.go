package hub

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Tokenizer interface {
	Tokenize(username string) string
}

type TokenizerFunc func(username string) string

func (tf TokenizerFunc) Tokenize(username string) string {
	return tf(username)
}

func HmacSha256Tokenizer(secret string) Tokenizer {
	return TokenizerFunc(func(username string) string {
		hasher := hmac.New(sha256.New, []byte(secret))
		hasher.Write([]byte(username))
		return hex.EncodeToString(hasher.Sum(nil))
	})
}
