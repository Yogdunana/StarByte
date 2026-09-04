package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
)

// uploadToMinIO 使用 S3 兼容 API（AWS Signature V4）上传对象到 MinIO
// 这是一个不依赖第三方 SDK 的极简实现，仅支持 PUT 操作
func uploadToMinIO(cfg *config.MinIOConfig, objectName string, data []byte, contentType string) error {
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		return fmt.Errorf("MinIO endpoint or bucket not configured")
	}

	// 构建 URL
	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, cfg.Endpoint, cfg.Bucket, objectName)

	// 创建请求
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// 添加 AWS Signature V4 认证
	signRequest(req, cfg.AccessKey, cfg.SecretKey, cfg.Bucket, objectName, data)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload to MinIO: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MinIO upload failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// signRequest 为 HTTP 请求添加 AWS Signature V4 认证头
func signRequest(req *http.Request, accessKey, secretKey, bucket, objectKey string, payload []byte) {
	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	timestamp := now.Format("20060102T150405Z")

	// 1. 计算 payload hash
	payloadHash := sha256Hex(payload)

	// 2. 设置必要头部
	req.Header.Set("x-amz-date", timestamp)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("Host", req.URL.Host)

	// 3. 构建规范请求
	canonicalURI := fmt.Sprintf("/%s/%s", bucket, objectKey)
	canonicalQueryString := ""

	// 收集和排序头部
	signedHeaders := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	sort.Strings(signedHeaders)

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		req.URL.Host, payloadHash, timestamp)
	signedHeadersStr := strings.Join(signedHeaders, ";")

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		"PUT", canonicalURI, canonicalQueryString,
		canonicalHeaders, signedHeadersStr, payloadHash)

	// 4. 构建待签字符串
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStr, "us-east-1", "s3")
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		timestamp, credentialScope, sha256Hex([]byte(canonicalRequest)))

	// 5. 计算签名
	signingKey := getSigningKey(secretKey, dateStr, "us-east-1", "s3")
	signature := hex.EncodeToString(hmacSha256(signingKey, []byte(stringToSign)))

	// 6. 构建 Authorization 头
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeadersStr, signature)
	req.Header.Set("Authorization", authHeader)
}

// sha256Hex 计算数据的 SHA256 哈希并返回十六进制字符串
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// hmacSha256 使用 key 对 data 进行 HMAC-SHA256
func hmacSha256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// getSigningKey 派生签名密钥
func getSigningKey(secretKey, dateStr, region, service string) []byte {
	kDate := hmacSha256([]byte("AWS4"+secretKey), []byte(dateStr))
	kRegion := hmacSha256(kDate, []byte(region))
	kService := hmacSha256(kRegion, []byte(service))
	kSigning := hmacSha256(kService, []byte("aws4_request"))
	return kSigning
}
