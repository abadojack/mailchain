package aes256cbc

import (
	"github.com/mailchain/mailchain/internal/pkg/crypto/cipher"
	"github.com/pkg/errors"
)

// bytesEncode encode the encrypted data to the hex format
func bytesEncode(data *encryptedData) ([]byte, error) {
	if err := data.verify(); err != nil {
		return nil, errors.WithMessage(err, "encrypted data is invalid")
	}
	compressedKey, err := compress(data.EphemeralPublicKey)
	if err != nil {
		return nil, errors.WithMessage(err, "could not compress EphemeralPublicKey")
	}
	encodedData := make([]byte, 1+len(data.InitializationVector)+len(compressedKey)+len(data.MessageAuthenticationCode)+len(data.Ciphertext))
	encodedData[0] = cipher.AES256CBC
	copy(encodedData[1:], data.InitializationVector)
	copy(encodedData[1+len(data.InitializationVector):], compressedKey)
	copy(encodedData[1+len(data.InitializationVector)+len(compressedKey):], data.MessageAuthenticationCode)
	copy(encodedData[1+len(data.InitializationVector)+len(compressedKey)+len(data.MessageAuthenticationCode):], data.Ciphertext)
	return encodedData, nil
}

// bytesDecode convert the hex format in to the encrypted data format
func bytesDecode(raw []byte) (*encryptedData, error) {
	macLen := 32
	ivLen := 16
	if len(raw) == 0 {
		return nil, errors.Errorf("raw must not be empty")
	}
	if len(raw) < macLen+ivLen+pubKeyBytesLenCompressed+2 {
		return nil, errors.Errorf("raw data does not have enough bytes to be encoded")
	}
	if raw[0] != cipher.AES256CBC {
		return nil, errors.Errorf("invalid prefix")
	}
	raw = raw[1:]
	decompressedKey := decompress(raw[ivLen : ivLen+pubKeyBytesLenCompressed])
	// iv and mac must be created this way to ensure the cap of the array is not different
	iv := make([]byte, ivLen)
	copy(iv, raw[:ivLen])
	mac := make([]byte, macLen)
	copy(mac, raw[ivLen+pubKeyBytesLenCompressed:ivLen+pubKeyBytesLenCompressed+macLen])

	ret := &encryptedData{
		InitializationVector:      iv,
		EphemeralPublicKey:        decompressedKey,
		MessageAuthenticationCode: mac,
		Ciphertext:                raw[ivLen+pubKeyBytesLenCompressed+macLen:],
	}
	if err := ret.verify(); err != nil {
		return nil, errors.WithMessage(err, "encrypted data is invalid")
	}

	return ret, nil
}
