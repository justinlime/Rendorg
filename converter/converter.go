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
	"github.com/niklasfasching/go-org/org"
)


var mu sync.Mutex

// TODO convert currently does not resolve internal heading links. (*heading1 or #heading1)
// TODO maybe also include link resolution to external files
func Convert(inputFile string) (OrgFile, error) {
    // Remove the previous entries
    mu.Lock()
    RmOrg(inputFile)
    mu.Unlock()
    // Create the file
	file, err := os.Open(inputFile)
	if err != nil {
        return OrgFile{}, fmt.Errorf("Couldnt open the requested file: %v", err)
	}
	defer file.Close()
    orgFile, err := NewOrg(inputFile)
    if err != nil {
        return OrgFile{}, fmt.Errorf("Failed to gather org file info: %v", err)
    }
    // Parser Config
    orgConfig := org.New().Silent()
    orgConfig.AutoLink = true
    orgConfig.DefaultSettings = map[string]string{
        // Default settings for each field, if none are supplied in the doc
        "TODO": "TODO | DONE",
        "EXCLUDE_TAGS": "noexport",
		"OPTIONS":      "toc:t <:t e:t f:t pri:t todo:t tags:t title:t ealb:nil",
    }
    writer := org.NewHTMLWriter()
    writer.HighlightCodeBlock = highlightCodeBlock
    // Generate the HTML
    prefix, err := generatePrefix(orgFile.Title)
    *prefix += "<body id=org-body>"

    body, err := orgConfig.Parse(file, inputFile).Write(writer)
    if err != nil {
        return OrgFile{}, fmt.Errorf("Failed to parse org body: %v", err) 
    }

    suffix := "</body>"

    htmlContents := *prefix + body + suffix

    // Resolve the ID links
    mu.Lock()
    htmlResolved := ResolveIDLinks(&htmlContents, orgFile)
    mu.Unlock()

    // Write the file 
    err = os.MkdirAll(fp.Dir(orgFile.HTMLPath), 0755)
    if err != nil {
        return OrgFile{}, fmt.Errorf("Failed to create the tmp path %v", err)
    }
    htmlFile, err := os.Create(orgFile.HTMLPath) 
    if err != nil {
        return OrgFile{}, fmt.Errorf("Failed to create the HTML file %v", err)
    }
    if _, err := htmlFile.Write([]byte(*htmlResolved)); err != nil {
        return OrgFile{}, fmt.Errorf("Failed to write to HTML file")
    }
    return orgFile, nil
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
    // init the org files for link resolution
    for _, org := range orgFiles {
        if fp.Ext(org) == ".org" {
            of, err := NewOrg(org)
            if err != nil {
                log.Error().Err(err).
                    Str("file", org).
                    Msg("failed to track org properties")
            }
            OrgFiles = append(OrgFiles, of)
        }
    }
    var wg sync.WaitGroup
    // Diminishing returns after 3 threads from the systems I've tested on
    // This is mostly done for startup time
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
                _, err := Convert(org)
                if err != nil {
                    log.Error().Err(err).
                        Str("file", org).
                        Msg("Failed to convert file")
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

func NewOrg(inputFile string) (OrgFile, error) {
    // Get the props
    title, err := GetProperty("title", inputFile)
    if err != nil || title == "" {
        fileName := fp.Base(inputFile)
        title = fileName[:len(fileName)-len(fp.Ext(fileName))]
    }
    id, err := GetProperty("id", inputFile)
    if err != nil {
        return OrgFile{}, err
    }
    ids, err := GetRoamIDs(inputFile)
    if err != nil {
        return OrgFile{}, err
    }
    webPath := strings.TrimSuffix(strings.ReplaceAll(inputFile, config.Cfg.InputDir, ""), ".org")
    htmlPath := strings.ReplaceAll(strings.ReplaceAll(inputFile, config.Cfg.InputDir, "/tmp/rendorg"), ".org", ".html")
    // ConSTRUCT
    orgFile := OrgFile {
        RealPath: inputFile, 
        HTMLPath: htmlPath,
        WebPath: webPath,
        ID: id,
        Title: title,
        LinkedToIDs: ids,
    }
    OrgFiles = append(OrgFiles, orgFile)
    return orgFile, nil
}
