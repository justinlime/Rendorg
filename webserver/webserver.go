package webserver

import (
	"os"
	"fmt"
	"strings"
	"net/http"
	fp "path/filepath"
	templ "html/template"

	"github.com/justinlime/Rendorg/v2/config"
	"github.com/justinlime/Rendorg/v2/utils"
	"github.com/justinlime/Rendorg/v2/templates"
	conv "github.com/justinlime/Rendorg/v2/converter"

	"github.com/rs/zerolog/log"
)

func orgHandler(orgFile conv.OrgFile, w http.ResponseWriter, r *http.Request) {
    var cssFiles []string
    var jsFiles []string

    styleDir := fp.Join(config.Cfg.ConfigDir, "style")
    _, err := os.Stat(styleDir)
    if err != nil {
        log.Warn().Err(err).Str("dir", styleDir).
            Msg("Couldn't stat style dir, nothing will be applied.")    
    } else {
        cssFiles , err = utils.GetPathsRecursively(styleDir) 
        if err != nil {
            log.Warn().Err(err).Str("dir", styleDir).
                Msg("Coudln't read style dir, nothing will be applied")
        }
    }

    jsDir := fp.Join(config.Cfg.ConfigDir, "js")
    _, err = os.Stat(jsDir)
    if err != nil {
        log.Warn().Err(err).Str("dir", jsDir).
            Msg("Couldn't stat js dir, nothing will be applied.")
    } else {
        jsFiles , err = utils.GetPathsRecursively(jsDir) 
        if err != nil {
            log.Warn().Err(err).Str("dir", jsDir).
                Msg("Coudln't read js dir, nothing will be applied")
        }
    }
    var HTMLCSSFiles []string
    var HTMLJSFiles  []string
    for _, css := range cssFiles {
        HTMLCSSFiles = append(HTMLCSSFiles, strings.ReplaceAll(css, config.Cfg.InputDir, ""))
    }
    for _, js := range jsFiles {
        HTMLJSFiles = append(HTMLJSFiles, strings.ReplaceAll(js, config.Cfg.InputDir, ""))
    }
    file, err := os.ReadFile(orgFile.HTMLPath)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
    templs, err := templ.ParseFS(
        templates.EHTML,
        "base.html",
        "nav.html",
    )
    if err != nil {
        log.Error().Err(err).Msg("Failed to parse the templates")
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
    err = templs.Execute(w, struct{
        Body templ.HTML
        Title string
        JS []string
        CSS []string
    }{
        Body: templ.HTML(file),
        Title: orgFile.Title,
        JS: HTMLJSFiles,
        CSS: HTMLCSSFiles,
    })
    if err != nil {
        log.Error().Err(err).Msg("Failed to execute the templates")
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}


// TODO maybe add optional path remapping
func Serve(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        http.ServeFile(w, r, "/tmp/rendorg/rendorg_index.html")
        return
    }
    noHTML := strings.TrimSuffix(r.URL.Path, ".html")
    for _, of := range conv.OrgFiles {
        if noHTML == of.WebPath {
            orgHandler(of, w, r)
            return
        }
    }
    http.ServeFile(w, r, fp.Join(config.Cfg.InputDir, r.URL.Path))
}

// TODO make auth optional, actually store in a secure way
func Auth(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// next.ServeHTTP(w, r)
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
