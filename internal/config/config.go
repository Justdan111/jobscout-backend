package config

import "os"

type Config struct {
	Port          string
	DataDir       string
	AnthropicKey  string
	Model         string
	ResendKey     string
	DigestTo      string
	DigestFrom    string
	AllowedOrigin string
}

func Load() Config {
	return Config{
		Port:          env("PORT", "8080"),
		DataDir:       env("DATA_DIR", "./data"),
		AnthropicKey:  os.Getenv("ANTHROPIC_API_KEY"),
		Model:         env("JOBSCOUT_MODEL", "claude-haiku-4-5-20251001"),
		ResendKey:     os.Getenv("RESEND_API_KEY"),
		DigestTo:      os.Getenv("DIGEST_TO"),
		DigestFrom:    env("DIGEST_FROM", "JobScout <onboarding@resend.dev>"),
		AllowedOrigin: env("ALLOWED_ORIGIN", "http://localhost:3000"),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
