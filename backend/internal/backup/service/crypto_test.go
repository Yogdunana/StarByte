package service

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEncryptionKey(t *testing.T) {
	assert.Nil(t, parseEncryptionKey(""))
	assert.Nil(t, parseEncryptionKey("   "))

	hexKey := strings.Repeat("ab", 32)
	got := parseEncryptionKey(hexKey)
	require.Len(t, got, 32)
	assert.Equal(t, byte(0xab), got[0])

	pass := parseEncryptionKey("passphrase-not-a-raw-key")
	require.Len(t, pass, 32)
	assert.NotEqual(t, got, pass)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := parseEncryptionKey("unit-test-backup-key")
	plain := bytes.Repeat([]byte("starbyte gzip payload\n"), 200)
	var enc bytes.Buffer
	require.NoError(t, encryptStream(&enc, bytes.NewReader(plain), key))
	assert.True(t, isEncryptedBlob(enc.Bytes()))
	assert.Greater(t, enc.Len(), len(plain))

	var dec bytes.Buffer
	require.NoError(t, decryptStream(&dec, bytes.NewReader(enc.Bytes()), key))
	assert.Equal(t, plain, dec.Bytes())
}

func TestDecryptRejectsTamperAndWrongKey(t *testing.T) {
	key := parseEncryptionKey("unit-test-backup-key")
	var enc bytes.Buffer
	require.NoError(t, encryptStream(&enc, strings.NewReader("secret dump"), key))
	blob := enc.Bytes()
	blob[len(blob)/2] ^= 0x01
	err := decryptStream(io.Discard, bytes.NewReader(blob), key)
	require.Error(t, err)

	var good bytes.Buffer
	require.NoError(t, encryptStream(&good, strings.NewReader("secret dump"), key))
	err = decryptStream(io.Discard, bytes.NewReader(good.Bytes()), parseEncryptionKey("other-key"))
	require.Error(t, err)
}

func TestDecryptRejectsShortBlob(t *testing.T) {
	key := parseEncryptionKey("unit-test-backup-key")
	err := decryptStream(io.Discard, bytes.NewReader([]byte("SBK1short")), key)
	require.Error(t, err)
}
