package main

import (
	"log"

	"github.com/brandur/wgt2"
	"github.com/joeshaw/envdecode"
	"github.com/zmb3/spotify"
)

var (
	DBFilename   = "./data.yaml"
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

	allTracks := make(map[spotify.ID]spotify.FullTrack)
	limit := 100
	offset := 0

	var options spotify.Options
	options.Limit = &limit
	options.Offset = &offset

	for {
		tracksPage, err := client.GetPlaylistTracksOpt(user.ID, playlist.ID, &options, "")
		if err != nil {
			log.Fatal(err.Error())
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

	// track IDs to add to the playlist
	var trackIDs []spotify.ID

	for _, artist := range db.Artists.Data {
		for _, track := range artist.TopTracks {
			if _, ok := allTracks[spotify.ID(track.ID)]; !ok {
				trackIDs = append(trackIDs, spotify.ID(track.ID))
			}
		}
	}

	log.Printf("Need to add %v track(s) to playlist", len(trackIDs))

	for i := 0; i < len(trackIDs); i += 100 {
		bound := i + 100
		if bound > len(trackIDs) {
			bound = len(trackIDs)
		}

		_, err := client.AddTracksToPlaylist(user.ID, playlist.ID, trackIDs[i:bound]...)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Added page of %v track(s) to playlist", len(trackIDs[i:bound]))
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
