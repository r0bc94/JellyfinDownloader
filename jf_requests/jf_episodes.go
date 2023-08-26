package jf_requests

import (
	"fmt"

	"github.com/fatih/color"
)

type Episode struct {
	SeriesName   string
	Name         string
	Id           string
	SeasonId     string
	SeasonName   string
	Container    string
	DownloadLink string
}

type Season struct {
	Id       string
	Name     string
	Episodes []Episode
}

func GetEpisodesFromId(token string, baseurl string, seriesId string) ([]Episode, error) {
	requestUrl := fmt.Sprintf("%s/Shows/%s/Episodes", baseurl, seriesId)

	res, err := MakeRequest(token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	items := res["Items"].([]any)
	var result []Episode

	for _, item := range items {

		// Check if media container arg is passed. If not, print a warning that this media
		// might be missing or corrupted.
		if item.(map[string]any)["Container"] == nil {
			color.Yellow("Could not get container format for episode \"%s\"; Might be missing or corrupted!", item.(map[string]any)["Name"].(string))
			continue
		}

		ep := Episode{
			SeriesName:   item.(map[string]any)["SeriesName"].(string),
			Name:         item.(map[string]any)["Name"].(string),
			Id:           item.(map[string]any)["Id"].(string),
			SeasonId:     item.(map[string]any)["SeasonId"].(string),
			SeasonName:   item.(map[string]any)["SeasonName"].(string),
			Container:    item.(map[string]any)["Container"].(string),
			DownloadLink: ""}

		ep.DownloadLink = GetDownloadLinkForId(baseurl, token, ep.Id)
		result = append(result, ep)
	}

	return result, nil
}

func FilterEpisodesForSeason(episodes []Episode, seasonId string) []Episode {
	var episodesForSeason []Episode

	for _, episode := range episodes {
		if episode.SeasonId == seasonId {
			episodesForSeason = append(episodesForSeason, episode)
		}
	}

	return episodesForSeason
}

// TODO: Assume that episodes are unordered
func OrderSeasonsByEpisodes(episodes []Episode) []Season {
	var seasons []Season

	var lastEp Episode
	for _, episode := range episodes {

		seasonIsNew := lastEp.SeasonId != "" && lastEp.SeasonId != episode.SeasonId
		if len(seasons) == 0 || seasonIsNew {
			newSeason := Season{
				Id:       episode.SeasonId,
				Name:     episode.SeasonName,
				Episodes: append(make([]Episode, 1), episode)}

			seasons = append(seasons, newSeason)
		} else {
			seasons[len(seasons)-1].Episodes = append(seasons[len(seasons)-1].Episodes, episode)
		}

		lastEp = episode
	}

	return seasons
}
