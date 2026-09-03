package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// AccessTokenType identifies access tokens.
	AccessTokenType = "access"
	// RefreshTokenType identifies refresh tokens.
	RefreshTokenType = "refresh"
)

// context keys used to store authenticated user information.
const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyTokenID  = "token_id"
)

// Claims is the custom JWT claims object. It embeds jwt.RegisteredClaims so
// that standard fields such as ID (jti), Issuer (iss) and ExpiresAt (exp) are
// available.
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	TokenType   string   `json:"token_type"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed access token for the given user.
func GenerateAccessToken(userID string, username string, roles, permissions []string, cfg *config.JWTConfig) (string, time.Time, error) {
	now := time.Now()
	accessExp := now.Add(time.Duration(cfg.AccessTokenExp) * time.Second)

	accessClaims := &Claims{
		UserID:      userID,
		Username:    username,
		TokenType:   AccessTokenType,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
		},
	}

	accessToken, err := signToken(accessClaims, cfg.Secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return accessToken, accessExp, nil
}

func signToken(claims *Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken validates and parses a token string, returning the claims.
// It verifies the signature, expiration time, and issuer.
func ParseToken(tokenString string, cfg *config.JWTConfig) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 验证 Issuer
	if claims.Issuer != cfg.Issuer {
		return nil, fmt.Errorf("invalid token issuer: %s", claims.Issuer)
	}

	return claims, nil
}

// JWTAuth is a gin middleware that validates a Bearer access token from the
// Authorization header and stores the user id and username in the context.
// If a Redis client is provided, it also checks the token blacklist to ensure
// logged-out tokens are immediately invalidated.
func JWTAuth(cfg *config.JWTConfig, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortUnauthorized(c, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortUnauthorized(c, "invalid Authorization header")
			return
		}

		claims, err := ParseToken(parts[1], cfg)
		if err != nil {
			abortUnauthorized(c, "invalid or expired token")
			return
		}

		if claims.TokenType != AccessTokenType {
			abortUnauthorized(c, "invalid token type")
			return
		}

		// Check token blacklist (if Redis is available)
		if rdb != nil && claims.ID != "" {
			blacklistKey := fmt.Sprintf("auth:blacklist:%s", claims.ID)
			n, err := rdb.Exists(c.Request.Context(), blacklistKey).Result()
			if err != nil {
				// Redis error: fail open (allow request through) but log
				_ = n
			} else if n > 0 {
				abortUnauthorized(c, "token has been revoked")
				return
			}
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyTokenID, claims.ID)
		c.Next()
	}
}

func abortUnauthorized(c *gin.Context, msg string) {
	response.Error(c, response.NewUnauthorizedError(msg))
	c.Abort()
}

// GetUserID retrieves the authenticated user id stored in the context.
func GetUserID(c *gin.Context) string {
	return c.GetString(ContextKeyUserID)
}

// GetUsername retrieves the authenticated username stored in the context.
func GetUsername(c *gin.Context) string {
	return c.GetString(ContextKeyUsername)
}

// GetTokenID retrieves the token id (jti) stored in the context, if any.
func GetTokenID(c *gin.Context) string {
	return c.GetString(ContextKeyTokenID)
}
