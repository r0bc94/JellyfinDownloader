package jf_requests

import (
	"errors"
	"fmt"
)

type Movie struct {
	Name         string
	Id           string
	Container    string
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
		DownloadLink: ""}

	mov.DownloadLink = GetDownloadLinkForId(baseurl, auth.Token, mov.Id)

	return &mov, nil
}
