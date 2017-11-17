package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Search struct {
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	CreatedAt           string `json:"created_at"`
	CreatedBy           string `json:"created_by"`
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	MostPopularMediaURL string `json:"most_popular_media_url"`
}

func SearchForGame(platform, name string) (string, error) {
	data := url.Values{}
	data.Set("name", name)

	url := fmt.Sprintf("http://retrogaming.cloud/api/v1/platform/%s/game?%s", platform, data.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(resp.Body)
	var search Search
	err = decoder.Decode(&search)
	if err != nil {
		return "", err
	}

	if len(search.Results) == 0 {
		return "", fmt.Errorf("Game not found.")
	}

	return search.Results[0].MostPopularMediaURL, nil
}
