package jf_requests

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func GetConfirmation() bool {
	fmt.Print("Continue? y/n: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y"
}

func GetUserChoice(number_of_choices int) (int, error) {
	fmt.Print("==> ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.Split(response, "\n")[0]
	if selection, err := strconv.Atoi(response); err == nil {
		if selection < 0 || selection > number_of_choices {
			return -1, errors.New("Invalid Selection")
		}

		return selection, nil
	} else {
		slog.Error(err.Error())
		return -1, errors.New("Only provide a single number")
	}
}
