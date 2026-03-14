package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type Config struct {
	Port               string
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	MongoURI           string
	MongoDatabase      string
	R2AccountID        string
	R2AccessKeyID      string
	R2SecretAccessKey  string
	R2BucketName       string
	R2PublicURL        string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "6868"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3000/auth/google/callback"),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:3000/auth/github/callback"),
		MongoURI:           getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:      getEnv("MONGO_DATABASE", "aeshield"),
		R2AccountID:        getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:      getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:  getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:       getEnv("R2_BUCKET_NAME", ""),
		R2PublicURL:        getEnv("R2_PUBLIC_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
