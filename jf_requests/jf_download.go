package jf_requests

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

func CreatePBar(length int64, description string) *progressbar.ProgressBar {
	desc := ""
	return progressbar.NewOptions64(
		length,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(1*time.Second),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionUseANSICodes(true),
	)
}

func DownloadFromUrl(downloadLink string, name string, outfile string, max int, current int) error {
	req, _ := http.NewRequest("GET", downloadLink, nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return errors.New(fmt.Sprintf("Request Failed: %s", err))
	}

	defer resp.Body.Close()

	f, err := os.OpenFile(outfile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open file: %s", err))
	}

	defer f.Close()

	bar := CreatePBar(resp.ContentLength, fmt.Sprintf("downloading %d/%d", current, max))
	io.Copy(io.MultiWriter(f, bar), resp.Body)

	return nil
}

func GetDownloadLinkForId(baseUrl string, token string, id string) string {
	return fmt.Sprintf(baseUrl+"/Items/%s/Download?api_key=%s", id, token)
}
