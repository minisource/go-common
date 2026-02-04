package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
)

// ============================================
// Password Hashing
// ============================================

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// HashPasswordWithCost hashes with custom cost
func HashPasswordWithCost(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword compares password with hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ============================================
// SHA Hashing
// ============================================

// SHA256Hash computes SHA-256 hash
func SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA512Hash computes SHA-512 hash
func SHA512Hash(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256HashBytes computes SHA-256 hash from bytes
func SHA256HashBytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ============================================
// HMAC
// ============================================

// HMACSign creates HMAC-SHA256 signature
func HMACSign(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// HMACSignBase64 creates HMAC-SHA256 signature as base64
func HMACSignBase64(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// HMACVerify verifies HMAC-SHA256 signature
func HMACVerify(message, signature, secret string) bool {
	expectedSig := HMACSign(message, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// ============================================
// AES Encryption
// ============================================

// Encryptor handles AES encryption/decryption
type Encryptor struct {
	key []byte
}

// NewEncryptor creates an AES encryptor with 32-byte key
func NewEncryptor(key string) (*Encryptor, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("%w: key must be 32 bytes", ErrInvalidKey)
	}
	return &Encryptor{key: keyBytes}, nil
}

// NewEncryptorFromHex creates encryptor from hex-encoded key
func NewEncryptorFromHex(hexKey string) (*Encryptor, error) {
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("%w: key must be 32 bytes", ErrInvalidKey)
	}
	return &Encryptor{key: keyBytes}, nil
}

// Encrypt encrypts plaintext using AES-GCM
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// ============================================
// Token Generation
// ============================================

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateRandomString generates random hex string
func GenerateRandomString(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length / 2)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSecureToken generates a secure token
func GenerateSecureToken(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateAPIKey generates an API key with prefix
func GenerateAPIKey(prefix string) (string, error) {
	token, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	key := base64.RawURLEncoding.EncodeToString(token)
	if prefix != "" {
		return prefix + "_" + key, nil
	}
	return key, nil
}

// ============================================
// OTP Generation
// ============================================

// GenerateOTP generates numeric OTP
func GenerateOTP(length int) (string, error) {
	if length <= 0 || length > 10 {
		length = 6
	}

	const digits = "0123456789"
	result := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		result[i] = digits[int(randomBytes[i])%len(digits)]
	}

	return string(result), nil
}

// GenerateAlphanumericCode generates alphanumeric code
func GenerateAlphanumericCode(length int) (string, error) {
	if length <= 0 {
		length = 8
	}

	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		result[i] = chars[int(randomBytes[i])%len(chars)]
	}

	return string(result), nil
}

// ============================================
// Key Derivation
// ============================================

// DeriveKey derives a key from password using SHA-256
func DeriveKey(password, salt string) []byte {
	combined := password + salt
	hash := sha256.Sum256([]byte(combined))
	return hash[:]
}

// ============================================
// Base64 Helpers
// ============================================

// Base64Encode encodes to standard base64
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode decodes standard base64
func Base64Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// Base64URLEncode encodes to URL-safe base64
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes URL-safe base64
func Base64URLDecode(encoded string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(encoded)
}

// ============================================
// Utility Functions
// ============================================

// SecureCompare performs constant-time comparison
func SecureCompare(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}

// MaskString masks sensitive data
func MaskString(s string, visibleChars int) string {
	if len(s) <= visibleChars*2 {
		return strings.Repeat("*", len(s))
	}
	return s[:visibleChars] + strings.Repeat("*", len(s)-visibleChars*2) + s[len(s)-visibleChars:]
}

// MaskEmail masks email address
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return MaskString(email, 2)
	}
	localPart := parts[0]
	domain := parts[1]
	if len(localPart) <= 2 {
		return localPart + "@" + domain
	}
	return localPart[:1] + strings.Repeat("*", len(localPart)-2) + localPart[len(localPart)-1:] + "@" + domain
}

// MaskPhone masks phone number
func MaskPhone(phone string) string {
	if len(phone) <= 4 {
		return strings.Repeat("*", len(phone))
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}
