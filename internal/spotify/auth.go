package spotify

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/spf13/viper"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/zmb3/spotify/v2"
)

var (
	ch    = make(chan *spotify.Client)
	state = "tida-utils"
)

type Service struct {
	client *spotify.Client
}

func NewService() *Service {

	// Check if Spotify app id and secret are set
	if viper.GetString("spotify.client_id") == "" || viper.GetString("spotify.client_secret") == "" {
		log.Fatal("Spotify client ID and secret not set.")
	}

	// Check if Spotify access token and refresh token are set
	if viper.GetString("spotify.access_token") == "" || viper.GetString("spotify.refresh_token") == "" {
		log.Warn("Spotify access token and refresh token not set.")
		client, err := auth()
		if err != nil {
			log.Fatal("Spotify auth failed: %v", err)
		}
		return &Service{client: client}
	}

	// Use Spotify refresh token to get a new access token and create client
	tok := &oauth2.Token{
		AccessToken:  viper.GetString("spotify.access_token"),
		RefreshToken: viper.GetString("spotify.refresh_token"),
		Expiry:       viper.GetTime("spotify.expiry"),
		TokenType:    viper.GetString("spotify.token_type"),
	}
	spotClientID := viper.GetString("spotify.client_id")
	spotClientSecret := viper.GetString("spotify.client_secret")
	redirectURI := viper.GetString("spotify.redirect_uri")
	auth := spotifyauth.New(spotifyauth.WithClientID(spotClientID), spotifyauth.WithClientSecret(spotClientSecret), spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))

	client := spotify.New(auth.Client(context.Background(), tok))

	newTok, _ := client.Token()
	viper.Set("spotify.access_token", newTok.AccessToken)
	viper.Set("spotify.expiry", newTok.Expiry)
	viper.Set("spotify.token_type", newTok.TokenType)
	viper.WriteConfig()

	return &Service{client: client}

}

func auth() (*spotify.Client, error) {
	spotClientID := viper.GetString("spotify.client_id")
	spotClientSecret := viper.GetString("spotify.client_secret")
	redirectURI := viper.GetString("spotify.redirect_uri")
	auth := spotifyauth.New(spotifyauth.WithClientID(spotClientID), spotifyauth.WithClientSecret(spotClientSecret), spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":28542", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Spotify - Logged in as:", user.ID)
	return client, nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	spotClientID := viper.GetString("spotify.client_id")
	spotClientSecret := viper.GetString("spotify.client_secret")
	redirectURI := viper.GetString("spotify.redirect_uri")
	auth := spotifyauth.New(spotifyauth.WithClientID(spotClientID), spotifyauth.WithClientSecret(spotClientSecret), spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))

	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// Save token to config
	viper.Set("spotify.access_token", tok.AccessToken)
	viper.Set("spotify.refresh_token", tok.RefreshToken)
	viper.Set("spotify.expiry", tok.Expiry)
	viper.Set("spotify.token_type", tok.TokenType)
	viper.WriteConfig()

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}
