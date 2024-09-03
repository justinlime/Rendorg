package converter

import (
    "os"
    "fmt"
    "strings"
	fp "path/filepath"

	"github.com/justinlime/Rendorg/v2/utils"
	"github.com/justinlime/Rendorg/v2/config"

	"github.com/rs/zerolog/log"
	"github.com/niklasfasching/go-org/org"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/alecthomas/chroma/v2/formatters/html"
)

// Boilerplate HTML, with linked JS/CSS files from /js and /css respectively
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

func GenIndex() error {
    index, err := generatePrefix("Rendorg")     
    if err != nil {
        return fmt.Errorf("Failed to generate index - %v", err)
    }
    *index += indexPrefix
    for _, of := range OrgFiles {
        link := fmt.Sprintf(`<a class="index-link" href="%s">%s</a>`, of.WebPath, of.Title)
        *index += "\n" + link + "\n"
    }
    *index += indexSuffix
    htmlFile, err := os.Create("/tmp/rendorg/rendorg_index.html")
    if err != nil {
        return err
    }
    if _, err := htmlFile.Write([]byte(*index)); err != nil {
        return err
    }
    return nil
}

var indexPrefix string = `
<body id=index-body>
  <h1 class="index-title" id="index-title">Rendorg</h1>
    <div id="search-container">
      <input id="searchbar" 
             onkeyup="search()" 
             type="text" name="search" 
             placeholder="Search...">
    </div>
    <div id="link-container">
`
var indexSuffix string = `
    </div>
    <script>
      let origList = [];
      for (let orig of document.getElementsByClassName('index-link')) {
         origList.push(orig.style.display) 
      }
      function search() {
        let input = document.getElementById('searchbar').value
        input = input.toLowerCase();
        let x = document.getElementsByClassName('index-link');

        for (i = 0; i < x.length; i++) {
          if (!x[i].innerHTML.toLowerCase().includes(input)) {
            x[i].style.display = "none";
          }
          else {
            x[i].style.display = origList[i];
          }
        }
      }
    </script>
</body>
`
