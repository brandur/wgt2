package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

const (
	Port      = 5002
	TargetDir = "./public/"
)

func main() {
	port := Port
	if os.Getenv("PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Fatal(err.Error())
		}

		if p < 1 {
			if err != nil {
				log.Fatal("PORT must be >= 1")
			}
		}
		port = p
	}

	fmt.Printf("Serving '%v' on port %v\n", path.Clean(TargetDir), port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), http.FileServer(http.Dir(TargetDir)))
	if err != nil {
		log.Fatal(err.Error())
	}
}
