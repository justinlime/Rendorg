package converter

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"sync"

	"github.com/justinlime/Rendorg/v2/config"
	// "github.com/justinlime/Rendorg/v2/utils"
)

type PropMatch struct {
    Prop string
    File string
}

// Scan an org file for tags using a key, such as :TITLE
// or :ID
func GetProperty(key string, filename string) (PropMatch, error) {
    file, err := os.Open(filename)
    if err != nil {
        return PropMatch{}, err
    }

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(strings.ToLower(line), strings.ToLower(key)) {
            var combined string
            parts := strings.Split(line, " ")
            for i, part := range parts {
                if i == 0 {
                    continue
                }
                combined += " " + part
            }
            return PropMatch {
                Prop: strings.TrimLeft(combined, " "),
                File: filename,
            }, nil
        } 
    }
    return PropMatch{}, fmt.Errorf("Could now locate property %s", key)
}

func GetAllProps(orgFiles []string) ([]PropMatch, error){
    var wg sync.WaitGroup
    ch := make(chan PropMatch, 10)
    for _, org := range orgFiles {
        if fp.Ext(org) == ".org" {
            wg.Add(1)
            go func() {
                defer wg.Done()
                match, err := GetProperty("ID:", org)
                if err != nil || match.Prop == "" {
                    return
                }
                ch <- match
            }()
        }
    }
    go func(){
        wg.Wait()
        close(ch)
    }()
    var matches []PropMatch
    for match := range ch {
        matches = append(matches, match)
    }
    return matches, nil
}

// Resolve org roam links in the file to actual HTML links
func ResolveLinks(inputFile string, contents *string, candidates []PropMatch) error {
    resolved := *contents
    for _, match := range candidates {
        origLink := fmt.Sprintf(`href="id:%s"`, match.Prop)
        replLink := fmt.Sprintf(`href="%s"`, strings.ReplaceAll(match.File, config.Cfg.InputDir, ""))
        resolved = strings.ReplaceAll(resolved, origLink, replLink)
    }
    if err := os.MkdirAll("/tmp/rendorg", 0755); err != nil {
        return fmt.Errorf("Failed to create tmp directory")
    }
    outPath := strings.ReplaceAll(inputFile, config.Cfg.InputDir, "/tmp/rendorg")
    htmlFile, err := os.Create(strings.ReplaceAll(outPath, ".org", ".html"))
    if err != nil {
        return err
    }
    if _, err := htmlFile.Write([]byte(resolved)); err != nil {
        return err
    }
    return nil
}
