package utils 

import (
	"io"
	"os"
	fp "path/filepath"
)

// Copy a file, to a directory, using absolute paths
func copyFile(src string, dst string) error {
    err := os.MkdirAll(fp.Dir(dst), 0755) 
    if err != nil {
       return err 
    }
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    _, err = io.Copy(dstFile, srcFile)
    if err != nil {
        return err
    }
    return nil 
}
// Copy a file or directory, to a directory, using absolute paths
func Copy(src string, dst string) error {
    file, err := os.Stat(src)
    if err != nil {
        return err
    }
    if !file.IsDir() {
        err = copyFile(src, dst) 
        if err != nil {
            return err     
        }
        return nil
    }
    err = os.MkdirAll(dst, 0755)
    if err != nil {
        return err
    }
    files, err := os.ReadDir(src)
    if err != nil {
        return err
    }
    for _, file := range files {
        srcPath := fp.Join(src, file.Name())
        dstPath := fp.Join(dst, file.Name())
        if file.IsDir() {
            err = Copy(srcPath, dstPath)
            if err != nil {
                return err
            }
        } else {
            err = copyFile(srcPath, dstPath)
            if err != nil {
                return err
            } 
        }
    }
    return nil
}
