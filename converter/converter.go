package converter

import (
	"os"
    "fmt"
    "sync"
    "time"
	"strings"
	fp "path/filepath"

	"github.com/justinlime/Rendorg/v2/utils"
	"github.com/justinlime/Rendorg/v2/config"

	"github.com/rs/zerolog/log"
	"github.com/alecthomas/chroma/v2"
	"github.com/niklasfasching/go-org/org"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/alecthomas/chroma/v2/formatters/html"
)

func Convert(inputFile string) (*string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
        return nil, fmt.Errorf("Couldnt open the requested file: %v", err)
	}
    title, err := GetProperty("title:", inputFile)
    if err != nil || title.Prop == "" {
        title.Prop = file.Name()[:len(file.Name())-len(fp.Ext(file.Name()))]
    }
	defer file.Close()
	d := org.New().Parse(file, inputFile)
	write := func(w org.Writer)  (*string, error) {
        prefix, err := generatePrefix(title.Prop)
        if err != nil {
            return nil, fmt.Errorf("Failed to generate prefix for org file: %v", err)
        }
		body, err := d.Write(w)
		if err != nil {
            return nil, fmt.Errorf("Failed to convert the requested file: %v", err)
		}
        suffix := `
        </body>
        `
        contents := *prefix + body + suffix
        return &contents, nil
	}
    writer := org.NewHTMLWriter()
    writer.HighlightCodeBlock = highlightCodeBlock
    html, err := write(writer)
    if err != nil {
        return nil, err 
    }
    return html, nil
}

func ConvertAll() {
    begin := time.Now()
    err := os.RemoveAll("/tmp/rendorg")
    if err != nil {
        log.Error().Err(err).Msg("Failed to clean up the temporary directory")
    }
    if err := os.MkdirAll("/tmp/rendorg", 0755); err != nil {
        log.Error().Err(err).Str("dir", "/tmp/rendorg").
            Msg("Failed to create temp directory")

    }
    orgFiles, err := utils.GetPathsRecursively(config.Cfg.InputDir)
    if err != nil {
        log.Error().Err(err).
            Str("dir", config.Cfg.InputDir).
            Msg("Failed to recurse through the input directory")
    }
    allProps , err := GetAllProps(orgFiles)
    if err != nil {
        log.Error().Err(err).
            Str("dir", config.Cfg.InputDir).
            Msg("Failed to index the org properties in the directory")
    }

    var wg sync.WaitGroup
    ch := make(chan struct{}, 3)
    var count int
    for _, org := range orgFiles {
        if fp.Ext(org) == ".org" {
            wg.Add(1)
            count += 1
            ch <- struct{}{}
            go func() {
                defer wg.Done()
                defer func() { <-ch }()
                content, err := Convert(org)
                if err != nil {
                    log.Error().Err(err).
                        Str("file", org).
                        Msg("Failed to convert file")
                }
                resolved := *content
                for _, match := range allProps {
                    origLink := fmt.Sprintf(`href="id:%s"`, match.Prop)
                    replLink := fmt.Sprintf(`href="%s"`, strings.ReplaceAll(match.File, config.Cfg.InputDir, ""))
                    resolved = strings.ReplaceAll(resolved, origLink, replLink)
                }
                outPath := strings.ReplaceAll(strings.ReplaceAll(org, config.Cfg.InputDir, "/tmp/rendorg"),
                    ".org", ".html")
                if err := os.MkdirAll(fp.Dir(outPath), 0755); err != nil {
                    log.Error().Err(err).Str("dir", fp.Dir(outPath)).
                        Msg("Failed to create the destination path for the HTML file")
                }
                htmlFile, err := os.Create(outPath)
                if err != nil {
                    log.Error().Err(err).Str("file", outPath).
                        Msg("Failed to create converted HTML file")
                }
                if _, err := htmlFile.Write([]byte(resolved)); err != nil {
                    log.Error().Err(err).Str("file", outPath).
                        Msg("Failed to write converted HTML file")
                }
            }()
        }
    }
    wg.Wait()
    if err := GenIndex(); err != nil {
        log.Error().Err(err).Msg("Failed to generate the index page") 
    }
    duration := fmt.Sprintf("%fs", time.Since(begin).Seconds())
    log.Info().Int("org_files_converted", count).Str("time_elapsed", duration).Msg("Conversion Complete")
}

// TODO add searching feature
func GenIndex() error {
    index, err := generatePrefix("Rendorg")     
    if err != nil {
        return fmt.Errorf("Failed to generate index - %v", err)
    }
    files, err := utils.GetPathsRecursively(config.Cfg.InputDir)
    if err != nil {
        return fmt.Errorf("Failed to read root dir - %v", err)
    }
    *index += `<h1 class="index-title" id="index-title">Rendorg</h1>`
    var links []string
    for _, file := range files {
        if fp.Ext(file) == ".org" {
            title, err := GetProperty("title:", file)
            if err != nil {
                title.Prop = strings.TrimSuffix(fp.Base(file), ".org")
            }
            linkPath := strings.ReplaceAll(file, config.Cfg.InputDir, "")
            links = append(links, fmt.Sprintf(`<a class="index-link" href="%s">%s</a>`, linkPath, title.Prop))
        }
    }
    *index += "\n" + strings.Join(links, "\n")
    *index += "</body>"
    htmlFile, err := os.Create("/tmp/rendorg/rendorg_index.html")
    if err != nil {
        return err
    }
    if _, err := htmlFile.Write([]byte(*index)); err != nil {
        return err
    }
    return nil
}

// Returns HTML boilerplate
func generatePrefix(title string) (*string, error) {
    prefix := fmt.Sprintf(`
    <!DOCTYPE html>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    `, title)

    var cssFiles []string
    var jsFiles []string

    styleDir := fp.Join(config.Cfg.InputDir, "style")
    _, err := os.Stat(styleDir)
    if err != nil {
        log.Warn().Err(err).Str("dir", styleDir).
            Msg("Couldn't stat style dir, nothing will be applied.")    
    } else {
        cssFiles , err = utils.GetPathsRecursively(fp.Join(config.Cfg.InputDir, "style")) 
        if err != nil {
            log.Warn().Err(err).Str("dir", styleDir).
                Msg("Coudln't read style dir, nothing will be applied")
        }
    }

    jsDir := fp.Join(config.Cfg.InputDir, "js")
    _, err = os.Stat(jsDir)
    if err != nil {
        log.Warn().Err(err).Str("dir", jsDir).
            Msg("Couldn't stat js dir, nothing will be applied.")
    } else {
        jsFiles , err = utils.GetPathsRecursively(fp.Join(config.Cfg.InputDir, "js")) 
        if err != nil {
            log.Warn().Err(err).Str("dir", jsDir).
                Msg("Coudln't read js dir, nothing will be applied")
        }
    }

    for _, css := range cssFiles {
        if fp.Ext(css) == ".css" {
            prefix += fmt.Sprintf(`<link rel="stylesheet" href="%s">` + "\n", strings.ReplaceAll(css, config.Cfg.InputDir, ""))
        }
    }
    for _, js := range jsFiles {
        if fp.Ext(js) == ".js" {
            prefix += fmt.Sprintf(`<script src="%s" defer></script>` + "\n", strings.ReplaceAll(js, config.Cfg.InputDir, ""))
        }
    }
    prefix += "<body>\n"
    return &prefix, nil
}

// Copied directly from https://github.com/niklasfasching/go-org/main.go
func highlightCodeBlock(source, lang string, inline bool, params map[string]string) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	options := []html.Option{}
	if params[":hl_lines"] != "" {
		ranges := org.ParseRanges(params[":hl_lines"])
		if ranges != nil {
			options = append(options, html.HighlightLines(ranges))
		}
	}
	_ = html.New(options...).Format(&w, styles.Get(config.Cfg.CodeStyle), it)
	if inline {
		return `<div class="highlight-inline">` + "\n" + w.String() + "\n" + `</div>`
	}
	return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}
