package common

type Config struct {
	Port        uint
	HashKey     string
	BlockKey    string
	SessName    string
	SessExpires uint
}

var Appcfg *Config

const ServiceICache = "cache.AbstractCache"

func init() {
	Appcfg = &Config{
		HashKey:     "the-big-and-secret-fash-key-here",
		BlockKey:    "lot-secret-of-characters-big-too",
		Port:        3307,
		SessName:    "gosessionid",
		SessExpires: 3600 * 24,
	}
}
