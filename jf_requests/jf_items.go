package jf_requests

import (
	"fmt"
	"strings"
)

type Item struct {
	Name string
	Id   string
}

func GetItem(rawItems []any) []Item {
	var result []Item
	for _, item := range rawItems {
		itm := Item{
			Name: item.(map[string]any)["Name"].(string),
			Id:   item.(map[string]any)["Id"].(string),
		}

		result = append(result, itm)
	}

	return result
}

// Returns all Root Items
func GetRootItems(auth *AuthResponse, baseurl string) ([]Item, error) {
	requestUrl := baseurl + fmt.Sprintf("/Users/%s/Items", auth.UserId)

	res, err := MakeRequest(auth.Token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	items := res["Items"].([]any)
	return GetItem(items), nil
}

func GetItemsForParentId(auth *AuthResponse, baseurl string, parentId string) ([]Item, error) {
	requestUrl := baseurl + fmt.Sprintf("/Users/%s/Items?IncludeItemTypes=Series&ParentId=%s", auth.UserId, parentId)

	res, err := MakeRequest(auth.Token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	items := res["Items"].([]any)
	return GetItem(items), nil
}

// Returns all items found on the given jellyfin server.
func GetAllItems(auth *AuthResponse, baseurl string) ([]Item, error) {
	rootItems, err := GetRootItems(auth, baseurl)
	if err != nil {
		return nil, err
	}

	var items []Item = make([]Item, 0, 256)
	for _, rootItem := range rootItems {
		childItems, err := GetItemsForParentId(auth, baseurl, rootItem.Id)
		if err != nil {
			return nil, err
		}

		items = append(items, childItems...)
	}

	return items, nil

}

// Returns the item whose name includes the given search term.
func GetItemsForText(auth *AuthResponse, baseUrl string, searchtext string) ([]Item, error) {
	all, err := GetAllItems(auth, baseUrl)
	if err != nil {
		return nil, err
	}

	var results []Item
	for _, item := range all {
		if strings.Contains(strings.ToLower(item.Name), strings.ToLower(searchtext)) {
			results = append(results, item)
		}
	}

	return results, nil
}
