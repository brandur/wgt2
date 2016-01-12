package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/brandur/wgt2"
)

const (
	ArtistsURL = "http://www.wave-gotik-treffen.de/english/bands.php"
	DBFilename = "./data.yaml"
)

func main() {
	body, err := fetchBody()
	if err != nil {
		log.Fatal(err.Error())
	}

	var dbArtists []*wgt2.RawArtist
	artists := extractArtists(body)
	for _, artist := range artists {
		log.Printf("Got raw artist: %v", artist)
		dbArtists = append(dbArtists, &wgt2.RawArtist{WGTName: artist})
	}

	db, err := wgt2.LoadDatabase(DBFilename)
	if err != nil {
		log.Fatal(err.Error())
	}

	db.RawArtists.SetData(dbArtists)
	db.Save()
}

func extractArtists(haystack string) []string {
	regex := regexp.MustCompile(`<span class="firstchar">(\p{L})</span>([\p{L}. ]+) \([A-Z]+\)`)
	rawArtists := regex.FindAllString(haystack, -1)

	var artists []string
	for _, rawArtist := range rawArtists {
		groups := regex.FindStringSubmatch(rawArtist)

		// index 0 holds entire match, 1 and 2 are the capture groups
		artists = append(artists, groups[1]+groups[2])
	}

	return artists
}

func fetchBody() (string, error) {
	resp, err := http.Get(ArtistsURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}
