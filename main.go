package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v3"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://127.0.0.1:8080/callback"

var (
	auth  = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))
	ch    = make(chan *spotify.Client)
	state = "abc123"
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

type TwitchArtist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BlockListState string `json:"blockListState"`
}

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

type TwitchResponse struct {
	Data struct {
		SearchDJCatalog struct {
			CatalogLastUpdatedAt string           `json:"catalogLastUpdatedAt"`
			Edges                []TwitchSongNode `json:"edges"`
		} `json:"searchDJCatalog"`
	} `json:"data"`
}

func sendRequest(songTitle string) ([]TwitchResponse, error) {
	url := "https://gql.twitch.tv/gql"

	// Create the payload
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

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Client-ID", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Content-Type", "application/json")

	// Send the request
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

func renderTable(edges []TwitchSongNode) {
	// Prepare data for the table
	tableData := [][]string{}
	for _, edge := range edges {
		song := edge.Node
		status := "Allowed"
		if song.IsBlockedTrack {
			status = "Blocked"
		}
		artists := []string{}
		for _, artist := range song.Artists {
			artists = append(artists, artist.Name)
		}
		tableData = append(tableData, []string{
			song.ID,
			song.Title,
			strings.Join(artists, ", "),
			strings.Join(song.Labels, ", "),
			strings.Join(song.Genres, ", "),
			fmt.Sprintf("%.2f", song.Duration),
			status,
		})
	}

	// Render the table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Title", "ID", "Artists", "Labels", "Genres", "Duration (s)", "Streamable"})
	table.AppendBulk(tableData)
	table.Render()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	cmd := &cli.Command{
		Name:  "isitstreamablebot",
		Usage: "Check if a song is streamable on Twitch",
		Commands: []*cli.Command{
			{
				Name:  "bot",
				Usage: "Start a discord bot to listen for song requests",
				Action: func(ctx context.Context, cmd *cli.Command) error {

					// Here you would start your Discord bot
					fmt.Println("Starting Discord bot...")

					return nil
				},
			},
			{
				Name:  "link",
				Usage: "Check if a spotify link is streamable on Twitch",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					link := cmd.Args().First()
					if link == "" {
						return fmt.Errorf("usage: %s <link>", cmd.Name)
					}

					// Extract Spotify ID from the link
					parts := strings.Split(link, "/")
					if len(parts) < 2 {
						return fmt.Errorf("invalid link format")
					}
					spotifyID := parts[len(parts)-1]
					if strings.Contains(spotifyID, "?") {
						spotifyID = strings.Split(spotifyID, "?")[0]
					}

					url := auth.AuthURL(state)
					fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

					// wait for auth to complete
					client := <-ch

					// Here you would check if the song is streamable using the Spotify API
					trackID := spotify.ID(spotifyID)
					track, err := client.GetTrack(context.Background(), trackID)
					if err != nil {
						log.Fatalf("Error getting track: %v", err)
						return nil
					}
					if track == nil {
						log.Println("Track not found.")
						return nil
					}

					fmt.Printf("Track: %s\n", track.Name)

					// Send the request
					response, err := sendRequest(track.Name)
					if err != nil {
						log.Fatalf("Error: %v", err)
					}

					for _, r := range response {
						if len(r.Data.SearchDJCatalog.Edges) == 0 {
							log.Println("No results found.")
							return nil
						}

						// Render the table
						renderTable(r.Data.SearchDJCatalog.Edges)

						log.Printf("Catalog Last Updated At: %s\n", r.Data.SearchDJCatalog.CatalogLastUpdatedAt)
					}

					return nil
				},
			},
			{
				Name:  "song",
				Usage: "Check if a song is streamable on Twitch",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()
					if len(args) < 1 {
						return fmt.Errorf("usage: %s <song title>", cmd.Name)
					}
					songTitle := strings.Join(args, " ")

					// Send the request
					response, err := sendRequest(songTitle)
					if err != nil {
						log.Fatalf("Error: %v", err)
					}

					for _, r := range response {
						if len(r.Data.SearchDJCatalog.Edges) == 0 {
							log.Println("No results found.")
							return nil
						}

						// Render the table
						renderTable(r.Data.SearchDJCatalog.Edges)

						log.Printf("Catalog Last Updated At: %s\n", r.Data.SearchDJCatalog.CatalogLastUpdatedAt)
					}
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}
