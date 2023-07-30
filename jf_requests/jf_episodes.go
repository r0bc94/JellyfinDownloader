package jf_requests

import "fmt"

type Episode struct {
	SeriesName   string
	Name         string
	Id           string
	SeasonId     string
	SeasonName   string
	Container    string
	DownloadLink string
}

func GetEpisodesFromId(token string, baseurl string, seriesId string) ([]Episode, error) {
	requestUrl := fmt.Sprintf("%s/Shows/%s/Episodes", baseurl, seriesId)

	res, err := MakeRequest(token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	items := res["Items"].([]any)
	var result []Episode
	for idx, item := range items {
		result = append(result, Episode{
			SeriesName:   item.(map[string]any)["SeriesName"].(string),
			Name:         item.(map[string]any)["Name"].(string),
			Id:           item.(map[string]any)["Id"].(string),
			SeasonId:     item.(map[string]any)["SeasonId"].(string),
			SeasonName:   item.(map[string]any)["SeasonName"].(string),
			Container:    item.(map[string]any)["Container"].(string),
			DownloadLink: ""})

		result[idx].PatchDownloadLink(baseurl, token)
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

func (ep *Episode) PatchDownloadLink(baseUrl string, token string) {
	ep.DownloadLink = fmt.Sprintf(baseUrl+"/Items/%s/Download?api_key=%s", ep.Id, token)
}
