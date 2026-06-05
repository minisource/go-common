//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// OTPFromRedis reads the OTP code stored by auth service (dev/e2e only).
func OTPFromRedis(t *testing.T, target, otpType string) string {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		fmt.Sscanf(v, "%d", &db)
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("otp:%s:%s", otpType, target)
	data, err := rdb.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("redis get %s: %v", key, err)
	}
	var otp struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(data), &otp); err != nil {
		t.Fatalf("parse otp json: %v", err)
	}
	if otp.Code == "" {
		t.Fatalf("empty otp code in redis key %s", key)
	}
	return otp.Code
}
