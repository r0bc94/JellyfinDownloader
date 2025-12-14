package jf_requests

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/fatih/color"
)

type Movie struct {
	Name         string
	Id           string
	Container    string
	Path         string
	CanDownload  bool
	DownloadLink string
}

func GetMovieFromItem(auth *AuthResponse, baseurl string, item *Item) (*Movie, error) {
	requestUrl := fmt.Sprintf("%s/Users/%s/Items/%s", baseurl, auth.UserId, item.Id)

	res, err := MakeRequest(auth.Token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	// Check if media container arg is passed. If not, print a warning that this media
	// might be missing or corrupted.
	if res["Container"] == nil {
		return nil, errors.New(fmt.Sprintf("Could not get container format for requested movie; Might be missing or corrupted!"))
	}

	mov := Movie{
		Name:         res["Name"].(string),
		Id:           res["Id"].(string),
		Container:    res["Container"].(string),
		CanDownload:  res["CanDownload"].(bool),
		Path:         res["Path"].(string),
		DownloadLink: ""}

	mov.DownloadLink = GetDownloadLinkForId(baseurl, auth.Token, mov.Id)

	return &mov, nil
}

func (movie *Movie) PrintAndGetConfirmation() bool {
	if movie.CanDownload {
		fmt.Println("The following Movie will be downloaded:")
		color.Green("Name: %s", movie.Name)

		return GetConfirmation()
	} else {
		color.Yellow("Cannot download the Move \"%s\" due to insufficient permission!", movie.Name)
		return false
	}
}

func (movie *Movie) Download(keepFilename bool) {
	var outfilename string
	if keepFilename {
		basename := path.Base(movie.Path)
		outfilename = basename
	} else {
		suffix := strings.Split(movie.Container, ",")[0]
		outfilename = fmt.Sprintf("%s_%s.%s", movie.Name, movie.Name, suffix)
	}

	DownloadFromUrl(movie.DownloadLink, movie.Name, outfilename, 1, 0)
}
