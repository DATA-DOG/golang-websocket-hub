package wshub

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

type Tokenizer interface {
	Tokenize(username string) string
}

type TokenizerFunc func(username string) string

func (tf TokenizerFunc) Tokenize(username string) string {
	return tf(username)
}

func HmacSha512Tokenizer(secret string) Tokenizer {
	return TokenizerFunc(func(username string) string {
		hasher := hmac.New(sha512.New, []byte(secret))
		hasher.Write([]byte(username))
		return hex.EncodeToString(hasher.Sum(nil))
	})
}
