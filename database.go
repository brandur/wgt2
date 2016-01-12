package wgt2

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Artist struct {
	// Genres that the artist belongs to. May be empty if unclassified.
	Genres []string `yaml:"genres"`

	// Canoical artist ID from Spotify.
	ID string `yaml:"id"`

	// Canonical artist name from Spotify.
	Name string `yaml:"name"`

	// Popularity of the artist.
	Popularity int `yaml:"popularity"`

	// Top tracks for the artist.
	TopTracks []Track `yaml:"tracks"`

	// URI for the artist.
	URI string `yaml:"uri"`

	// Name according to the WGT website
	WGTName string `yaml:"wgt_name"`
}

type ArtistCollection struct {
	// Maps WGTName to ID.
	ByWGTName map[string]string `yaml:"by_wgt_name"`

	// Map of artists keyed to a ID.
	Data map[string]*Artist `yaml:"data"`
}

func (c *ArtistCollection) AddArtist(artist *Artist) error {
	if artist.ID == "" {
		return fmt.Errorf("Artist ID cannot be empty")
	}

	if artist.URI == "" {
		return fmt.Errorf("Artist URI cannot be empty")
	}

	if artist.WGTName == "" {
		return fmt.Errorf("Artist WGTName cannot be empty")
	}

	c.Data[artist.URI] = artist
	c.ByWGTName[artist.WGTName] = artist.URI

	return nil
}

func (c *ArtistCollection) GetArtistByWGTName(wgtName string) *Artist {
	uri := c.ByWGTName[wgtName]
	if uri != "" {
		return c.Data[uri]
	}
	return nil
}

type Database struct {
	// Artists with data filled in from Spotify.
	Artists *ArtistCollection `yaml:"artists"`

	// Name of the file to which to save the database.
	Filename string

	// List of raw artist data from the WGT website.
	RawArtists *RawArtistCollection `yaml:"raw_artists"`
}

func NewDatabase() *Database {
	db := &Database{
		Artists: &ArtistCollection{
			ByWGTName: make(map[string]string),
			Data:      make(map[string]*Artist),
		},
		RawArtists: &RawArtistCollection{},
	}
	return db
}

func LoadDatabase(filename string) (*Database, error) {
	var db *Database

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("No database file; initializing new database")
		db = NewDatabase()
	} else {
		log.Printf("Reading: %v", filename)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal([]byte(data), &db)
		if err != nil {
			return nil, err
		}
	}

	db.Filename = filename
	return db, nil
}

func (db *Database) Save() error {
	if db.Filename == "" {
		return fmt.Errorf("Filename not specified")
	}

	log.Printf("Saving: %v", db.Filename)

	data, err := yaml.Marshal(&db)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(db.Filename, data, 0755)
	return err
}

type RawArtist struct {
	// Name according to the WGT website
	WGTName string `yaml:"wgt_name"`
}

type RawArtistCollection struct {
	// List of raw artists. Probably alphabetically ordered, but should be
	// treated as arbitrary.
	Data []*RawArtist `yaml:"data"`
}

func (c *RawArtistCollection) SetData(data []*RawArtist) error {
	for _, artist := range data {
		if artist.WGTName == "" {
			return fmt.Errorf("RawArtist WGTName cannot be empty")
		}
	}

	c.Data = data
	return nil
}

type Track struct {
	ID         string `yaml:"id"`
	Name       string `yaml:"name"`
	Popularity int    `yaml:"popularity"`
	URI        string `yaml:"uri"`
}
