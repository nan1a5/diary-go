package config

import (
	"encoding/base64"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	JWTSecret      string
	AESKey         []byte
	DBDsn          string
	UploadDir          string
	JWTExpireHours     int
	EnableRegistration bool
}

func LoadConfig() *Config {
	_ = godotenv.Load() // 如果没有 .env 也 OK，优先环境变量
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "secret")
	aesBase64 := getEnv("AES_KEY_BASE64", "")
	dbDsn := getEnv("DB_DSN", "./data.db")
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	jwtExpireHours := toInt(getEnv("JWT_EXPIRE_HOURS", "72"))
	enableRegistration := getEnv("ENABLE_REGISTRATION", "true") == "true"

	var aesKey []byte
	if aesBase64 != "" {
		k, err := base64.StdEncoding.DecodeString(aesBase64)
		if err != nil {
			log.Fatalf("invalid AES_KEY_BASE64: %v", err)
		}
		if len(k) != 32 {
			log.Fatalf("AES key must be 32 bytes (AES-256)")
		}
		aesKey = k
	} else {
		// 警告：开发时可临时生成，但生产一定要设定 ENV
		log.Println("WARN: AES key not set. Diary content won't be encrypted.")
	}

	return &Config{
		Port:               port,
		JWTSecret:          jwtSecret,
		AESKey:             aesKey,
		DBDsn:              dbDsn,
		UploadDir:          uploadDir,
		JWTExpireHours:     jwtExpireHours,
		EnableRegistration: enableRegistration,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
