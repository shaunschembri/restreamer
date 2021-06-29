package restream

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/shaunschembri/restreamer/pkg/restream/request"
)

type decrypter interface {
	decrypt([]byte) ([]byte, error)
	init(ctx context.Context) error
	info() string
}

type aes128 struct {
	iv         string
	keyURL     string
	bufferSize int
	request    request.Request
	mode       cipher.BlockMode
}

func (a aes128) info() string {
	return "AES128"
}

func (a *aes128) init(ctx context.Context) error {
	keyFileResponse, err := a.request.Do(ctx, a.keyURL)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer keyFileResponse.Body.Close()

	key, err := io.ReadAll(keyFileResponse.Body)
	if err != nil {
		return fmt.Errorf("cannot read key from %s: %w", a.keyURL, err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("cannot create new AES cipher: %w", err)
	}

	iv, err := hex.DecodeString(fmt.Sprintf("%032s", strings.ReplaceAll(a.iv, "0x", "")))
	if err != nil {
		return fmt.Errorf("cannot decode IV %s: %w", iv, err)
	}

	a.mode = cipher.NewCBCDecrypter(block, iv)

	return nil
}

func (a *aes128) decrypt(payload []byte) ([]byte, error) {
	payloadSize := len(payload)
	if payloadSize%aes.BlockSize != 0 {
		return nil, fmt.Errorf("payload size is not a multiple of %d", aes.BlockSize)
	}

	// Decrypt payload
	a.mode.CryptBlocks(payload, payload)

	// If last byte in payload is bigger then the block size (16) or the last byte
	// is not a multiple of 4, then there is no padding and the entire decrypted payload
	// is returned.
	lastByte := int(payload[payloadSize-1])
	if lastByte > aes.BlockSize || lastByte%4 != 0 {
		return payload, nil
	}

	// As per PKCS#7 specification the last byte value will contain the number of bytes
	// that were added as padding and all these extra bytes will have the same value as the
	// last byte.
	for _, paddingByte := range payload[payloadSize-lastByte:] {
		if paddingByte != byte(lastByte) {
			return payload, nil
		}
	}

	return payload[:payloadSize-lastByte], nil
}
