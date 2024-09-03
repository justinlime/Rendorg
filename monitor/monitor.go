package monitor

import (
	"os"
	"io/fs"
	fp "path/filepath"

	"github.com/justinlime/Rendorg/v2/config"
	conv "github.com/justinlime/Rendorg/v2/converter"

	// "github.com/justinlime/Rendorg/v2/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

func watchDir(watcher *fsnotify.Watcher, dir string) {
    // Fast additions and removals may trigger a nil pointer deref
    defer func() {
        if r := recover(); r != nil {
            log.Debug().Msg("Recovered from panic, file likely moved too fast.")
        }
    }()
    d, err := os.Stat(dir)
    if err != nil {
        log.Warn().Err(err).
            Msg("Failed to stat file")
    }
    if !d.IsDir() {
        return
    }
    if err := watcher.Add(dir); err != nil {
        log.Warn().Err(err).
            Str("dir", dir).
            Msg("Failed to add directory to the watcher")
    }
    log.Debug().Str("dir", dir).Msg("Added directrory to the watcher")
    fp.Walk(dir, func(path string, info fs.FileInfo, err error) error {
        if err != nil {
            log.Warn().Err(err).
                Str("dir", dir).
                Msg("Failed to recurse through directory")
            return nil
        }
        if info.IsDir() && path != dir {
            log.Debug().
                Str("path", path).
                Msg("Added directory to the watcher")
            if err := watcher.Add(path); err != nil {
                log.Warn().Err(err).
                    Str("dir", dir).
                    Msg("Failed to add directory to the watcher")
                return nil
            }
        }
        return nil
    })
}
// FIXME when editing large files, some editors may send a write message
// twice, add some type of timeout period to stop this.
func Monitor() {
	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating watcher")
	}
	defer watcher.Close()

    watchDir(watcher, config.Cfg.InputDir)

	log.Info().Str("dir", config.Cfg.InputDir).Msg("Watching directory")

	done := make(chan bool)
	go func() {
        convLinkedFrom := func (event fsnotify.Event, orgFile conv.OrgFile) {
            for _, of := range orgFile.LinkedFrom() {
                if _, err := conv.Convert(of.RealPath); err != nil {
                    log.Error().Err(err).
                        Str("event", event.Op.String()).
                        Str("file", of.RealPath).
                        Msg("Failed to re-convert linked org file")
                }
                log.Info().
                    Str("event", event.Op.String()).
                    Str("file", of.RealPath).
                    Msg("Re-converted linked org file")
            }
        }
		for {
			select {
			case event, ok := <-watcher.Events:
                if !ok {
                    log.Error().Msg("Failed to watch directory")
                    return
                }
                switch {
                // Rename will send a create signal directly after
                case event.Has(fsnotify.Remove) ||
                     event.Has(fsnotify.Rename):
                    o := conv.GetOrg(event.Name)
                    if o == nil {
                       continue 
                    }
                    org := *o
                    conv.RmOrg(event.Name)
                    convLinkedFrom(event, org)
                    log.Info().
                        Str("file", event.Name).
                        Str("event", event.Op.String()).
                        Msg("Monitor")
                case event.Has(fsnotify.Create):
                    watchDir(watcher, event.Name)
                    if fp.Ext(event.Name) == ".org" {
                        of, err := conv.Convert(event.Name)
                        if err != nil {
                            log.Error().Err(err).
                                Str("event", event.Op.String()).
                                Str("file", event.Name).
                                Msg("Failed to convert org file")
                        }
                        log.Info().
                            Str("event", event.Op.String()).
                            Str("file", event.Name).
                            Msg("Converted org file")
                        convLinkedFrom(event, of)
                    }
                case event.Has(fsnotify.Write):
                    watchDir(watcher, event.Name)
                    if fp.Ext(event.Name) == ".org" {
                        _, err := conv.Convert(event.Name)
                        if err != nil {
                            log.Error().Err(err).
                                Str("event", event.Op.String()).
                                Str("file", event.Name).
                                Msg("Failed to convert org file")
                        }
                        log.Info().
                            Str("event", event.Op.String()).
                            Str("file", event.Name).
                            Msg("Converted org file")
                    }
                default:
                    continue
                }
                if !event.Has(fsnotify.Chmod) {
                    if err := conv.GenIndex(); err != nil {
                        log.Error().Err(err).
                            Str("event", event.Name).
                            Msg("Failed to regenerate index after event.")
                    }
                }
				if !ok {
                    log.Error().Err(err).Msg("Error")
					return
				}
			}
		}
	}()
    <- done
}
