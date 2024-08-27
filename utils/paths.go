package utils

import (
	"os"
	"fmt"
	"strings"
	fp "path/filepath"
)

// Get the path of every single file in a directory (not including the directories themselves)
func GetPathsRecursively(dir string) ([]string, error){
    var paths []string
    var readDir func(string) error
    readDir = func (newDir string)  error {
        files, err := os.ReadDir(newDir)        
        if err != nil {
            return err
        }
        for _, file := range files {
            if file.IsDir() {
                if err := readDir(fp.Join(newDir, file.Name())); err != nil {
                    return err
                }
            } else {
                paths = append(paths, fp.Join(newDir, file.Name()))
            }
        }
        return nil
    }
    if err := readDir(dir); err != nil {
        return []string{}, err
    }
    return paths, nil
}

// Convert ENV vars in a path, and expand ~ and . shorthands if present
func ValidatePath(p *string) error {
    var parsedPath string
    if strings.HasPrefix(*p, "/") {
       parsedPath = "/" 
    } else if strings.HasPrefix(*p, ".") {
        cwd, err := os.Getwd()
        if err != nil {
            return err
        } else {
            parsedPath = cwd
        }
    }
    for _, seg := range strings.Split(*p, "/") {
        var parsedSeg string
        // Parse ENV vars if they're in the path
        if strings.HasPrefix(seg, "$") {
            env := os.Getenv(seg[1:])
            if env != "" {
                parsedSeg = env
            } else {
                return fmt.Errorf("Failed parsing ENV variable in provided path: %v", *p)
            }
        // Parse the ~ shorthand if its in the path
        } else if strings.HasPrefix(seg, "~") {
            homeDir, err := os.UserHomeDir()
            if err != nil {
                return fmt.Errorf("Failed to locate user home directory when parsing path the path %v: %v", *p ,err)
            }
            parsedPath = homeDir
        } else {
            parsedSeg = seg
        }
        parsedPath = fp.Join(parsedPath, parsedSeg)
    }
    if !strings.HasPrefix(parsedPath, "/") {
        return fmt.Errorf("Path must be absolute: %s", parsedPath)
    }
    *p = parsedPath
    return nil
}
