package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

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

// Extensions represents the extensions in the payload
type Extensions struct {
	PersistedQuery PersistedQuery `json:"persistedQuery"`
}

// PersistedQuery represents the persisted query details
type PersistedQuery struct {
	Version    int    `json:"version"`
	Sha256Hash string `json:"sha256Hash"`
}

func main() {
	url := "https://gql.twitch.tv/gql"

	// Create the payload
	payloadData := []PayloadData{
		{
			OperationName: "DJMusicCatalogSearchQuery",
			Variables: Variables{
				SearchInput: SearchInput{
					SearchType: "TRACK",
					SortBy:     "BEST_MATCH",
					Term:       "September",
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
		log.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Add headers
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("client-id", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer res.Body.Close()

	// Read the response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("%s", string(body))
}
