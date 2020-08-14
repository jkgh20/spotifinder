package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"otherside/api/seatgeekLayer"
	"otherside/api/spotifyLayer"
	"strings"
	"time"

	"otherside/api/redisLayer"

	"github.com/gorilla/mux"
	"github.com/zmb3/spotify"
)

var applicationPort = "8081"

var cityPostcodeMap map[string]string
var availableGenres []string

func main() {
	cityPostcodeMap = generateCityPostcodeMap()
	availableGenres = generateGenres()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/cities", Cities)
	router.HandleFunc("/genres", Genres)
	router.HandleFunc("/authenticate", Authenticate)
	router.HandleFunc("/callback", Callback)
	router.HandleFunc("/localevents", LocalEvents)
	router.HandleFunc("/toptracks", TopTracks).Methods("POST")
	router.HandleFunc("/buildplaylist", BuildPlaylist).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+applicationPort, router))
}

func generateCityPostcodeMap() map[string]string {
	postCodeMap := map[string]string{
		"Austin TX":     "78759",
		"Atlanta GA":    "30301",
		"Washington DC": "20001",
		"Nashville TN":  "37011",
	}

	return postCodeMap
}

func generateGenres() []string {
	genres := []string{
		"rock",
		"indie",
		"hip-hop",
		"jazz",
		"pop",
	}

	return genres
}

//GET
func Cities(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	var cities []string
	for key := range cityPostcodeMap {
		cities = append(cities, key)
	}

	citiesJSON, err := json.Marshal(cities)
	if err != nil {
		fmt.Printf("Error Marshalling city keys: " + err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(citiesJSON)
	}
}

//GET
func Genres(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	genresJSON, err := json.Marshal(availableGenres)
	if err != nil {
		fmt.Printf("Error Marshalling city keys: " + err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(genresJSON)
	}
}

//GET
func LocalEvents(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	cities, ok := r.URL.Query()["cities"]
	if !ok || len(cities[0]) < 1 {
		fmt.Printf("cities parameter missing from localevents request.")
	}

	genres, ok := r.URL.Query()["genres"]
	if !ok || len(genres[0]) < 1 {
		fmt.Printf("genres parameter missing from localevents request.")
	}

	citiesArray := QueryStringToArray(cities[0])
	genreArray := QueryStringToArray(genres[0])

	var postCodeArray []string

	for _, val := range citiesArray {
		postCodeArray = append(postCodeArray, cityPostcodeMap[val])
	}

	localSeatGeekEvents := seatgeekLayer.FindLocalEvents(postCodeArray, genreArray)

	localSeatGeekEventsJSON, err := json.Marshal(localSeatGeekEvents)
	if err != nil {
		fmt.Printf("Error Marshalling localseatgeekevents data: " + err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(localSeatGeekEventsJSON)
	}
}

func QueryStringToArray(queryString string) []string {

	testsArray := strings.Split(strings.Trim(queryString, "[]"), ",")
	return testsArray
}

//POST
func TopTracks(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	var localSeatGeekEvents []seatgeekLayer.SeatGeekEvent
	var topTracks []spotify.FullTrack

	err := json.NewDecoder(r.Body).Decode(&localSeatGeekEvents)
	if err != nil {
		fmt.Printf("Error decoding Spotify Top Tracks: " + err.Error())
	}

	t4 := time.Now()
	for _, event := range localSeatGeekEvents {
		for _, performer := range event.Performers {
			artistID := spotifyLayer.SearchAndFindSpotifyArtistID(performer)
			if artistID != "" {
				topTracks = append(topTracks, spotifyLayer.GetTopSpotifyArtistTrack(artistID))
			}
		}
	}

	fmt.Println("[Time benchmark] Top tracks " + time.Since(t4).String())
	topTracksJSON, err := json.Marshal(topTracks)
	if err != nil {
		fmt.Printf("Error Marshalling Spotify top tracks: " + err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(topTracksJSON)
	}
}

//POST
func BuildPlaylist(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	playlistName, ok := r.URL.Query()["name"]
	if !ok || len(playlistName[0]) < 1 {
		fmt.Printf("playlistName parameter missing from buildplaylist request.")
	}

	playlistDesc, ok := r.URL.Query()["desc"]
	if !ok || len(playlistDesc[0]) < 1 {
		fmt.Printf("playlistDesc parameter missing from buildplaylist request.")
	}

	var topTracks []spotify.FullTrack

	err := json.NewDecoder(r.Body).Decode(&topTracks)
	if err != nil {
		fmt.Printf("Error decoding Spotify Top Tracks: " + err.Error())
	}

	playlistID := spotifyLayer.GeneratePlayList(playlistName[0], playlistDesc[0])

	spotifyLayer.AddTracksToPlaylist(playlistID, topTracks)
}

//POST
func ConfigureCallbackURL(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	redirectURL, ok := r.URL.Query()["redirectURL"]
	if !ok || len(redirectURL[0]) < 1 {
		fmt.Printf("redirectURL parameter missing from configuration request.")
	}
}

//GET
//Callback is called from the Spotify authentication flow, and redirects to <Host>/#/callback
func Callback(w http.ResponseWriter, r *http.Request) {
	state, ok := r.URL.Query()["state"]
	if !ok || len(state[0]) < 1 {
		fmt.Printf("State parameter missing from callback.")
	}

	spotifyLayer.SetNewSpotifyClient(w, r, state[0])

	clientOrigin, err := redisLayer.GetKeyString(state[0])
	if err != nil {
		fmt.Printf("Error getting state/origin Redis key: " + err.Error())
	}

	redirectURL := clientOrigin + "/#/callback?state=" + state[0]

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

//GET
func Authenticate(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	state, ok := r.URL.Query()["state"]
	if !ok || len(state[0]) < 1 {
		fmt.Printf("State parameter missing from authenticate request.")
	}

	err := redisLayer.SetKeyString(state[0], r.Header.Get("Origin"))
	if err != nil {
		fmt.Printf("Error setting state/origin Redis key: " + err.Error())
	}

	authenticationUrl := spotifyLayer.ObtainAuthenticationURL(state[0])
	fmt.Fprint(w, authenticationUrl)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
