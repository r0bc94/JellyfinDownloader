package jf_requests

import (
	"errors"
	"fmt"
	"strings"
)

type Item struct {
	Name string
	Id   string
	Type string
}

func GetItem(rawItems []any, parentItem *Item) []Item {
	var result []Item
	for _, item := range rawItems {
		itm := Item{
			Name: item.(map[string]any)["Name"].(string),
			Id:   item.(map[string]any)["Id"].(string),
			Type: item.(map[string]any)["Id"].(string),
		}

		if itmtype, ok := item.(map[string]any)["Type"].(string); ok {
			itm.Type = itmtype
		} else if parentItem != nil {
			itm.Type = parentItem.Type
		} else {
			itm.Type = ""
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
	return GetItem(items, nil), nil
}

func GetItemsForParentId(auth *AuthResponse, baseurl string, parentItem *Item) ([]Item, error) {
	requestUrl := baseurl + fmt.Sprintf("/Users/%s/Items?ParentId=%s", auth.UserId, parentItem.Id)

	res, err := MakeRequest(auth.Token, requestUrl, "GET", nil)
	if err != nil {
		return nil, err
	}

	items := res["Items"].([]any)
	return GetItem(items, parentItem), nil
}

// Returns all items found on the given jellyfin server.
func GetAllItems(auth *AuthResponse, baseurl string) ([]Item, error) {
	rootItems, err := GetRootItems(auth, baseurl)
	if err != nil {
		return nil, err
	}

	var items []Item = make([]Item, 0, 256)
	for _, rootItem := range rootItems {
		childItems, err := GetItemsForParentId(auth, baseurl, &rootItem)
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

func GetItemForId(auth *AuthResponse, baseurl string, id string) (*Item, error) {
	requestUrl := baseurl + fmt.Sprintf("/Users/%s/Items/%s", auth.UserId, id)
	res, err := MakeRequest(auth.Token, requestUrl, "GET", nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to find item with id: %s - %s", id, err))
	}

	resList := make([]any, 1, 1)
	resList[0] = res
	return &GetItem(resList, nil)[0], nil
}
