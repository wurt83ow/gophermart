package config

import (
	"flag"
	"os"
)

type Options struct {
	flagRunAddr     string
	flagLogLevel    string
	flagDataBaseDSN string
}

func NewOptions() *Options {
	return &Options{}
}

// parseFlags handles command line arguments
// and stores their values in the corresponding variables
func (o *Options) ParseFlags() {
	regStringVar(&o.flagRunAddr, "a", ":8080", "address and port to run server")
	regStringVar(&o.flagLogLevel, "l", "info", "log level")
	regStringVar(&o.flagDataBaseDSN, "d", "", "")

	// parse the arguments passed to the server into registered variables
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		o.flagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		o.flagLogLevel = envLogLevel
	}

	if envDataBaseDSN := os.Getenv("DATABASE_DSN"); envDataBaseDSN != "" {
		o.flagDataBaseDSN = envDataBaseDSN
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

func regStringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

func getStringFlag(name string) string {
	return flag.Lookup(name).Value.(flag.Getter).Get().(string)
}
