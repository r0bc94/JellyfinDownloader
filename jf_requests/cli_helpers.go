package jf_requests

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetConfirmation() bool {
	fmt.Print("Continue? y/n: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y"
}
