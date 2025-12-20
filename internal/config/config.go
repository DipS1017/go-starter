package config

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/caarlos0/env/v11"
)

var Cfg Config

type Config struct {
	AppName       string `env:"APP_NAME"    envDefault:"go-starter-api"`
	Port          string `env:"PORT"        envDefault:"8080"`
	AppEnv        string `env:"APP_ENV"     envDefault:"local"`
	LogLevel      string `env:"LOG_LEVEL"   envDefault:"debug"`
	PostgresqlURL string `env:"POSTGRESQL_URL"   envDefault:""`
	DBHost        string `env:"DB_HOST"`
	DBDatabase    string `env:"DB_DATABASE"`
	DBUsername    string `env:"DB_USERNAME"`
	DBPassword    string `env:"DB_PASSWORD"`
	DBSchema      string `env:"DB_SCHEMA"`
	DBPort        int    `env:"DB_PORT"`

	// Oauth2 and JWT fields
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURI  string `env:"GOOGLE_REDIRECT_URI"`
	GoogleUserinfoURI  string `env:"GOOGLE_USERINFO_URI"`

	// Apple Oauth2 and JWT fields
	AppleTeamID   string `env:"APPLE_TEAM_ID"`
	AppleClientID string `env:"APPLE_CLIENT_ID"`
	AppleKeyID    string `env:"APPLE_KEY_ID"`

	AppleCertificatePath string `env:"APPLE_CERTIFICATE_PATH"`

	AppleRedirectURI string `env:"APPLE_REDIRECT_URI"`

	JWTSecret             string   `env:"JWT_SECRET_KEY"`
	SecretKeyDuration     int64    `env:"SECRET_KEY_EXPIRES_IN" envDefault:"3600"`
	AccessTokenDuration   int64    `env:"JWT_ACCESS_EXPIRES_IN" envDefault:"3600"`
	RefreshTokenDuration  int64    `env:"JWT_REFRESH_EXPIRES_IN" envDefault:"604800"`
	PasswordResetDuration int64    `env:"PASSWORD_RESET_EXPIRES_IN" envDefault:"604800"`
	EmailVerifyDuration   int64    `env:"PASSWORD_RESET_EXPIRES_IN" envDefault:"604800"`
	AllowOrigins          []string `env:"ALLOWED_ORIGINS" envSeparator:","`

	// Email fields
	EmailSender  string `env:"EMAIL_SENDER"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPUser     string `env:"SMTP_USER"`
	SMTPServer   string `env:"SMTP_SERVER"`
	SMTPPort     string `env:"SMTP_PORT"`

	FrontendURL string `env:"APP_FRONTEND_URL"`

	BackendURL string `env:"APP_BACKEND_URL"`

	APIKey string `env:"API_KEY"`

	MaxWrongPasswordAttempt int64 `env:"MAX_WRONG_PASSWORD_ATTEMPT" envDefault:"5"`

	// S3 fields
	S3Region   string `env:"S3_REGION"`
	S3Endpoint string `env:"S3_ENDPOINT"`
	S3Bucket   string `env:"S3_BUCKET"`

	S3SignedURLDuration string `env:"S3_SIGNED_URL_DURATION" envDefault:"3600s"`

	IsMinio bool `env:"IS_MINIO" envDefault:"false"`

	StripeWebhookSecret string `env:"STRIPE_WEBHOOK_SECRET"`
	StripeWebhookURL    string `env:"STRIPE_WEBHOOK_URL"`
	StripeAPIKey        string `env:"STRIPE_API_KEY"`

	StripeCheckoutCallback string `env:"STRIPE_CHECKOUT_CALLBACK" envDefault:"/profile/portfolio"`

	StripeVideoPriceID         string   `env:"STRIPE_VIDEO_PRICE_ID" envDefault:""`
	StripeSubscriptionPriceIds []string `env:"STRIPE_SUBSCRIPTION_PRICE_IDS" envSeparator:","`
}

var (
	ProfileDir = "profiles"
	MediaDir   = "media"
	PublicDir  = "public"
)

func Init() {
	if err := env.Parse(&Cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	if Cfg.S3Region == "" && os.Getenv("AWS_DEFAULT_REGION") != "" {
		Cfg.S3Region = os.Getenv("AWS_DEFAULT_REGION")
	}

	if Cfg.PostgresqlURL == "" {
		Cfg.PostgresqlURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s", Cfg.DBUsername, Cfg.DBPassword, Cfg.DBHost, Cfg.DBPort, Cfg.DBDatabase, Cfg.DBSchema)
	}
	if Cfg.S3Endpoint != "" {
		Cfg.IsMinio = true
	}
}
