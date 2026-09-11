package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/google/uuid"
)

func newCheckinSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func signCheckinToken(secret string, activityID uuid.UUID, nonce int64, exp time.Time) string {
	expUnix := exp.Unix()
	payload := fmt.Sprintf("%s|%d|%d", activityID.String(), nonce, expUnix)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return fmt.Sprintf("v1.%d.%d.%s", nonce, expUnix, hex.EncodeToString(mac.Sum(nil)))
}

func verifyCheckinToken(a *model.Activity, token string, now time.Time) error {
	if a == nil || strings.TrimSpace(token) == "" || a.CheckinSecret == "" {
		return fmt.Errorf("missing token")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 4 || parts[0] != "v1" {
		return fmt.Errorf("malformed token")
	}
	nonce, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("malformed nonce")
	}
	expUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return fmt.Errorf("malformed exp")
	}
	if nonce != a.CheckinNonce {
		return fmt.Errorf("rotated")
	}
	if now.Unix() > expUnix {
		return fmt.Errorf("expired")
	}
	expected := signCheckinToken(a.CheckinSecret, a.ID, nonce, time.Unix(expUnix, 0))
	if !hmac.Equal([]byte(expected), []byte(token)) {
		return fmt.Errorf("bad mac")
	}
	return nil
}
