package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"

	flag "github.com/ogier/pflag"
)

type Conf struct {
	ClientID     string
	ClientSecret string
	Port         string
}

type CallbackResponse struct {
	code  string
	state string
}

func exitWithError(message string) {
	os.Exit(2)
}

func main() {
	conf := &Conf{}

	flag.StringVarP(&conf.ClientID, "client-id", "i", "", "OAuth client ID")
	flag.StringVarP(&conf.ClientSecret, "client-secret", "s", "", "OAuth client secret")
	flag.StringVarP(&conf.Port, "port", "p", "8080", "Port to listen on")
	flag.Parse()

	if conf.ClientID == "" || conf.ClientSecret == "" {
		fmt.Printf("usage: procure --client-id <id> --client-secret <id>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	respChan := make(chan *CallbackResponse)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("See console.\n"))

		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse incoming form\n")
			return
		}

		response := &CallbackResponse{
			code:  r.Form.Get("code"),
			state: r.Form.Get("state"),
		}

		if response.code == "" {
			fmt.Fprintf(os.Stderr, "Server did not provide 'code' parameter\n")
			return
		}

		if response.state == "" {
			fmt.Fprintf(os.Stderr, "Server did not provide 'state' parameter\n")
			return
		}

		respChan <- response
	})

	go func() {
		fmt.Printf("Listening for callback on :%v\n", conf.Port)
		err := http.ListenAndServe("localhost:"+conf.Port, nil)
		if err != nil {
			panic(err)
		}
	}()

	spotifyEndpoint := oauth2.Endpoint{
		AuthURL:  "https://accounts.spotify.com/authorize",
		TokenURL: "https://accounts.spotify.com/api/token",
	}

	oauthConf := &oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  "http://localhost:" + conf.Port,
		Scopes:       []string{"playlist-modify-public", "user-read-email"},
		Endpoint:     spotifyEndpoint,
	}

	// should use a real state here
	url := oauthConf.AuthCodeURL("state")
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)

	response := <-respChan
	// should validate state here

	token, err := oauthConf.Exchange(oauth2.NoContext, response.code)
	if err != nil {
		exitWithError(err.Error())
	}

	fmt.Printf("\n")
	fmt.Printf("Success!\n")
	fmt.Printf("Access token: %v\n", token.AccessToken)
	fmt.Printf("Refresh token: %v\n", token.RefreshToken)
	fmt.Printf("Access token expires in: %v\n", token.Expiry.Sub(time.Now()))
}
