package webserver

import (
	"fmt"
	"net/http"
	fp "path/filepath"

	"github.com/justinlime/Rendorg/v2/config"
	conv "github.com/justinlime/Rendorg/v2/converter"

	"github.com/rs/zerolog/log"
)


// TODO maybe add optional path remapping
func Serve(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        http.ServeFile(w, r, "/tmp/rendorg/rendorg_index.html")
        return
    }
    for _, of := range conv.OrgFiles {
        if r.URL.Path == of.WebPath {
            http.ServeFile(w, r, of.HTMLPath)
            return
        }
    }
    http.ServeFile(w, r, fp.Join(config.Cfg.InputDir, r.URL.Path))
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
