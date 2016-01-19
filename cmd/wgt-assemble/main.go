package main

import (
	"log"

	"github.com/brandur/wgt2"
	"github.com/joeshaw/envdecode"
	"github.com/zmb3/spotify"
)

var (
	DBFilename          = "./data.yaml"
	PlaylistName        = "WGT 2016"
	PlaylistNamePopular = "WGT 2016 - Popular"
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

	playlist, err := getPlaylist(client, user, PlaylistName)
	if err != nil {
		log.Fatal(err.Error())
	}

	var trackIDs []spotify.ID
	for _, artist := range db.Artists.Data {
		for _, track := range artist.TopTracks {
			trackIDs = append(trackIDs, spotify.ID(track.ID))
		}
	}

	err = addTracksToPlaylist(client, user, playlist, trackIDs)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func addTracksToPlaylist(client *spotify.Client, user *spotify.PrivateUser, playlist *spotify.SimplePlaylist, trackIDs []spotify.ID) error {
	allTracks := make(map[spotify.ID]spotify.FullTrack)
	limit := 100
	offset := 0

	var options spotify.Options
	options.Limit = &limit
	options.Offset = &offset

	for {
		tracksPage, err := client.GetPlaylistTracksOpt(user.ID, playlist.ID, &options, "")
		if err != nil {
			return err
		}

		log.Printf("Fetched playlist page of %v track(s)", len(tracksPage.Tracks))

		for _, playlistTrack := range tracksPage.Tracks {
			track := playlistTrack.Track
			allTracks[track.ID] = track
		}

		// The Spotify API always returns full pages unless it has none left to
		// return.
		if len(tracksPage.Tracks) < 100 {
			break
		}

		offset = offset + len(tracksPage.Tracks)
	}

	log.Printf("Current playlist has %v track(s)", len(allTracks))

	var trackIDsToAdd []spotify.ID
	for _, id := range trackIDs {
		if _, ok := allTracks[id]; !ok {
			trackIDsToAdd = append(trackIDs, id)
		}
	}

	log.Printf("Need to add %v track(s) to playlist", len(trackIDsToAdd))

	for i := 0; i < len(trackIDsToAdd); i += 100 {
		bound := i + 100
		if bound > len(trackIDsToAdd) {
			bound = len(trackIDsToAdd)
		}

		_, err := client.AddTracksToPlaylist(user.ID, playlist.ID, trackIDsToAdd[i:bound]...)
		if err != nil {
			return err
		}
		log.Printf("Added page of %v track(s) to playlist", len(trackIDsToAdd[i:bound]))
	}

	return nil
}

func getPlaylist(client *spotify.Client, user *spotify.PrivateUser, playlistName string) (*spotify.SimplePlaylist, error) {
	page, err := client.CurrentUsersPlaylists()
	if err != nil {
		return nil, err
	}

	for _, playlist := range page.Playlists {
		if playlist.Name == playlistName {
			log.Printf("Found playlist: %v", playlistName)
			return &playlist, nil
		}
	}

	// otherwise create a new playlist for the user
	log.Printf("Creating playlist: %v", playlistName)

	playlist, err := client.CreatePlaylistForUser(user.ID, playlistName, true)
	return &playlist.SimplePlaylist, err
}
