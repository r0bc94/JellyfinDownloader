package jf_requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type AuthRequestBody struct {
	Username string
	Pw       string
}

type AuthResponse struct {
	Token  string
	UserId string
}

func ExecuteRequest(request *http.Request) (map[string]any, error) {
	client := &http.Client{}
	res, err := client.Do(request)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Request Failed: %s", err))
	}

	defer res.Body.Close()

	var content_raw []byte
	content_raw, err = io.ReadAll(res.Body)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not read response body: %s", err))
	} else if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Request Failed (Code %d): %s", res.StatusCode, content_raw))
	}

	var content_json map[string]any
	err = json.Unmarshal(content_raw, &content_json)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to Parse JSON from Response: %s", err))
	}

	return content_json, nil

}

// Authorizes the given user with the provided password against the given Jellyfin hostname
// When successfull, an auth token wich can be used for further requests is returned.
func Authorize(baseUrl string, username string, password string) (*AuthResponse, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	sanitizedBaseUrl := baseUrl
	// Strip the leading / from the baseurl, if there is any
	if string(baseUrl[len(baseUrl)-1]) == "/" {
		sanitizedBaseUrl = baseUrl[:len(baseUrl)-1]
	}

	requestUrl := fmt.Sprintf("%s/Users/AuthenticateByName", sanitizedBaseUrl)

	// Create Request Body with Credentials
	reqbody := &AuthRequestBody{Username: username, Pw: password}
	reqbody_json, err := json.Marshal(reqbody)

	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(reqbody_json))
	req.Header.Set("Content-Type", "application/json")

	// Fix Header by inserting the Authorization header with artificial Values
	emby_auth_header := "MediaBrowser Client=\"Go\", Device=\"Test\", DeviceId=\"Test\", Version=\"1.0.0\""
	req.Header.Set("X-Emby-Authorization", emby_auth_header)

	response, err := ExecuteRequest(req)

	if err != nil {
		return nil, err
	}

	accessToken := response["AccessToken"].(string)
	userId := response["SessionInfo"].(map[string]any)["UserId"].(string)
	return &AuthResponse{Token: accessToken, UserId: userId}, nil
}

func MakeRequest(token string, requestUrl string, method string, body any) (map[string]any, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Create Request Body
	reqbody_json, err := json.Marshal(body)

	req, err := http.NewRequest(method, requestUrl, bytes.NewBuffer(reqbody_json))
	req.Header.Set("Content-Type", "application/json")

	// Fix Header by inserting the Authorization header with artificial Values
	emby_auth_header := "MediaBrowser Client=\"Go\", Device=\"Test\", DeviceId=\"Test\", Version=\"1.0.0\""
	emby_auth_header += fmt.Sprintf(", Token=\"%s\"", token)

	req.Header.Set("X-Emby-Authorization", emby_auth_header)

	result, err := ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	return result, nil
}
