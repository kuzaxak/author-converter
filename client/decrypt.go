package client

import (
	"strconv"
	"unicode/utf8"
)

func decryptKey(readerSecret, userId string) string {
	return reverse(readerSecret) + "@_@" + userId
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func decrypt(text []byte, key string) ([]byte, error) {
	eTextUn, err := strconv.Unquote("\"" + string(text) + "\"")
	if err != nil {
		return nil, err
	}
	runes := []rune(eTextUn)

	var decryptedText []byte
	var runeBuf [utf8.UTFMax]byte

	for i := 0; i < len(runes); i++ {
		in := int(runes[i]) ^ int(key[i%len(key)])
		if in < utf8.RuneSelf {
			decryptedText = append(decryptedText, byte(in))
		} else {
			n := utf8.EncodeRune(runeBuf[:], rune(in))
			decryptedText = append(decryptedText, runeBuf[:n]...)
		}
	}

	return decryptedText, nil
}
