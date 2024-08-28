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

type OrgFile struct {
    RealPath string
    HTMLPath string
    WebPath  string
    ID       string 
    Title    string
    LinkedIDs []string
}

var OrgFiles []OrgFile

func Convert(inputFile string) (OrgFile, error) {
	file, err := os.Open(inputFile)
	if err != nil {
        return OrgFile{}, fmt.Errorf("Couldnt open the requested file: %v", err)
	}
	defer file.Close()
    orgFile, err := NewOrg(inputFile)
    if err != nil {
        return OrgFile{}, fmt.Errorf("Failed to gather org file info: %v", err)
    }
    // Convert the file
	d := org.New().Parse(file, inputFile)
	write := func(w org.Writer) (*string, error) {
        prefix, err := generatePrefix(orgFile.Title)
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
    htmlContents, err := write(writer)
    if err != nil {
        return OrgFile{}, err 
    }
    // Resolve the links
    htmlResolved := ResolveLinks(htmlContents, orgFile)
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


// Resolve org roam links in the file to actual HTML links
func ResolveLinks(contents *string, orgFile OrgFile) *string {
    resolved := *contents
    for _, org := range OrgFiles {
        // log.Info().Strs("ids", org.LinkedIDs).Str("comapred-id", org.ID).Msg("Test")
        if utils.Contains(orgFile.LinkedIDs, org.ID) {
            origLink := fmt.Sprintf(`href="id:%s"`, org.ID)
            replLink := fmt.Sprintf(`href="%s"`, org.HTMLPath)
            resolved = strings.ReplaceAll(resolved, origLink, replLink)
        }
    }
    return &resolved
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
    linkedIDs, err := GetRoamIDs(inputFile)
    if err != nil {
        return OrgFile{}, err
    }
    webPath := strings.TrimSuffix(strings.ReplaceAll(inputFile, config.Cfg.InputDir, ""), ".org")
    htmlPath := strings.ReplaceAll(strings.ReplaceAll(inputFile, config.Cfg.InputDir, "/tmp/rendorg"), ".org", ".html")
    // Construct the Struct
    orgFile := OrgFile {
        RealPath: inputFile, 
        HTMLPath: htmlPath,
        WebPath: webPath,
        ID: id,
        Title: title,
        LinkedIDs: linkedIDs,
    }
    OrgFiles = append(OrgFiles, orgFile)
    return orgFile, nil
}
