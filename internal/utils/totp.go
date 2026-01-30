package utils

import (
	"image"
	"log/slog"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
)

func GenerateTOTP(email string) (string, string, image.Image, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.Cfg.AppName,
		AccountName: email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		slog.Error("TOTP Image generation error", "err", err)
		return "", "", nil, err
	}

	img, err := key.Image(256, 256)
	if err != nil {
		slog.Error("TOTP Image generation error", "err", err)
		return "", "", nil, err
	}
	return key.URL(), key.Secret(), img, err
}

func ValidateTOTP(code, secret string) bool {
	// THIS SHOULD MATCH THE GENRATION SETTINGS
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false
	}
	return valid
}
