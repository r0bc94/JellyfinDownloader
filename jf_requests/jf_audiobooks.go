package jf_requests

import (
	"fmt"
	"log/slog"

	"github.com/fatih/color"
)

type Audiobook struct {
	Name     string
	Id       string
	Episodes []AudiobookEpisode
}

type AudiobookEpisode struct {
	Name        string
	Id          string
	CanDownload bool
	Filename    string
}

func GetAudioBookEpisodesFromItem(token string, baseurl string, item *Item) (*Audiobook, error) {
	requestUrl := baseurl + fmt.Sprintf("/Items?ParentId=%s&fields=candownload,path&recursive=true&mediaTypes=Audio", item.Id)

	res, err := MakeRequest(token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	items := res["Items"].([]any)
	var episodes []AudiobookEpisode = make([]AudiobookEpisode, 0)
	for _, itm := range items {
		casted_itm, ok := itm.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("failed to process response: %s", items)
		}

		episode := AudiobookEpisode{
			Name:        casted_itm["Name"].(string),
			Id:          casted_itm["Id"].(string),
			CanDownload: casted_itm["CanDownload"].(bool),
		}

		if filename, failed := GetFilenameForResponse(itm.(map[string]any)); failed != nil {
			color.Yellow("Did not found a filename for episode: \"%s\". It will be ignored..", episode.Name)
			slog.Debug(failed.Error())
			continue
		} else {
			episode.Filename = filename
		}

		episodes = append(episodes, episode)
	}

	audiobook := Audiobook{
		Id:       item.Id,
		Name:     item.Name,
		Episodes: episodes,
	}

	return &audiobook, nil
}

func (audiobook *Audiobook) PrintAndGetConfirmation() bool {
	fmt.Println("The following Audiobook will be downloaded:")
	color.Green(audiobook.Name)
	undownloadbleItemsPresent := false

	for episode_index, episode := range audiobook.Episodes {
		outstring := fmt.Sprintf("└ %d. %s", episode_index+1, episode.Name)

		// Strike out episodes which can not be downloaded from the Jellyfin server due to the CanDownload attribute
		// set to false
		if !episode.CanDownload {
			outstring = fmt.Sprintf("\033[9m%s\033[0m", outstring)
			undownloadbleItemsPresent = true
		}

		color.Cyan(outstring)
	}

	if undownloadbleItemsPresent {
		color.Yellow("Some items cannot be downloaded due to insufficient permission!")
		color.Yellow("The affected Items are struck through.")
	}

	return GetConfirmation()
}

func (audiobook *Audiobook) Download(baseUrl string, token string, keepFilenames bool) {
	for idx, episode := range audiobook.Episodes {
		if episode.CanDownload {
			var outfilename string
			if keepFilenames {
				outfilename = episode.Filename
			} else {
				suffix := GetSuffixFromFilename(episode.Filename)
				outfilename = fmt.Sprintf("%d - %s.%s", idx+1, episode.Name, suffix)
			}

			downloadLink := GetDownloadLinkForId(baseUrl, token, episode.Id)
			DownloadFromUrl(downloadLink, episode.Name, outfilename, len(audiobook.Episodes), idx)
		} else {
			color.Yellow("Skipping non downloadable item: %s", episode.Name)
		}
	}
}
