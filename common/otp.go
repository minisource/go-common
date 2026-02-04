package common

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

type OtpConfig struct {
	ExpireTime time.Duration `env:"OTP_EXPIRE_TIME"`
	Digits     int           `env:"OTP_DIGITS"`
	Limiter    time.Duration `env:"OTP_LIMITER"`
}

func (cfg OtpConfig) GenerateOtp() string {
	rand.Seed(time.Now().UnixNano())
	min := int(math.Pow(10, float64(cfg.Digits-1)))   // 10^d-1 100000
	max := int(math.Pow(10, float64(cfg.Digits)) - 1) // 999999 = 1000000 - 1 (10^d) -1

	var num = rand.Intn(max-min) + min
	return strconv.Itoa(num)
}
