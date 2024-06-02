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

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		fmt.Sprintf("Downloading %d/%d %s: ", current+1, max, name),
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)

	return nil
}

func DownloadEpisodes(episodes []Episode) {
	for idx, episode := range episodes {
		suffix := strings.Split(episode.Container, ",")[0]
		outfilename := fmt.Sprintf("%s_%s.%s", episode.SeriesName, episode.Name, suffix)
		DownloadFromUrl(episode.DownloadLink, episode.Name, outfilename, len(episodes), idx)
	}
}

func DownloadMovie(movie *Movie) {
	suffix := strings.Split(movie.Container, ",")[0]
	outfilename := fmt.Sprintf("%s_%s.%s", movie.Name, movie.Name, suffix)
	DownloadFromUrl(movie.DownloadLink, movie.Name, outfilename, 1, 0)
}

func GetDownloadLinkForId(baseUrl string, token string, id string) string {
	return fmt.Sprintf(baseUrl+"/Items/%s/Download?api_key=%s", id, token)
}
