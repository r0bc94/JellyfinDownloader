package main

import (
	"bufio"
	"flag"
	"fmt"
	"jf_requests/jf_requests"
	"os"
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
}

// Parses the command line arguments and returns a struct containing all found arguments.
func ParseCLIArgs() *Arguments {
	var args = Arguments{}

	flag.StringVar(&args.BaseUrl, "url", "", "Base URL which points to the Jellyfin Instance")
	flag.StringVar(&args.SeriesId, "seriesid", "", "ID which points to the series which should be downloaded")
	flag.StringVar(&args.SeasonId, "seasonid", "", "If given, only the episodes with the provided season Id will be downloaded")
	flag.StringVar(&args.Username, "username", "", "Username used to login to the Jellyfin instance. If not provided, password will be prompted.")
	flag.StringVar(&args.Password, "password", "", "Passwort for the Jellyfin instance. If not provided, username will be prompted.")

	flag.Parse()

	return &args
}

// Checks, if all necessarry cli arguments are passed.
func CheckArguments(args *Arguments) (bool, string) {
	if args.BaseUrl == "" {
		return false, "No URL was given. See -h for more information"
	}

	if args.SeriesId == "" {
		return false, "No SeriesID was given. See -h for more information."
	}

	return true, ""
}

func GetUsername(args *Arguments) string {
	if args.Username != "" {
		return args.Username
	}

	fmt.Printf("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')

	return strings.TrimSuffix(username, "\n")
}

func GetPassword(args *Arguments) string {
	if args.Password != "" {
		return args.Password
	}

	fmt.Printf("Password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	return string(bytePassword)
}

func PrintSummarry(episodes []jf_requests.Episode) bool {
	fmt.Println("The following Episodes will be downloaded:")
	color.Green("Series: %s", episodes[0].SeriesName)
	color.Green("Episodes:")
	for idx, episode := range episodes {
		color.Cyan("  %d. %s", idx, episode.Name)
	}

	fmt.Print("Continue? y/n: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y"
}

func GetEpisodesToDownload(creds *jf_requests.AuthResponse, args *Arguments) ([]jf_requests.Episode, error) {
	episodes, err := jf_requests.GetEpisodesFromId(creds.Token, args.BaseUrl, args.SeriesId)
	if err != nil {
		return nil, err
	}

	if args.SeasonId != "" {
		return jf_requests.FilterEpisodesForSeason(episodes, args.SeasonId), nil
	}

	return episodes, nil

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
		fmt.Println(err)
		os.Exit(1)
	}

	episodesToDownload, err := GetEpisodesToDownload(creds, args)
	if err != nil {
		color.Red("Failed to obtain episodes to download: %s", err)
		os.Exit(1)
	}

	shouldDownload := PrintSummarry(episodesToDownload)

	if shouldDownload {
		jf_requests.Download(episodesToDownload)
	}

}
