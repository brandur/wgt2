package main

//
// Usage caveats:
//
//   * Not even remotely safe for concurrent runs. Two programs running
//     simultaneously can easily create two playlists with the same name or add
//     tracks twice.
//
//   * Tracks are only added and never removed.
//

import (
	"log"

	"github.com/brandur/wgt2"
	"github.com/joeshaw/envdecode"
	"github.com/zmb3/spotify"
)

var (
	DBFilename          = "./data.yaml"
	PlaylistName        = "WGT 2016"
	PlaylistNameObscure = "WGT 2016 - Obscure"
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

	var trackIDs []spotify.ID

	//
	// Main playlist
	//

	trackIDs = nil
	for _, artist := range db.Artists.Data {
		for _, track := range artist.TopTracks {
			trackIDs = append(trackIDs, spotify.ID(track.ID))
		}
	}

	err = updatePlaylist(client, user, PlaylistName, trackIDs)
	if err != nil {
		log.Fatal(err.Error())
	}

	//
	// Popular playlist
	//

	trackIDs = nil
	for _, artist := range db.Artists.Data {
		// an arbitrary threshold
		if artist.Popularity < 20 {
			continue
		}

		for _, track := range artist.TopTracks {
			trackIDs = append(trackIDs, spotify.ID(track.ID))
		}
	}

	err = updatePlaylist(client, user, PlaylistNamePopular, trackIDs)
	if err != nil {
		log.Fatal(err.Error())
	}

	//
	// Obscure playlist
	//

	trackIDs = nil
	for _, artist := range db.Artists.Data {
		// an arbitrary threshold
		if artist.Popularity >= 20 {
			continue
		}

		for _, track := range artist.TopTracks {
			trackIDs = append(trackIDs, spotify.ID(track.ID))
		}
	}

	err = updatePlaylist(client, user, PlaylistNameObscure, trackIDs)
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

		log.Printf("Playlist %v: of %v track(s)", playlist.Name, len(tracksPage.Tracks))

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

	log.Printf("Playlist %v: %v track(s)", playlist.Name, len(allTracks))

	var trackIDsToAdd []spotify.ID
	for _, id := range trackIDs {
		if _, ok := allTracks[id]; !ok {
			trackIDsToAdd = append(trackIDs, id)
		}
	}

	log.Printf("Playlist %v: Need to add %v track(s)", playlist.Name,
		len(trackIDsToAdd))

	for i := 0; i < len(trackIDsToAdd); i += 100 {
		bound := i + 100
		if bound > len(trackIDsToAdd) {
			bound = len(trackIDsToAdd)
		}

		_, err := client.AddTracksToPlaylist(user.ID, playlist.ID, trackIDsToAdd[i:bound]...)
		if err != nil {
			return err
		}
		log.Printf("Playlist %v: added page of %v track(s)", playlist.Name,
			len(trackIDsToAdd[i:bound]))
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

func updatePlaylist(client *spotify.Client, user *spotify.PrivateUser, playlistName string, trackIDs []spotify.ID) error {
	playlist, err := getPlaylist(client, user, playlistName)
	if err != nil {
		return err
	}

	err = addTracksToPlaylist(client, user, playlist, trackIDs)
	if err != nil {
		return err
	}

	return nil
}
