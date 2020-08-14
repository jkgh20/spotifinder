package spotifyLayer

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/zmb3/spotify"
)

var spotifyAuth = spotify.NewAuthenticator(redirectURL, spotify.ScopePlaylistModifyPublic)
var spotifyClient spotify.Client

var applicationPort = "8081"
var baseURL = "http://localhost:" + applicationPort + "/"
var redirectURL = baseURL + "callback"

var SPOTIFY_ID = os.Getenv("SPOTIFY_ID")
var SPOTIFY_SECRET = os.Getenv("SPOTIFY_SECRET")
var userID = flag.String("user", "themobil9", "the Spotify user ID to look up")

func ObtainAuthenticationURL(state string) string {
	spotifyAuth.SetAuthInfo(SPOTIFY_ID, SPOTIFY_SECRET)

	return spotifyAuth.AuthURL(state)
}

func SetNewSpotifyClient(w http.ResponseWriter, r *http.Request) {
	token, err := spotifyAuth.Token("dummyState", r)

	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}

	spotifyClient = spotifyAuth.NewClient(token)

	http.Redirect(w, r, baseURL, http.StatusMovedPermanently)
}

func GetTopFourSpotifyArtistTracks(artistID spotify.ID) []spotify.FullTrack {
	topTracks, err := spotifyClient.GetArtistsTopTracks(artistID, "US")
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return nil
	} else {
		var topFourTracks []spotify.FullTrack

		for i, track := range topTracks {
			topFourTracks = append(topFourTracks, track)
			if i == 3 {
				break
			}
		}
		return topFourTracks
	}
}

func GeneratePlayList(playlistName string, description string) spotify.ID {
	flag.Parse()

	if *userID == "" {
		fmt.Fprintf(os.Stderr, "Error: missing user ID\n")
		flag.Usage()
		return ""
	}

	playList, err := spotifyClient.CreatePlaylistForUser(*userID, playlistName, description, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return ""
	} else {
		fmt.Printf("Created playlist %s for user %s", playlistName, *userID)
	}

	return playList.ID
}

func AddTracksToPlaylist(playlistID spotify.ID, tracksToAdd []spotify.FullTrack) {
	for _, track := range tracksToAdd {
		_, err := spotifyClient.AddTracksToPlaylist(playlistID, track.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}
	}
}

func SearchAndFindSpotifyArtistID(artistName string) spotify.ID {
	searchResults, err := spotifyClient.Search(artistName, spotify.SearchTypeArtist)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return ""
	} else {
		if len(searchResults.Artists.Artists) != 0 {
			return searchResults.Artists.Artists[0].ID
		}
	}

	return ""
}
