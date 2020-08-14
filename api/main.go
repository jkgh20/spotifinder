package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"otherside/api/seatgeekLayer"
	"otherside/api/spotifyLayer"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zmb3/spotify"
)

var applicationPort = "8081"
var callbackRedirectURL = "http://localhost:8080/#/callback"

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/authenticate", Authenticate)
	router.HandleFunc("/callback", Callback)
	router.HandleFunc("/localevents", LocalEvents)
	router.HandleFunc("/toptracks", TopTracks).Methods("POST")
	router.HandleFunc("/buildplaylist", BuildPlaylist).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+applicationPort, router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	//Don't serve a view; this will just be an API so keep frontend/backend separate
}

//GET
func LocalEvents(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	postCodes, ok := r.URL.Query()["postcodes"]
	if !ok || len(postCodes[0]) < 1 {
		fmt.Printf("postcodes parameter missing from localevents request.")
	}

	genres, ok := r.URL.Query()["genres"]
	if !ok || len(genres[0]) < 1 {
		fmt.Printf("genres parameter missing from localevents request.")
	}

	postCodeArray := QueryStringToArray(postCodes[0])
	genreArray := QueryStringToArray(genres[0])

	for _, val := range postCodeArray {
		fmt.Println(val)
	}

	for _, val := range genreArray {
		fmt.Println(val)
	}

	localSeatGeekEvents := seatgeekLayer.FindLocalEvents(postCodeArray, genreArray)

	localSeatGeekEventsJSON, err := json.Marshal(localSeatGeekEvents)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
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
		fmt.Fprintf(os.Stderr, err.Error())
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
		fmt.Fprintf(os.Stderr, err.Error())
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
		fmt.Fprintf(os.Stderr, err.Error())
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
func Callback(w http.ResponseWriter, r *http.Request) {
	state, ok := r.URL.Query()["state"]
	if !ok || len(state[0]) < 1 {
		fmt.Printf("State parameter missing from callback.")
	}

	spotifyLayer.SetNewSpotifyClient(w, r, state[0])

	redirectURL := callbackRedirectURL + "?state=" + state[0]

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

//GET
func Authenticate(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	state, ok := r.URL.Query()["state"]
	if !ok || len(state[0]) < 1 {
		fmt.Printf("State parameter missing from authenticate request.")
	}

	authenticationUrl := spotifyLayer.ObtainAuthenticationURL(state[0])

	fmt.Fprint(w, authenticationUrl)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
