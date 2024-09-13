package main

import (
	"flag"

	"github.com/justinlime/Rendorg/v2/config"
	"github.com/justinlime/Rendorg/v2/logger"
	"github.com/justinlime/Rendorg/v2/utils"

	"github.com/justinlime/Rendorg/v2/monitor"
	"github.com/justinlime/Rendorg/v2/webserver"
	conv "github.com/justinlime/Rendorg/v2/converter"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)
var (
    // flags
    inputDir string
    debug bool
)

func init() {
    flag.StringVar(&inputDir, "input", ".", "Directory containing your org files")
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
}
func main() {
    config.InitConfig(inputDir)
    log.Info().
        Str("config_dir", config.Cfg.ConfigDir).
        Str("input_dir", config.Cfg.InputDir).
        Msg("Using the following directories")
    conv.ConvertAll()
    // for _, of := range conv.OrgFiles {
    //     for _, lt := range of.LinkedTo() {
    //         log.Info().Str("file", of.Title).Str("linked_to", lt.Title).Msg("Link")
    //     }
    //     for _, lf := range of.LinkedFrom() {
    //         log.Info().Str("file", of.Title).Str("linked_from", lf.Title).Msg("Link")
    //     }
    // }
    go monitor.Monitor()
    webserver.StartServer()
}
