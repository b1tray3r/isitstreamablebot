package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	Version    = 1
	Sha256Hash = "bf84de47edcd146c548e6fc10664441cf32940bab09af63a23f31bedc32bda2c"

	SearchTypeTrack = "TRACK"
	SortByBestMatch = "BEST_MATCH"

	TwitchGQLURL           = "https://gql.twitch.tv/gql"
	CatalogSearchOperation = "DJMusicCatalogSearchQuery"
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

// TwitchSong represents the song details in the response
type TwitchSong struct {
	Cursor string         `json:"cursor"`
	Node   TwitchSongNode `json:"node"`
}

type TwitchSongNode struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Artists        []TwitchArtist `json:"artists"`
	Labels         []string       `json:"labels"`
	IsBlockedTrack bool           `json:"isBlockedTrack"`
	Genres         []string       `json:"genres"`
	Duration       float64        `json:"duration"`
}

// GetArtists returns a comma-separated string of artist names
func (t *TwitchSongNode) GetArtists() string {
	artists := ""
	for _, artist := range t.Artists {
		artists += artist.Name + ", "
	}
	artists = artists[:len(artists)-2]
	return artists
}

// TwitchResponse represents the structure of the response from Twitch
type TwitchResponse struct {
	Data struct {
		SearchDJCatalog struct {
			CatalogLastUpdatedAt string       `json:"catalogLastUpdatedAt"`
			Edges                []TwitchSong `json:"edges"`
		} `json:"searchDJCatalog"`
	} `json:"data"`
}

// GetSongList extracts the song list from the Twitch response
func (r *TwitchResponse) GetSongList() []TwitchSongNode {
	songList := []TwitchSongNode{}
	for _, edge := range r.Data.SearchDJCatalog.Edges {
		songList = append(songList, edge.Node)
	}
	return songList
}

type TwitchDJCatalogRequest struct {
	OperationName string
	SearchType    string
	SortBy        string
	Term          string
	Version       int
	Sha256Hash    string
}

func NewTwitchDJCatalogRequest(searchType, sortBy, term string) *TwitchDJCatalogRequest {
	return &TwitchDJCatalogRequest{
		OperationName: CatalogSearchOperation,
		SearchType:    searchType,
		SortBy:        sortBy,
		Term:          term,
		Version:       Version,
		Sha256Hash:    Sha256Hash,
	}
}

func (r *TwitchDJCatalogRequest) GetPayload() ([]byte, error) {
	payload := PayloadData{
		OperationName: r.OperationName,
		Variables: Variables{
			SearchInput: SearchInput{
				SearchType: r.SearchType,
				SortBy:     r.SortBy,
				Term:       r.Term,
			},
		},
		Extensions: Extensions{
			PersistedQuery: PersistedQuery{
				Version:    r.Version,
				Sha256Hash: r.Sha256Hash,
			},
		},
	}

	return json.Marshal(payload)
}

// sendRequest sends a request to the Twitch API and returns the response
func SendRequest(songTitle string) (*TwitchResponse, error) {
	payload := NewTwitchDJCatalogRequest(SearchTypeTrack, SortByBestMatch, songTitle)
	payloadData, err := payload.GetPayload()
	if err != nil {
		return nil, fmt.Errorf("failed to get payload: %v", err)
	}

	req, err := http.NewRequest("POST", TwitchGQLURL, bytes.NewReader(payloadData))
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

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", res.Status)
	}

	var response TwitchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}
