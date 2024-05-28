package config

import (
	"flag"
	"os"
)

type Options struct {
	flagRunAddr, flagLogLevel, flagDataBaseDSN,
	flagJWTSigningKey, flagAccrualSystemAddress,
	flagConcurrency, flagTaskExecutionInterval string
}

func NewOptions() *Options {
	return new(Options)
}

// parseFlags handles command line arguments
// and stores their values in the corresponding variables.
func (o *Options) ParseFlags() {
	regStringVar(&o.flagRunAddr, "a", ":8080", "address and port to run server")
	regStringVar(&o.flagConcurrency, "c", "5", "Concurrency")
	regStringVar(&o.flagDataBaseDSN, "d", "", "")
	regStringVar(&o.flagTaskExecutionInterval, "i", "3000", "Task execution interval in milliseconds")
	regStringVar(&o.flagJWTSigningKey, "j", "test_key", "jwt signing key")
	regStringVar(&o.flagLogLevel, "l", "info", "log level")
	regStringVar(&o.flagAccrualSystemAddress, "r", ":8082", "acrual system address")

	// parse the arguments passed to the server into registered variables
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		o.flagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		o.flagLogLevel = envLogLevel
	}

	if envDataBaseDSN := os.Getenv("DATABASE_URI"); envDataBaseDSN != "" {
		o.flagDataBaseDSN = envDataBaseDSN
	}

	if envJWTSigningKey := os.Getenv("JWT_SIGNING_KEY"); envJWTSigningKey != "" {
		o.flagJWTSigningKey = envJWTSigningKey
	}

	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		o.flagAccrualSystemAddress = envAccrualSystemAddress
	}

	if envConcurrency := os.Getenv("CONCURRENCY"); envConcurrency != "" {
		o.flagConcurrency = envConcurrency
	}

	if envTaskExecutionInterval := os.Getenv("TASK_EXECUTION_INTERVAL"); envTaskExecutionInterval != "" {
		o.flagTaskExecutionInterval = envTaskExecutionInterval
	}
}

func (o *Options) RunAddr() string {
	return getStringFlag("a")
}

func (o *Options) LogLevel() string {
	return getStringFlag("l")
}

func (o *Options) DataBaseDSN() string {
	return getStringFlag("d")
}

func (o *Options) JWTSigningKey() string {
	return getStringFlag("j")
}

func (o *Options) AccrualSystemAddress() string {
	return getStringFlag("r")
}

func (o *Options) Concurrency() string {
	return getStringFlag("c")
}

func (o *Options) TaskExecutionInterval() string {
	return getStringFlag("i")
}

func regStringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

func getStringFlag(name string) string {
	return flag.Lookup(name).Value.(flag.Getter).Get().(string)
}

// GetAsString reads an environment or returns a default value.
func GetAsString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
