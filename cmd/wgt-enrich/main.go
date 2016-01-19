package main

import (
	"log"

	"github.com/brandur/wgt2"
	"github.com/joeshaw/envdecode"
	"github.com/zmb3/spotify"
)

const (
	DBFilename = "./data.yaml"
)

var (
	// These unfortunately don't match correctly with the string given on the
	// WGT website. Override them manually so as not to present incorrect data.
	// A mapping to an empty string just means "don't bother with the search".
	ManualOverrides = map[string]string{
		// Incorrectly maps to John Legend.
		"Legend": "",

		// Incorrectly maps to NAO.
		"Näo": "",
	}
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

	for _, rawArtist := range db.RawArtists.Data {
		dbArtist := db.Artists.GetArtistByWGTName(rawArtist.WGTName)
		if dbArtist != nil {
			log.Printf("Have artist already: %v", dbArtist.Name)
			continue
		}

		searchName := rawArtist.WGTName
		if val, ok := ManualOverrides[searchName]; ok {
			if val == "" {
				log.Printf("Skipping '%v' due to override", searchName)
				continue
			}

			log.Printf("Using manual override '%v' for '%v'", val, searchName)
			searchName = val
		}

		res, err := client.Search(searchName, spotify.SearchTypeArtist)
		if err != nil {
			log.Fatal(err.Error())
		}

		if len(res.Artists.Artists) < 1 {
			log.Printf("Artist not found: %v", rawArtist.WGTName)
			continue
		}

		artist := res.Artists.Artists[0]

		dbArtist = &wgt2.Artist{
			Genres:     artist.Genres,
			ID:         string(artist.ID),
			Name:       artist.Name,
			Popularity: artist.Popularity,
			URI:        string(artist.URI),
			WGTName:    rawArtist.WGTName,
		}

		log.Printf("Found artist: %v (from: %v; popularity: %v/100)",
			artist.Name, rawArtist.WGTName, artist.Popularity)

		tracks, err := client.GetArtistsTopTracks(artist.ID, "US")
		if err != nil {
			log.Fatal(err.Error())
		}

		for _, track := range tracks {
			dbTrack := wgt2.Track{
				ID:         string(track.ID),
				Name:       track.Name,
				Popularity: track.Popularity,
				URI:        string(track.URI),
			}
			dbArtist.TopTracks = append(dbArtist.TopTracks, dbTrack)
		}

		db.Artists.AddArtist(dbArtist)
	}

	err = db.Save()
	if err != nil {
		log.Fatal(err.Error())
	}
}
