package netpipe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
)

func makeReader(key string, reader io.Reader) io.Reader {
	hash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(hash[0:16]) // aes128
	if err != nil {
		log.Fatal(err)
	}

	stream := cipher.NewCTR(block, hash[16:32])
	return &cipher.StreamReader{S: stream, R: reader}
}

func makeWriter(key string, writer io.Writer) io.Writer {
	hash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(hash[0:16]) // aes128
	if err != nil {
		log.Fatal(err)
	}

	stream := cipher.NewCTR(block, hash[16:32])
	return &cipher.StreamWriter{S: stream, W: writer}
}

func genKey() string {
	data := make([]byte, 6)
	_, _ = rand.Read(data)
	return base64.StdEncoding.EncodeToString(data)
}
