package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"jf_requests/jf_requests"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"golang.org/x/term"
)

type Arguments struct {
	BaseUrl  string
	Username string
	Password string
	SeriesId string
	SeasonId string
	Name     string
}

// Parses the command line arguments and returns a struct containing all found arguments.
func ParseCLIArgs() *Arguments {
	var args = Arguments{}

	flag.StringVar(&args.BaseUrl, "url", "", "Base URL which points to the Jellyfin Instance")
	flag.StringVar(&args.SeriesId, "seriesid", "", "ID which points to the series which should be downloaded")
	flag.StringVar(&args.SeasonId, "seasonid", "", "If given, only the episodes with the provided season Id will be downloaded")
	flag.StringVar(&args.Username, "username", "", "Username used to login to the Jellyfin instance. If not provided, password will be prompted.")
	flag.StringVar(&args.Password, "password", "", "Passwort for the Jellyfin instance. If not provided, username will be prompted.")
	flag.StringVar(&args.Name, "name", "", "Name of the Show or Movie you want to download.")

	flag.Parse()

	return &args
}

// Checks, if all necessarry cli arguments are passed.
func CheckArguments(args *Arguments) (bool, string) {
	if args.BaseUrl == "" {
		return false, "No URL was given. See -h for more information"
	}

	if args.SeriesId == "" && args.Name == "" {
		return false, "No SeriesID or Name was given. See -h for more information."
	}

	return true, ""
}

func GetUsername(args *Arguments) string {
	if args.Username != "" {
		return args.Username
	} else if username := os.Getenv("JF_USERNAME"); username != "" {
		return username
	}

	fmt.Printf("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')

	return strings.TrimSuffix(username, "\n")
}

func GetPassword(args *Arguments) string {
	if args.Password != "" {
		return args.Password
	} else if password := os.Getenv("JF_PASSWORD"); password != "" {
		return password
	}

	fmt.Printf("Password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	return string(bytePassword)
}

func GetConfirmation() bool {
	fmt.Print("Continue? y/n: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y"
}

func PrintMovieSummary(movie *jf_requests.Movie) bool {
	fmt.Println("The following Movie will be downloaded:")
	color.Green("Name: %s", movie.Name)

	return GetConfirmation()
}

func PrintSeasonSelection(seasons []jf_requests.Season) (string, error) {
	fmt.Println("Which Seasons do you want to download:")

	color.Cyan("  0. All")
	for idx, season := range seasons {
		color.Cyan("  %d. %s", idx+1, season.Name)
	}

	fmt.Print("==> ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.Split(response, "\n")[0]
	if selection, err := strconv.Atoi(response); err == nil {
		if selection < 0 || selection > len(seasons) {
			return "", errors.New("Invalid Selection")
		} else if selection == 0 {
			return "", nil
		}

		return seasons[selection-1].Id, nil
	} else {
		fmt.Println(err)
		return "", errors.New("Only provide a single number")
	}
}

func PrintSeriesSummary(episodes []jf_requests.Episode) bool {
	fmt.Println("The following Episodes will be downloaded:")
	color.Green("Series: %s", episodes[0].SeriesName)
	color.Green("Episodes:")
	for idx, episode := range episodes {
		color.Cyan("  %d. %s", idx+1, episode.Name)
	}

	return GetConfirmation()
}

func PrintItemSelection(itemsToSelect []jf_requests.Item) (*jf_requests.Item, error) {
	fmt.Println("Found multiple Shows for the given Searchterm. Please Select the show you want to download:")

	for idx, show := range itemsToSelect {
		color.Cyan("  %d. %s", idx+1, show.Name)
	}

	fmt.Print("==> ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.Split(response, "\n")[0]
	if selection, err := strconv.Atoi(response); err == nil {
		if selection < 0 || selection > len(itemsToSelect) {
			return nil, errors.New("Invalid Selection")
		}

		return &itemsToSelect[selection-1], nil
	} else {
		fmt.Println(err)
		return nil, errors.New("Only provide a single number")
	}
}

func DownloadSeries(auth *jf_requests.AuthResponse, baseurl string, item *jf_requests.Item, seasonId string) bool {
	episodes, err := jf_requests.GetEpisodesFromId(auth.Token, baseurl, item.Id)
	if err != nil {
		color.Red("Failed to obtain Episode Information for given id: %s", err)
		return false
	}

	seasons := jf_requests.OrderSeasonsByEpisodes(episodes)

	if seasonId == "" {
		seasonId, _ = PrintSeasonSelection(seasons)
	}

	if seasonId != "" {
		episodes = jf_requests.FilterEpisodesForSeason(episodes, seasonId)
	}

	if PrintSeriesSummary(episodes) {
		jf_requests.DownloadEpisodes(episodes)
	} else {
		return false
	}

	return true
}

func DownloadMovie(auth *jf_requests.AuthResponse, baseurl string, item *jf_requests.Item) bool {
	movie, err := jf_requests.GetMovieFromItem(auth, baseurl, item)
	if err != nil {
		color.Red("Failed to obtain Movie for given id: %s", err)
		return false
	}

	if PrintMovieSummary(movie) {
		jf_requests.DownloadMovie(movie)
	} else {
		return false
	}

	return true
}

func Download(args *Arguments, auth *jf_requests.AuthResponse) bool {
	if args.SeriesId != "" {
		item, err := jf_requests.GetItemForId(auth, args.BaseUrl, args.SeriesId)
		if err != nil {
			color.Red("Failed to obtain items for given id: %s", err)
			return false
		}

		if item.Type == "Series" {
			return DownloadSeries(auth, args.BaseUrl, item, args.SeasonId)
		} else {
			return DownloadMovie(auth, args.BaseUrl, item)
		}

	} else if args.Name != "" {
		items, err := jf_requests.GetItemsForText(auth, args.BaseUrl, args.Name)
		if err != nil {
			color.Red("Failed to obtain Episode Information for given id: %s", err)
			return false
		}

		item, err := PrintItemSelection(items)
		if err != nil {
			color.Red(err.Error())
			return false
		}

		if item.Type == "Series" {
			return DownloadSeries(auth, args.BaseUrl, item, "")
		} else {
			return DownloadMovie(auth, args.BaseUrl, item)
		}

	}

	return false
}

func main() {
	args := ParseCLIArgs()

	if status, msg := CheckArguments(args); !status {
		color.Red("Wrong Arguments: %s\n", msg)
		os.Exit(1)
	}

	username := GetUsername(args)
	password := GetPassword(args)

	creds, err := jf_requests.Authorize(args.BaseUrl, username, password)
	if err != nil {
		color.Red("Authentication Failed! Maybe wrong credentials provided?")
		os.Exit(1)
	}

	result := Download(args, creds)
	if !result {
		os.Exit(1)
	}
}
