package webserver

import (
	"fmt"
	"net/http"
	fp "path/filepath"
	"strings"

	// conv "github.com/justinlime/Rendorg/v2/converter"
	"github.com/justinlime/Rendorg/v2/config"
	"github.com/justinlime/Rendorg/v2/utils"

	// conv "github.com/justinlime/Rendorg/v2/converter"

	"github.com/rs/zerolog/log"
)


// TODO maybe add optional path remapping
func Serve(w http.ResponseWriter, r *http.Request) {
    files, err := utils.GetPathsRecursively(config.Cfg.InputDir)
    if err != nil {
        fmt.Fprintf(w, "Failure! couldn't read paths!")
    }
    rootEntry := func(path string) string {
        for _, file := range files {
            mappedRoot := strings.ReplaceAll(file, config.Cfg.InputDir, "")
            if path == mappedRoot {
                return file 
            }
        } 
        return ""
    }
    if r.URL.Path == "/" {
        http.ServeFile(w, r, "/tmp/rendorg/rendorg_index.html")
    } else if match := rootEntry(r.URL.Path); match != "" {
        if fp.Ext(r.URL.Path) == ".org" {
            name := strings.ReplaceAll(r.URL.Path, ".org", ".html")
            outPath := fp.Join("/tmp/rendorg", name)
            http.ServeFile(w, r, outPath)
        } else {
            http.ServeFile(w, r, match)
        }
    }
}

// Todo make auth optional, actually store in a secure way
func Auth(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, pass, ok := r.BasicAuth()
        if ok {
            if user == config.Cfg.Username && pass == config.Cfg.Password {
				next.ServeHTTP(w, r)
				return
            }
        }
        w.Header().Set("WWW-Authenticate", `Basic realm="Protected Area"`)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    })
}

func StartServer() {
    mux := http.NewServeMux()
    authed := Auth(http.HandlerFunc(Serve))
    mux.Handle("/", authed)
    port := fmt.Sprintf(":%d", config.Cfg.ListenPort)
    log.Info().Str("port", port).Msg("Listening and serving")
    http.ListenAndServe(port, mux)
}
