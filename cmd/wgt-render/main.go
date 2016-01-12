package main

import (
	"bufio"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/brandur/wgt2"
	"github.com/yosssi/ace"
)

const (
	DBFilename = "./data.yaml"
	TargetDir  = "./public/"
)

func main() {
	// create an output directory (needed for both build and serve)
	err := os.MkdirAll(TargetDir, 0755)
	if err != nil {
		log.Fatal(err.Error())
	}

	template, err := ace.Load("index", "", &ace.Options{
		DynamicReload: true,
		FuncMap:       templateFuncMap(),
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := wgt2.LoadDatabase(DBFilename)
	if err != nil {
		log.Fatal(err.Error())
	}

	var artists []*wgt2.Artist
	for _, artist := range db.Artists.Data {
		artists = append(artists, artist)
	}

	file, err := os.Create(TargetDir + "index")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	data := map[string]interface{}{
		"artists": artists,
	}
	if err := template.Execute(writer, data); err != nil {
		log.Fatal(err.Error())
	}
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"JoinStrings": func(strs []string) string {
			return strings.Join(strs, ", ")
		},
	}
}
