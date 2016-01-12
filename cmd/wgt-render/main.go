package main

import (
	"bufio"
	"html/template"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/brandur/wgt2"
	"github.com/yosssi/ace"
)

const (
	DBFilename      = "./data.yaml"
	TargetAssetsDir = "./public/assets/"
	TargetDir       = "./public/"
)

type artistSlice []*wgt2.Artist

func (s artistSlice) Len() int           { return len(s) }
func (s artistSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s artistSlice) Less(i, j int) bool { return s[i].Name < s[j].Name }

func main() {
	// create an output directory (the assets subdirectory here because its
	// parent will be created as a matter of course)
	err := os.MkdirAll(TargetAssetsDir, 0755)
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

	// Go doesn't exactly make sorting easy ...
	sort.Sort(artistSlice(artists))

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
