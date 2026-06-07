// Package auth 放置认证相关的底层安全工具。
//
// 这里不直接处理 HTTP，也不决定用户是否有权限；只负责 token、密码哈希等可复用能力。
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// NewToken 生成浏览器会话或邀请链接使用的随机 token。
// token 原文只返回给客户端或邀请链接，服务端持久化时只保存 HashToken 的结果。
func NewToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// HashToken 将 token 原文转换为稳定 hash，避免数据库泄漏时直接暴露可用 token。
func HashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

// HashPassword 使用 PBKDF2-SHA256 生成密码哈希。
// 当前不引入额外密码库，格式中保留算法、迭代次数、salt 和 hash，便于后续迁移。
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	const iterations = 210000
	derived := pbkdf2SHA256([]byte(password), salt, iterations, 32)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, base64.RawURLEncoding.EncodeToString(salt), base64.RawURLEncoding.EncodeToString(derived)), nil
}

// VerifyPassword 校验用户输入密码和已保存哈希是否匹配。
// 比较 hash 时使用 constant-time compare，避免把匹配进度暴露给时序攻击。
func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	salt, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

// pbkdf2SHA256 是本地 PBKDF2 实现，用于避免 MVP 阶段额外引入密码依赖。
func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	var derived []byte
	var blockIndex uint32 = 1
	for len(derived) < keyLen {
		block := pbkdf2Block(password, salt, iterations, blockIndex)
		derived = append(derived, block...)
		blockIndex++
	}
	return derived[:keyLen]
}

// pbkdf2Block 生成 PBKDF2 的单个 block，并按规范对多轮 HMAC 结果做 XOR。
func pbkdf2Block(password, salt []byte, iterations int, blockIndex uint32) []byte {
	indexBytes := []byte{byte(blockIndex >> 24), byte(blockIndex >> 16), byte(blockIndex >> 8), byte(blockIndex)}
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)
	mac.Write(indexBytes)
	u := mac.Sum(nil)
	output := append([]byte(nil), u...)

	for i := 1; i < iterations; i++ {
		mac = hmac.New(sha256.New, password)
		mac.Write(u)
		u = mac.Sum(nil)
		for j := range output {
			output[j] ^= u[j]
		}
	}
	return output
}
