package twitchalerts

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

// Events - This is used to pass custom events into the EventHandler.
// To use create a EventHandler struct and then implement the Events interface to the struct
type Events interface {
	OnStream(streamer string)
	OnError(error string)
}

// StreamRes - Response from the Twitch API
type StreamRes struct {
	Data       []StreamData `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

// StreamData - Included in the response, one for each stream the user currently has going
type StreamData struct {
	Id           string    `json:"id"`
	UserId       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameId       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	StreamType   string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  uint32    `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailUrl string    `json:"thumbnail_url"`
	TagsIds      *[]string `json:"tags_ids"`
	Tags         *[]string `json:"tags"`
	IsMature     bool      `json:"is_mature"`
}

// Pagination - Included in the response
type Pagination struct {
	Cursor string `json:"cursor"`
}

// Config - Stores all persistently required data
// Created if it does not exist when package starts
type Config struct {
	UserId    string   `json:"user_id"`
	Token     string   `json:"token"`
	Streamers []string `json:"streamers"`
	Delay     uint32   `json:"delay"`
}

var currentlyStreaming []string
var config Config

// Run - Main function loads and runs the Package
func Run(handler Events) {
	// Gets Working Directory
	getwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	// Gets Config File
	dat, err := os.ReadFile(getwd + "/TwitchAlertsConfig.json")

	// Check if Config File Exists
	if os.IsNotExist(err) {
		// Create Config File
		file, err := os.Create(getwd + "/TwitchAlertsConfig.json")

		if err != nil {
			log.Fatalln(err)
		}

		// Default Config
		config := Config{
			UserId:    "",
			Token:     "",
			Delay:     80,
			Streamers: []string{},
		}

		// Turn struct to JSON
		marshal, err := json.Marshal(config)
		if err != nil {
			log.Fatalln(err)
		}

		// Write Config File
		_, err = file.WriteString(string(marshal))
		if err != nil {
			log.Fatalln(err)
		}

		log.Fatalln("Config Created Please input your id, token, and streamers to watch!")
	}

	// Turn JSON into Config
	err = json.Unmarshal(dat, &config)

	if err != nil {
		handler.OnError(err.Error())
		return
	}

	// Makes the Map that is used for cache
	var cache = make(map[string]time.Time)
	// Loops continuously while program runs to keep checker going
	for {
		// Loops through streamers in config
		for _, streamer := range config.Streamers {
			// Checks if cooldown of 30 seconds has expired
			if cache[streamer].Before(time.Now()) {
				// Deletes streamer from cache
				delete(cache, streamer)
				// Runs checker function on another thread
				go checkStreamer(strings.Clone(streamer), &handler)
				// Adds streamer back to cache
				cache[streamer] = time.Now().Add(time.Duration(30) * time.Second)
			}

			// Delay Between requests
			time.Sleep(time.Duration(config.Delay))
		}

	}
}

// Function used for checking if a streamer is live
func checkStreamer(streamer string, eh *Events) {

	// Makes Token Bearer String
	var bearer = "Bearer " + strings.Clone(config.Token)

	// Base Request
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?user_login="+strings.Clone(streamer), nil)

	if err != nil {
		var handler Events
		handler = *eh
		handler.OnError(err.Error())
		return
	}

	// Adds Authorisation Headers to Request
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Client-Id", strings.Clone(config.UserId))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		var handler Events
		handler = *eh
		handler.OnError(err.Error())
		return
	}

	// Reads Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		var handler Events
		handler = *eh
		handler.OnError(err.Error())
		return
	}

	var stream StreamRes

	// Turns JSON Body into StreamRes
	err = json.Unmarshal(body, &stream)

	// If no active Streams
	if len(stream.Data) == 0 {
		// If currentlyStreaming contains the streamer remove them
		if slices.Contains(currentlyStreaming, strings.Clone(stream.Data[0].UserName)) {
			var pos = indexOf(strings.Clone(streamer), currentlyStreaming)
			removeIndex(currentlyStreaming, pos)
		}
		return
	}

	// If no errors and there is an active stream
	if err == nil {
		// if steamer is in currentlyStreaming just return
		if slices.Contains(currentlyStreaming, strings.Clone(stream.Data[0].UserName)) {
			return
		}

		currentlyStreaming = append(currentlyStreaming, strings.Clone(stream.Data[0].UserName))

		var handler Events
		handler = *eh

		handler.OnStream(stream.Data[0].UserName)
	}
}

// Removes an item from the array at an index
func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

// Finds the index of a given item
func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}
