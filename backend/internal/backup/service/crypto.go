package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const (
	encMagic     = "SBK1"
	encHeaderLen = 4 + aes.BlockSize // magic + IV
	encMACLen    = sha256.Size
	encMinLen    = encHeaderLen + encMACLen
)

func parseEncryptionKey(raw string) []byte {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if b, err := hex.DecodeString(s); err == nil && len(b) == 32 {
		return b
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b
	}
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}

func deriveEncKeys(key []byte) (aesKey, macKey []byte) {
	a := sha256.Sum256(append([]byte("starbyte-backup-aes:"), key...))
	m := sha256.Sum256(append([]byte("starbyte-backup-mac:"), key...))
	return a[:], m[:]
}

func isEncryptedBlob(head []byte) bool {
	return len(head) >= 4 && string(head[:4]) == encMagic
}

// encryptStream writes AES-256-CTR ciphertext plus HMAC-SHA256 (encrypt-then-MAC).
func encryptStream(dst io.Writer, src io.Reader, key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes")
	}
	aesKey, macKey := deriveEncKeys(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}
	mac := hmac.New(sha256.New, macKey)
	header := append([]byte(encMagic), iv...)
	if _, err := io.MultiWriter(dst, mac).Write(header); err != nil {
		return err
	}
	stream := cipher.NewCTR(block, iv)
	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf[:n], buf[:n])
			if _, err := io.MultiWriter(dst, mac).Write(buf[:n]); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	_, err = dst.Write(mac.Sum(nil))
	return err
}

// decryptStream verifies HMAC-SHA256 then AES-256-CTR decrypts. src may be a stream.
func decryptStream(dst io.Writer, src io.Reader, key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes")
	}
	header := make([]byte, encHeaderLen)
	if _, err := io.ReadFull(src, header); err != nil {
		return fmt.Errorf("读取加密头失败: %w", err)
	}
	if !isEncryptedBlob(header) {
		return fmt.Errorf("不是 StarByte AES 备份")
	}
	aesKey, macKey := deriveEncKeys(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, macKey)
	if _, err := mac.Write(header); err != nil {
		return err
	}
	stream := cipher.NewCTR(block, header[4:])
	window := make([]byte, 0, encMACLen+32*1024)
	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			window = append(window, buf[:n]...)
			if len(window) > encMACLen {
				plainN := len(window) - encMACLen
				chunk := window[:plainN]
				if _, err := mac.Write(chunk); err != nil {
					return err
				}
				stream.XORKeyStream(chunk, chunk)
				if _, err := dst.Write(chunk); err != nil {
					return err
				}
				window = append(window[:0], window[plainN:]...)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if len(window) != encMACLen {
		return fmt.Errorf("加密备份不完整")
	}
	if !hmac.Equal(window, mac.Sum(nil)) {
		return fmt.Errorf("加密备份 HMAC 校验失败")
	}
	return nil
}
