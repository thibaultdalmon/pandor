package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// BuildFile returns a pointer to a FILE if it has been created successfully
// or raise an error if it exists or an other error has occured during
// creation
func BuildFile(dir, name string) (*os.File, error) {
	var out *os.File
	if FileExists, err := Exists(dir + name); FileExists {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("File %v already exists", dir+name)
	}
	err := BuildDir(dir)
	if err != nil {
		return nil, err
	}
	out, err = os.Create(dir + name)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BuildDir creates a directory
func BuildDir(dir string) error {
	if dirExists, err := Exists(dir); dirExists {
		if err != nil {
			return err
		}
		return nil
	}
	err := os.MkdirAll(dir, 0777)
	return err
}

// DownloadAndSaveToDir is an helper function to save the result of a GET Request
func DownloadAndSaveToDir(url, file, dir string) error {

	out, err := BuildFile(dir, file)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
