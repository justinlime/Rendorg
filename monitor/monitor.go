package monitor

import (
	"io/fs"
	"os"
	fp "path/filepath"

	"github.com/justinlime/Rendorg/v2/config"
	conv "github.com/justinlime/Rendorg/v2/converter"

	// "github.com/justinlime/Rendorg/v2/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

func watchDir(watcher *fsnotify.Watcher, dir string) {
    if err := watcher.Add(dir); err != nil {
        log.Warn().Err(err).Str("dir", dir).Msg("Failed to add directory to the watcher")
    }
    log.Debug().Str("dir", dir).Msg("Added directrory to the watcher")
    fp.Walk(dir, func(path string, info fs.FileInfo, err error) error {
        if err != nil {
            log.Warn().Str("dir", dir).Err(err).Msg("Failed to recurse through directory")
            return nil
        }
        if info.IsDir() && path != dir {
            log.Debug().Str("path", path).Msg("Added directory to the watcher")
            if err := watcher.Add(path); err != nil {
                log.Warn().Err(err).Str("dir", dir).Msg("Failed to add directory to the watcher")
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

	// Add the directory to the watcher
    watchDir(watcher, config.Cfg.InputDir)

	log.Info().Str("dir", config.Cfg.InputDir).Msg("Watching directory")

	// Channel to receive events
	done := make(chan bool)
	go func() {
		for {
			select {
			// case event, ok := <-watcher.Events:
			case event, ok := <-watcher.Events:
                if !ok {
                    log.Error().Msg("Failed to watch directory")
                    return
                }
                switch {
                case event.Has(fsnotify.Remove) ||
                     event.Has(fsnotify.Rename):
                    org := conv.GetOrg(event.Name)
                    // Regnerate all the files that are linked to this one
                    if org != nil {
                        for _, of := range org.LinkedFrom {
                            var newTo []*conv.OrgFile
                            for _, o := range of.LinkedTo {
                                if o.RealPath != org.RealPath {
                                    newTo = append(newTo, o)
                                }
                            }
                            of.LinkedTo = newTo
                            var newOrgs []conv.OrgFile
                            for _, of := range conv.OrgFiles {
                                if of.RealPath != org.RealPath {
                                   newOrgs = append(newOrgs, of)
                                }
                            }
                            conv.OrgFiles = newOrgs
                            os.RemoveAll(org.HTMLPath)
                            if _, err := conv.Convert(of.RealPath); err != nil {
                                log.Error().Err(err).
                                    Str("file", of.RealPath).
                                    Msg("Failed to convert org file")
                            }
                            log.Info().Str("file", of.RealPath).Msg("Re-converted linked file")
                        }
                    }
                    log.Info().Str("file", event.Name).
                        Str("event", event.Op.String()).
                        Msg("Monitor")
                default:
                    continue
                }
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error().Err(err).Msg("Error")
			}
		}
	}()
	// Block the main goroutine
	<-done
}
