package converter

import (
	"os"
	"fmt"
	"bufio"
	"strings"
    fp "path/filepath"

	"github.com/justinlime/Rendorg/v2/config"
	"github.com/justinlime/Rendorg/v2/utils"
)

// Scan an org file for tags using a key, such as :TITLE
// or :ID
func GetProperty(key string, filename string) (string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return "", err
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
            return strings.TrimLeft(combined, " "), nil
        } 
    }
    return "", fmt.Errorf("Could now locate property %s", key)
}

// Resolve org roam links in the file to actual HTML links
func ResolveLinks(contents *string) (*string, error) {
    orgFiles, err := utils.GetPathsRecursively(config.Cfg.InputDir)
    if err != nil {
        return nil, fmt.Errorf("Failed to get the filepaths from the output dir")
    }
    orgIDs := make(map[string]string)
    for _, org := range orgFiles {
        if fp.Ext(org) == ".org" {
            id, err := GetProperty("ID:", org)
            if err != nil {
                continue
            }
            orgIDs[org] = id
        }
    }
    resolved := *contents
    for org, id := range orgIDs {
        origLink := fmt.Sprintf(`href="id:%s"`, id)
        replLink := fmt.Sprintf(`href="%s"`, strings.ReplaceAll(org, config.Cfg.InputDir, ""))
        resolved = strings.ReplaceAll(resolved, origLink, replLink)
    }
    return &resolved, nil
}
