package main

import (
	"flag"

	"github.com/justinlime/Rendorg/v2/utils"
	"github.com/justinlime/Rendorg/v2/logger"
	"github.com/justinlime/Rendorg/v2/config"
	"github.com/justinlime/Rendorg/v2/webserver"
    conv "github.com/justinlime/Rendorg/v2/converter"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)
var (
    // flags
    inputDir string
    configDir string
    debug bool
)

func init() {
    flag.StringVar(&inputDir, "input", ".", "Directory containing your org files")
    flag.StringVar(&configDir, "config", "~/.config/rendorg", "Directory containing your rendorg.toml file")
    flag.BoolVar(&debug, "debug", false, "Enable debugging logs")
    flag.Parse()
    if debug {
        logger.InitLogger(zerolog.DebugLevel)
    } else {
        logger.InitLogger(zerolog.InfoLevel)
    }
    // Validate the input info
    if err := utils.ValidatePath(&inputDir); err != nil {
        log.Fatal().
            Str("input_dir", inputDir).
            Err(err).
            Msg("Could not validate input")
    } 
    if err := utils.ValidatePath(&configDir); err != nil {
        log.Fatal().
            Str("input_dir", inputDir).
            Err(err).
            Msg("Could not validate input")
    }
}
func main() {
    log.Info().
        Str("input_dir", inputDir).
        Str("config_dir", configDir).
        Msg("Using the following directories")
    config.InitConfig(configDir, inputDir)
    conv.ConvertAll()
    webserver.StartServer()
    select{}
}
