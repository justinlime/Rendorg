package converter

import (
	"bufio"
	"fmt"
	"os"
	"strings"
    "regexp"
	// "github.com/justinlime/Rendorg/v2/utils"
)

// Scan an org file for tags using a key, such as :TITLE
// or :ID
func GetProperty(key string, filePath string) (string, error) {
    key1 := strings.ToLower(fmt.Sprintf(":%s:", key))
    key2 := strings.ToLower(fmt.Sprintf("#+%s:", key))

    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(strings.ToLower(line), key1) ||
           strings.Contains(strings.ToLower(line), key2) {
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
    return "", nil
}

// Returns the ID's of all RoamLinks in a file
func GetRoamIDs(filePath string) ([]string, error) {
    var matchedIDs []string
    file, err := os.Open(filePath)
    if err != nil {
        return []string{}, err
    }
    scanner := bufio.NewScanner(file)
    for scanner.Scan(){
        line := scanner.Text()
        if strings.Contains(strings.ToLower(line), "[[id:") {
            re, err := regexp.Compile(`\[\[id:([^\]]+)`)
            if err != nil {
                return []string{}, err
            }
            ids := re.FindAllStringSubmatch(line, -1)
            for _, id := range ids {
                if len(id) > 1 {
                    matchedIDs = append(matchedIDs, id[1])
                }
            }
        }
    }
    return matchedIDs, nil
}


// Resolve org roam links in the file to actual HTML links
func ResolveLinks(contents *string, orgFile OrgFile) *string {
    resolved := *contents
    for _, of := range orgFile.LinkedTo {
        for _, linked := range OrgFiles {
            if of.ID == linked.ID {
                origLink := fmt.Sprintf(`href="id:%s"`, linked.ID)
                replLink := fmt.Sprintf(`href="%s"`, linked.WebPath)
                resolved = strings.ReplaceAll(resolved, origLink, replLink)
            }
        }
    }
    return &resolved
}
