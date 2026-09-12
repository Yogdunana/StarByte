package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

func smtpSecretCipher() (cipher.AEAD, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(os.Getenv("STARBYTE_CONFIG_ENCRYPTION_KEY")))
	if err != nil || len(key) != 32 {
		return nil, errors.New("请配置有效的 STARBYTE_CONFIG_ENCRYPTION_KEY（Base64 编码的 32 字节密钥）")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// EncryptSMTPPassword authenticates the ciphertext and binds it to this setting.
func EncryptSMTPPassword(password string) (string, error) {
	aead, err := smtpSecretCipher()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", errors.New("无法生成 SMTP 密码加密随机数")
	}
	encrypted := aead.Seal(nonce, nonce, []byte(password), []byte(SMTPSettingsKey))
	return "v1:" + base64.StdEncoding.EncodeToString(encrypted), nil
}

func (s SMTPRuntime) Resolve(base EmailConfig) (EmailConfig, error) {
	cfg := s.Overlay(base).ApplyEnvPassword()
	if s.PasswordCiphertext == "" {
		return cfg, nil
	}
	aead, err := smtpSecretCipher()
	if err != nil {
		return EmailConfig{}, err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s.PasswordCiphertext, "v1:"))
	if err != nil || !strings.HasPrefix(s.PasswordCiphertext, "v1:") || len(raw) < aead.NonceSize()+aead.Overhead() {
		return EmailConfig{}, errors.New("SMTP 密码密文无效，请重新保存密码")
	}
	password, err := aead.Open(nil, raw[:aead.NonceSize()], raw[aead.NonceSize():], []byte(SMTPSettingsKey))
	if err != nil {
		return EmailConfig{}, errors.New("SMTP 密码无法解密，请检查配置密钥或重新保存密码")
	}
	cfg.Password = string(password)
	cfg.PasswordSource = "web"
	return cfg, nil
}
