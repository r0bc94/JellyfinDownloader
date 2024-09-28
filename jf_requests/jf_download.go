package jf_requests

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
		progressbar.OptionThrottle(65*time.Millisecond),
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

func DownloadEpisodes(seasons []Season) {
	for _, season := range seasons {
		for idx, episode := range season.Episodes {
			suffix := strings.Split(episode.Container, ",")[0]
			outfilename := fmt.Sprintf("%s_%s.%s", season.Name, episode.Name, suffix)
			DownloadFromUrl(episode.DownloadLink, episode.Name, outfilename, len(season.Episodes), idx)
		}
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
