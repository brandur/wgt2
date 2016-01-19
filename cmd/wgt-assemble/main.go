package main

import (
	"log"

	"github.com/brandur/wgt2"
	"github.com/joeshaw/envdecode"
	"github.com/zmb3/spotify"
)

var (
	DBFilename = "./data.yaml"
	PlaylistName = "WGT 2016"
)

type Conf struct {
	ClientID     string `env:"CLIENT_ID,required"`
	ClientSecret string `env:"CLIENT_SECRET,required"`
	RefreshToken string `env:"REFRESH_TOKEN,required"`
}

func main() {
	var conf Conf
	err := envdecode.Decode(&conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	client := wgt2.GetSpotifyClient(conf.ClientID, conf.ClientSecret, conf.RefreshToken)

	db, err := wgt2.LoadDatabase(DBFilename)
	if err != nil {
		log.Fatal(err.Error())
	}

	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err.Error())
	}

	playlist, err := getPlaylist(client, user)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, artist := range db.Artists.Data {
		var trackIDs []spotify.ID
		for _, track := range artist.TopTracks {
			trackIDs = append(trackIDs, spotify.ID(track.ID))
		}

		log.Printf("Adding tracks for: %v", artist.Name)
		_, err := client.AddTracksToPlaylist(user.ID, playlist.ID, trackIDs...)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func getPlaylist(client *spotify.Client, user *spotify.PrivateUser) (*spotify.SimplePlaylist, error) {
	page, err := client.CurrentUsersPlaylists()
	if err != nil {
		return nil, err
	}

	for _, playlist := range page.Playlists {
		if playlist.Name == PlaylistName {
			log.Printf("Found playlist: %v", PlaylistName)
			return &playlist, nil
		}
	}

	// otherwise create a new playlist for the user
	log.Printf("Creating playlist: %v", PlaylistName)

	playlist, err := client.CreatePlaylistForUser(user.ID, PlaylistName, true)
	return &playlist.SimplePlaylist, err
}
