package main

import (
	"bufio"
	"flag"
	"fmt"
	"jf_requests/jf_requests"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/lmittmann/tint"
	"golang.org/x/term"
)

const VERSION string = "v1.4.0-prerelease-1"

type Arguments struct {
	BaseUrl       string
	Username      string
	Password      string
	SeriesId      string
	SeasonId      string
	Name          string
	KeepFilenames bool
	Version       bool
	Debug         bool
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
	flag.BoolVar(&args.KeepFilenames, "keepFilenames", false, "Keeps the original episode filenames for series.")
	flag.BoolVar(&args.Version, "version", false, "Shows the Version Informations and Exit")
	flag.BoolVar(&args.Debug, "debug", false, "Show verbose debug output which may be useful to find certain problems")

	flag.Parse()

	return &args
}

// Checks, if all necessarry cli arguments are passed.
func CheckArguments(args *Arguments) (bool, string) {
	if args.BaseUrl == "" {
		return false, "No URL was given. See -h for more information"
	}

	// Check if the URL was specified in the correct format.
	urlpattern := `https?\:\/\/[\d\w._-]+(:\d+)?\/?([/\d\w._-]*?)?$`
	match, err := regexp.Match(urlpattern, []byte(args.BaseUrl))
	if !match || err != nil {
		return false, "URL was supplied in the wrong pattern. The URL must be supplied like so: http(s)://myserver(:123)(/). Instead of the whole hostname, you can also specify the IPv4 address which is pointing to your Jellyfin server."
	}

	// Remove a leading / if it was provided
	args.BaseUrl = strings.TrimSuffix(args.BaseUrl, "/")

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

	if runtime.GOOS == "windows" {
		return strings.TrimSuffix(username, "\r\n")

	}

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

func PrintItemSelection(itemsToSelect []jf_requests.Item) (*jf_requests.Item, error) {
	fmt.Println("Found multiple Shows for the given Searchterm. Please Select the show you want to download:")

	for idx, show := range itemsToSelect {
		color.Cyan("  %d. %s", idx+1, show.Name)
	}

	choice, err := jf_requests.GetUserChoice(len(itemsToSelect))
	if err != nil {
		return nil, err
	}

	return &itemsToSelect[choice-1], nil
}

func DownloadSeries(auth *jf_requests.AuthResponse, baseurl string, item *jf_requests.Item, seasonId string, keepFilenames bool) bool {
	series, err := jf_requests.GetSeriesFromItem(auth.Token, baseurl, item)
	if err != nil {
		color.Red("Failed to obtain Episode Information for given id: %s", err)
		return false
	}

	color.Green("Series: %s\n", item.Name)
	var selected_seasons []jf_requests.Season
	if seasonId != "" {
		if selected_season, geterr := series.GetSeasonForId(seasonId); geterr == nil {
			selected_seasons = []jf_requests.Season{*selected_season}
		} else {
			err = geterr
		}

	} else {
		selected_seasons, err = series.PrintAndGetSelection()
	}

	if err != nil {
		color.Red(err.Error())
		return false
	}

	confirm := series.PrintAndGetConfirmation(selected_seasons)

	if confirm {
		for _, season := range selected_seasons {
			season.Download(baseurl, auth.Token, keepFilenames)
		}
	}

	return true
}

func DownloadMovie(auth *jf_requests.AuthResponse, baseurl string, item *jf_requests.Item, keepFilename bool) bool {
	movie, err := jf_requests.GetMovieFromItem(auth, baseurl, item)
	if err != nil {
		color.Red("Failed to obtain Movie for given id: %s", err)
		return false
	}

	if movie.PrintAndGetConfirmation() {
		movie.Download(keepFilename)
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
			return DownloadSeries(auth, args.BaseUrl, item, args.SeasonId, args.KeepFilenames)
		} else {
			return DownloadMovie(auth, args.BaseUrl, item, args.KeepFilenames)
		}

	} else if args.Name != "" {
		items, err := jf_requests.GetItemsForText(auth, args.BaseUrl, args.Name)
		if err != nil {
			color.Red("Failed to obtain Episode Information for given id: %s", err)
			return false
		}

		var item *jf_requests.Item
		if len(items) == 0 {
			color.Yellow("Did not found anything for the given Searchterm on the Server.")
			return false
		} else if len(items) == 1 {
			item = &items[0]
		} else {
			item, err = PrintItemSelection(items)
			if err != nil {
				color.Red(err.Error())
				return false
			}
		}

		if item.Type == "Series" {
			return DownloadSeries(auth, args.BaseUrl, item, args.SeasonId, args.KeepFilenames)
		} else {
			return DownloadMovie(auth, args.BaseUrl, item, args.KeepFilenames)
		}

	}

	return false
}

func ShowVersionInfo() {
	fmt.Printf("JellyfinDownloader Version: %s\n", VERSION)
}

func getLogLevel(args *Arguments) slog.Level {
	if args.Debug {
		return slog.LevelDebug
	} else {
		return slog.LevelInfo
	}
}

func main() {
	args := ParseCLIArgs()

	// Configure Logger
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      getLogLevel(args),
			TimeFormat: time.Kitchen,
		}),
	))

	if args.Version {
		ShowVersionInfo()
		os.Exit(0)
	}

	if status, msg := CheckArguments(args); !status {
		color.Red("Wrong Arguments: %s\n", msg)
		os.Exit(1)
	}

	username := GetUsername(args)
	password := GetPassword(args)

	creds, err := jf_requests.Authorize(args.BaseUrl, username, password)
	if err != nil {
		color.Red("Authentication Failed! Did you enter the correct credentials?")
		os.Exit(1)
	}

	result := Download(args, creds)
	if !result {
		os.Exit(1)
	}
}
