package utils

import (
	"crypto/rand"
	"math/big"
	"net/url"
	"path"

	"github.com/webpoint-solutions-llc/go-starter/internal/config"
)

func JoinS3URL(key string) (string, error) {
	u, err := url.Parse(config.Cfg.S3Endpoint)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, config.Cfg.S3Bucket, key)
	return u.String(), nil
}

var urlSafeCharset = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

var bigInt = big.NewInt(int64(len(urlSafeCharset)))

func GenerateSafeCode() (string, error) {
	code := make([]rune, 9) // 8 chars + 1 for the dash
	for i := 0; i < 9; i++ {
		if i == 4 {
			code[i] = '-' // insert dash in the middle
			continue
		}
		n, err := rand.Int(rand.Reader, bigInt)
		if err != nil {
			return "", err
		}
		code[i] = urlSafeCharset[n.Int64()]
	}
	return string(code), nil
}
