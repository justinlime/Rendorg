package config

import (
	"os"
	fp "path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/rs/zerolog/log"
)

type Config struct {
    CodeStyle      string `toml:"code-highlighting-style"`
    ListenPort     int `toml:"listen-port"`
    Username       string `toml:"username"`
    Password       string `toml:"password"`
    InputDir       string
} 

var Cfg Config

func InitConfig(configDir, inputDir string) {
    Cfg.InputDir = inputDir
    ensureDir := func (dir string) {
        if err := os.MkdirAll(dir, 0755); err != nil {
            log.Fatal().Err(err).Str("dir", dir).Msg("Failed to create dir")
        }
    }
    ensureDir(inputDir)
    ensureDir(configDir)
    configFile := fp.Join(configDir, "rendorg.toml")
    if _, err := os.Stat(configFile); err != nil {
        if err := os.WriteFile(configFile, []byte(defaultConfig), 0755); err != nil {
            log.Fatal().Err(err).Str("dir", configDir).Msg("Failed to create rendorg.toml")
        }
    }
    if _, err := toml.DecodeFile(configFile, &Cfg); err != nil {
        log.Fatal().Err(err).Str("file", configFile).Msg("Failed to parse rendorg.toml")
    }    
}
