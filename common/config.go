package common

import (
	"github.com/xiusin/logger"
)

type Config struct {
	LogLevel    logger.Level
	Port        uint
	HashKey     string
	BlockKey    string
	SessName    string
	SessExpires uint
}

var Appcfg *Config

const ServiceICache = "cache.AbstractCache"

const ServiceEmbedAssets = "*embed.FS"

func init() {
	Appcfg = &Config{
		LogLevel:    logger.DebugLevel,
		HashKey:     "the-big-and-secret-fash-key-here",
		BlockKey:    "lot-secret-of-characters-big-too",
		Port:        3307,
		SessName:    "gosessionid",
		SessExpires: 3600 * 24,
	}
}
