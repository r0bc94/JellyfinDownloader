package jf_requests

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func DownloadFromUrl(episode *Episode, outfile string, max int, current int) error {
	req, _ := http.NewRequest("GET", episode.DownloadLink, nil)
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

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		fmt.Sprintf("Downloading %d/%d %s: ", current+1, max, episode.Name),
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)

	return nil
}

func Download(episodes []Episode) {
	for idx, episode := range episodes {
		suffix := strings.Split(episode.Container, ",")[0]
		outfilename := fmt.Sprintf("%s_%s.%s", episode.SeriesName, episode.Name, suffix)
		DownloadFromUrl(&episode, outfilename, len(episodes), idx)
	}
}
