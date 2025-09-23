package jf_requests

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Episode struct {
	Name        string
	Id          string
	Container   string
	CanDownload bool
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
	requestUrl := fmt.Sprintf("%s/Shows/%s/Episodes?fields=candownload", baseurl, item.Id)

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
	var currentSeason *Season
	var seasonId string
	var lastSeasonId string
	for index := 0; index < len(items); index += 1 {
		if index+1 < len(items) {
			seasonId = items[index+1].(map[string]any)["SeasonId"].(string)
		}

		// Create Initial Season
		if index == 0 {
			currentSeason = &Season{
				Id:   seasonId,
				Name: items[index].(map[string]any)["SeasonName"].(string),
			}

			lastSeasonId = seasonId
		}

		ep := Episode{
			Name:        items[index].(map[string]any)["Name"].(string),
			Id:          items[index].(map[string]any)["Id"].(string),
			Container:   items[index].(map[string]any)["Container"].(string),
			CanDownload: items[index].(map[string]any)["CanDownload"].(bool)}

		currentSeason.Episodes = append(currentSeason.Episodes, ep)

		if seasonId != lastSeasonId {
			seasons = append(seasons, *currentSeason)
			currentSeason = &Season{
				Id:   seasonId,
				Name: items[index+1].(map[string]any)["SeasonName"].(string),
			}

			lastSeasonId = seasonId
		}

		if index+1 == len(items) {
			seasons = append(seasons, *currentSeason)
			break
		}
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
	fmt.Println("Which Season do you want to download:")

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
		return []Season{series.Seasons[choice-1]}, nil
	}

}

func (series *Series) PrintAndGetConfirmation(seasonsToDownload []Season) bool {
	fmt.Println("The following Episodes will be downloaded:")
	color.Green(series.Name)
	undownloadbleItemsPresent := false

	for season_index, season := range seasonsToDownload {
		color.Cyan("  └ %d. %s", season_index+1, season.Name)
		for episode_index, episode := range season.Episodes {
			outstring := fmt.Sprintf("    └ %d. %s", episode_index+1, episode.Name)

			// Strike out episodes which can not be downloaded from the Jellyfin server due to the CanDownload attribute
			// set to false
			if !episode.CanDownload {
				outstring = fmt.Sprintf("\033[9m%s\033[0m", outstring)
				undownloadbleItemsPresent = true
			}
			color.Cyan(outstring)
		}
	}

	if undownloadbleItemsPresent {
		color.Yellow("Some items cannot be downloaded due to missing permissions!")
		color.Yellow("The affected Items are struck through.")
	}

	return GetConfirmation()
}

func (season *Season) Download(baseUrl string, token string) {
	for idx, episode := range season.Episodes {
		if episode.CanDownload {
			suffix := strings.Split(episode.Container, ",")[0]
			seasonid := strings.Split(season.Name, " ")
			outfilename := fmt.Sprintf("S%sE%d %s.%s", seasonid[len(seasonid)-1], int(idx)+1, episode.Name, suffix)
			downloadLink := GetDownloadLinkForId(baseUrl, token, episode.Id)
			DownloadFromUrl(downloadLink, episode.Name, outfilename, len(season.Episodes), idx)
		} else {
			color.Yellow("Skipping non downloadable item: %s", episode.Name)
		}
	}
}
