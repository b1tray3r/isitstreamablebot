package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// PersistedQuery represents the persisted query details
type PersistedQuery struct {
	Version    int    `json:"version"`
	Sha256Hash string `json:"sha256Hash"`
}

// Extensions represents the extensions in the payload
type Extensions struct {
	PersistedQuery PersistedQuery `json:"persistedQuery"`
}

// PayloadData represents the structure of the payload
type PayloadData struct {
	OperationName string     `json:"operationName"`
	Variables     Variables  `json:"variables"`
	Extensions    Extensions `json:"extensions"`
}

// Variables represents the variables in the payload
type Variables struct {
	SearchInput SearchInput `json:"searchInput"`
}

// SearchInput represents the search input details
type SearchInput struct {
	SearchType string `json:"searchType"`
	SortBy     string `json:"sortBy"`
	Term       string `json:"term"`
}

// TwitchArtist represents the artist details in the response
type TwitchArtist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BlockListState string `json:"blockListState"`
}

// TwitchSongNode represents the song node in the response
type TwitchSongNode struct {
	Cursor string `json:"cursor"`
	Node   struct {
		ID             string         `json:"id"`
		Title          string         `json:"title"`
		Artists        []TwitchArtist `json:"artists"`
		Labels         []string       `json:"labels"`
		IsBlockedTrack bool           `json:"isBlockedTrack"`
		Genres         []string       `json:"genres"`
		Duration       float64        `json:"duration"`
	}
}

// TwitchResponse represents the structure of the response from Twitch
type TwitchResponse struct {
	Data struct {
		SearchDJCatalog struct {
			CatalogLastUpdatedAt string           `json:"catalogLastUpdatedAt"`
			Edges                []TwitchSongNode `json:"edges"`
		} `json:"searchDJCatalog"`
	} `json:"data"`
}

// sendRequest sends a request to the Twitch API and returns the response
func SendRequest(songTitle string) ([]TwitchResponse, error) {
	url := "https://gql.twitch.tv/gql"

	// Create the payload
	// This is mostly taken from the Twitch Browser DevTools
	payloadData := []PayloadData{
		{
			OperationName: "DJMusicCatalogSearchQuery",
			Variables: Variables{
				SearchInput: SearchInput{
					SearchType: "TRACK",
					SortBy:     "BEST_MATCH",
					Term:       songTitle,
				},
			},
			Extensions: Extensions{
				PersistedQuery: PersistedQuery{
					Version:    1,
					Sha256Hash: "bf84de47edcd146c548e6fc10664441cf32940bab09af63a23f31bedc32bda2c",
				},
			},
		},
	}

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Client-ID", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer res.Body.Close()

	var response []TwitchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}
