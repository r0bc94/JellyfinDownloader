package jf_requests

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Episode struct {
	Name         string
	Id           string
	Container    string
	DownloadLink string
}

type Season struct {
	Id       string
	Name     string
	Episodes []Episode
}

type Series struct {
	Name    string
	Id      string
	Seasons []Season
}

func GetSeriesFromItem(token string, baseurl string, item *Item) (*Series, error) {
	requestUrl := fmt.Sprintf("%s/Shows/%s/Episodes", baseurl, item.Id)

	res, err := MakeRequest(token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	var result Series = Series{
		Id:   item.Id,
		Name: item.Name,
	}

	items := res["Items"].([]any)

	var seasons []Season

	var currentSeason Season
	for _, item := range items {

		// Check if media container arg is passed. If not, print a warning that this media
		// might be missing or corrupted.
		if item.(map[string]any)["Container"] == nil {
			color.Yellow("Could not get container format for episode \"%s\"; Might be missing or corrupted!", item.(map[string]any)["Name"].(string))
			continue
		}

		seasonId := item.(map[string]any)["SeasonId"].(string)
		if currentSeason.Id == "" || currentSeason.Id != seasonId {
			season := Season{
				Id:   seasonId,
				Name: item.(map[string]any)["SeasonName"].(string),
			}

			if currentSeason.Id != "" {
				seasons = append(seasons, currentSeason)
			}

			currentSeason = season
		}

		ep := Episode{
			Name:         item.(map[string]any)["Name"].(string),
			Id:           item.(map[string]any)["Id"].(string),
			Container:    item.(map[string]any)["Container"].(string),
			DownloadLink: ""}

		ep.DownloadLink = GetDownloadLinkForId(baseurl, token, ep.Id)
		currentSeason.Episodes = append(currentSeason.Episodes, ep)
	}

	result.Seasons = seasons

	return &result, nil
}

func (series *Series) GetSeasonForId(seasonId string) (*Season, error) {
	for _, season := range series.Seasons {
		if season.Id == seasonId {
			return &season, nil
		}
	}

	return nil, errors.New(fmt.Sprint("No Season found for id %s", seasonId))
}

func (series *Series) PrintAndGetSelection() ([]Season, error) {
	fmt.Println("Which Seasons do you want to download:")

	color.Cyan("  0. All")
	for idx, season := range series.Seasons {
		color.Cyan("  %d. %s", idx+1, season.Name)
	}

	choice, err := GetUserChoice(len(series.Seasons))
	if err != nil {
		return nil, errors.New("Only provide a single number")
	}

	if choice == 0 {
		return series.Seasons, nil
	} else {
		return []Season{series.Seasons[choice]}, nil
	}

}

func (series *Series) PrintAndGetConfirmation(seasonsToDownload []Season) bool {
	fmt.Println("The following Episodes will be downloaded:")
	color.Green(series.Name)

	for season_index, season := range seasonsToDownload {
		color.Cyan("  └ %d. %s", season_index+1, season.Name)
		for episode_index, episode := range season.Episodes {
			color.Cyan("    └ %d. %s", episode_index+1, episode.Name)
		}
	}

	return GetConfirmation()
}

func (season *Season) Download() {
	for idx, episode := range season.Episodes {
		suffix := strings.Split(episode.Container, ",")[0]
		outfilename := fmt.Sprintf("%s_%s.%s", season.Name, episode.Name, suffix)
		DownloadFromUrl(episode.DownloadLink, episode.Name, outfilename, len(season.Episodes), idx)
	}
}
