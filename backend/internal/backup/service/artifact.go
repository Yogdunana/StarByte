package service

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type unwrapInfo struct {
	Encrypted  bool
	DecryptOK  bool
	GzipOK     bool
	ChecksumOK bool
	SizeBytes  int64
}

func peekHead(f *os.File, n int) ([]byte, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	buf := make([]byte, n)
	got, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	if _, seekErr := f.Seek(0, io.SeekStart); seekErr != nil {
		return nil, seekErr
	}
	return buf[:got], nil
}

// unwrapStored opens a downloaded artifact: checksum, optional AES, then gzip.
// On success the returned reader owns f (and any decrypt temp) and must be Closed.
func unwrapStored(f *os.File, wantChecksum string, key []byte) (io.ReadCloser, unwrapInfo, error) {
	info := unwrapInfo{}
	st, err := f.Stat()
	if err != nil {
		return nil, info, err
	}
	info.SizeBytes = st.Size()
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, info, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, info, err
	}
	info.ChecksumOK = hex.EncodeToString(h.Sum(nil)) == wantChecksum
	if !info.ChecksumOK {
		return nil, info, errChecksum()
	}

	head, err := peekHead(f, 4)
	if err != nil {
		return nil, info, err
	}
	info.Encrypted = isEncryptedBlob(head)

	owned := []*os.File{f}
	var gzSrc io.Reader = f
	if info.Encrypted {
		if len(key) == 0 {
			return nil, info, errDecrypt("备份已加密但未配置 STARBYTE_BACKUP_ENCRYPTION_KEY")
		}
		dec, err := os.CreateTemp("", "starbyte-decrypt-*.gz")
		if err != nil {
			return nil, info, err
		}
		if err := decryptStream(dec, f, key); err != nil {
			_ = dec.Close()
			_ = os.Remove(dec.Name())
			return nil, info, errDecrypt("解密失败: " + err.Error())
		}
		if _, err := dec.Seek(0, io.SeekStart); err != nil {
			_ = dec.Close()
			_ = os.Remove(dec.Name())
			return nil, info, err
		}
		info.DecryptOK = true
		gzSrc = dec
		owned = append(owned, dec)
	} else {
		info.DecryptOK = true
	}

	gz, err := gzip.NewReader(gzSrc)
	if err != nil {
		for _, extra := range owned[1:] {
			name := extra.Name()
			_ = extra.Close()
			_ = os.Remove(name)
		}
		return nil, info, err
	}
	info.GzipOK = true
	return &unwrappedDump{gz: gz, files: owned}, info, nil
}

type unwrappedDump struct {
	gz    *gzip.Reader
	files []*os.File
}

func (u *unwrappedDump) Read(p []byte) (int, error) { return u.gz.Read(p) }

func (u *unwrappedDump) Close() error {
	err := u.gz.Close()
	for _, f := range u.files {
		name := f.Name()
		_ = f.Close()
		_ = os.Remove(name)
	}
	return err
}
