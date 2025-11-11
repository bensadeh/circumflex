package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

const DownloadDir string = "circumflex/downloads"

// Guess the PDF filename by extracting the last part of the download url
func GuessDownloadFileName(url string) (string, error) {
	absoluteFilepath := path.Base(url)

	if absoluteFilepath == "." || absoluteFilepath == "/" {
		return "", fmt.Errorf("Could not guess filename, potential invalid URL!")
	}

	return absoluteFilepath, nil
}

func GetDownloadDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		// If getting the current working directory fails as well,
		// the user might have greater issues than simply downloading a file
		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	} else {
		dir = fmt.Sprintf("%s/%s", dir, DownloadDir)
	}

	return dir, nil
}

// Writes to the destination file as it downloads it, without
// loading the entire file into memory.
func DownloadFile(url string, absoluteFilepath string) error {
	out, err := os.Create(absoluteFilepath)
	if err != nil {
		return fmt.Errorf("Could not create file")
	}
	defer out.Close()

	// Timeout download attempt after 10 seconds
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Could not download file")
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("Could not write download file data to destination")
	}

	return nil
}
