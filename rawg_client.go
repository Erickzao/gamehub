package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	baseAPIURL = "https://api.rawg.io/api"
)

type RAWGResponse struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []RAWGGame `json:"results"`
}

type RAWGGame struct {
	ID              int         `json:"id"`
	Name            string      `json:"name"`
	Released        string      `json:"released"`
	BackgroundImage string      `json:"background_image"`
	Rating          float64     `json:"rating"`
	RatingTop       int         `json:"rating_top"`
	Added           int         `json:"added"`
	Metacritic      int         `json:"metacritic"`
	Playtime        int         `json:"playtime"`
	Updated         string      `json:"updated"`
	ReviewsCount    int         `json:"reviews_count"`
	Description     string      `json:"description"`
	Genres          []Genre     `json:"genres"`
	Platforms       []Platform  `json:"platforms"`
	Stores          []RAWGStore `json:"stores"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Platform struct {
	Platform PlatformDetail `json:"platform"`
}

type PlatformDetail struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RAWGStore struct {
	Store RAWGStoreDetail `json:"store"`
}

type RAWGStoreDetail struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Image string `json:"image_background"`
}

func generateStoreURL(storeName string, gameName string) string {
	encodedName := url.QueryEscape(gameName)
	switch strings.ToLower(storeName) {
	case "steam":
		return fmt.Sprintf("https://store.steampowered.com/search/?term=%s", encodedName)
	case "playstation store":
		return fmt.Sprintf("https://store.playstation.com/search/%s", encodedName)
	case "xbox store":
		return fmt.Sprintf("https://www.xbox.com/games/search?q=%s", encodedName)
	case "nintendo":
		return fmt.Sprintf("https://www.nintendo.com/search/?q=%s&p=1&cat=gme", encodedName)
	case "gog":
		return fmt.Sprintf("https://www.gog.com/games?query=%s", encodedName)
	case "epic games":
		return fmt.Sprintf("https://store.epicgames.com/browse?q=%s", encodedName)
	default:
		return fmt.Sprintf("https://www.google.com/search?q=buy+%s+game", encodedName)
	}
}

func getStores(stores []RAWGStore, gameName string) []Store {
	var storeList []Store
	for _, store := range stores {
		storeURL := generateStoreURL(store.Store.Name, gameName)
		storeList = append(storeList, Store{
			ID:    store.Store.ID,
			Name:  store.Store.Name,
			URL:   storeURL,
			Image: store.Store.Image,
		})
	}
	return storeList
}

func validateURL(endpoint string) error {
	if !strings.HasPrefix(endpoint, "/") {
		return fmt.Errorf("endpoint must start with /")
	}

	validEndpoint := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '/' || r == '-' || r == '_' || r == '?' || r == '=' || r == '&':
			return r
		}
		return -1
	}, endpoint)

	if validEndpoint != endpoint {
		return fmt.Errorf("invalid characters in endpoint")
	}

	return nil
}

func makeRequest(urlStr string) (*http.Response, error) {
	_, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "GameHub/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return resp, nil
}

func fetchGames(endpoint string) ([]Game, error) {
	if err := validateURL(endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	apiKey := os.Getenv("RAWG_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RAWG_API_KEY not configured")
	}

	urlStr := fmt.Sprintf("%s%s", baseAPIURL, endpoint)
	if strings.Contains(endpoint, "?") {
		urlStr = fmt.Sprintf("%s&key=%s&page_size=20", urlStr, apiKey)
	} else {
		urlStr = fmt.Sprintf("%s?key=%s&page_size=20", urlStr, apiKey)
	}

	resp, err := makeRequest(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var rawgResp RAWGResponse
	if err := json.Unmarshal(body, &rawgResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	var games []Game
	for _, rawgGame := range rawgResp.Results {
		game := Game{
			ID:              fmt.Sprintf("%d", rawgGame.ID),
			Title:           rawgGame.Name,
			Description:     rawgGame.Description,
			BackgroundImage: rawgGame.BackgroundImage,
			Genres:          getGenres(rawgGame.Genres),
			Rating:          rawgGame.Rating,
			RatingTop:       rawgGame.RatingTop,
			ReleaseDate:     rawgGame.Released,
			Added:           rawgGame.Added,
			Metacritic:      rawgGame.Metacritic,
			Playtime:        rawgGame.Playtime,
			Updated:         rawgGame.Updated,
			Reviews:         rawgGame.ReviewsCount,
			Platforms:       getPlatforms(rawgGame.Platforms),
			Stores:          getStores(rawgGame.Stores, rawgGame.Name),
		}
		games = append(games, game)
	}

	return games, nil
}

func getGenres(genres []Genre) []string {
	var genreNames []string
	for _, genre := range genres {
		genreNames = append(genreNames, genre.Name)
	}
	return genreNames
}

func getPlatforms(platforms []Platform) []string {
	var platformNames []string
	for _, platform := range platforms {
		platformNames = append(platformNames, platform.Platform.Name)
	}
	return platformNames
}

func fetchGameByID(id string) (*Game, error) {
	if id == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/games/%s", id)
	if err := validateURL(endpoint); err != nil {
		return nil, fmt.Errorf("invalid game ID: %w", err)
	}

	apiKey := os.Getenv("RAWG_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RAWG_API_KEY not configured")
	}

	urlStr := fmt.Sprintf("%s%s?key=%s", baseAPIURL, endpoint, apiKey)

	resp, err := makeRequest(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var rawgGame RAWGGame
	if err := json.Unmarshal(body, &rawgGame); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	game := &Game{
		ID:              fmt.Sprintf("%d", rawgGame.ID),
		Title:           rawgGame.Name,
		Description:     rawgGame.Description,
		BackgroundImage: rawgGame.BackgroundImage,
		Genres:          getGenres(rawgGame.Genres),
		Rating:          rawgGame.Rating,
		RatingTop:       rawgGame.RatingTop,
		ReleaseDate:     rawgGame.Released,
		Added:           rawgGame.Added,
		Metacritic:      rawgGame.Metacritic,
		Playtime:        rawgGame.Playtime,
		Updated:         rawgGame.Updated,
		Reviews:         rawgGame.ReviewsCount,
		Platforms:       getPlatforms(rawgGame.Platforms),
		Stores:          getStores(rawgGame.Stores, rawgGame.Name),
	}

	return game, nil
} 