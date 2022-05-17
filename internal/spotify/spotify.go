package spotify

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zibbp/tidal-utils/internal/file"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type Service struct {
	client *spotify.Client
}

func NewService() *Service {
	log.Info("Logging into Spotify")
	ctx := context.Background()
	configClientId := viper.Get("spotify.client_id").(string)
	configClientSecret := viper.Get("spotify.client_secret").(string)
	config := &clientcredentials.Config{
		ClientID:     configClientId,
		ClientSecret: configClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	return &Service{
		client: client,
	}
}

func (s *Service) GetPlaylists(playlistIds file.UserSpotifyPlaylists) ([]spotify.FullPlaylist, error) {
	var playlists []spotify.FullPlaylist
	for _, playlistId := range playlistIds.Playlists {
		playlist, err := s.GetPlaylist(spotify.ID(playlistId))
		if err != nil {
			log.Errorf("Couldn't get Spotify playlist: %v", err)
		}
		playlists = append(playlists, *playlist)
	}

	return playlists, nil
}

func (s *Service) GetPlaylist(id spotify.ID) (*spotify.FullPlaylist, error) {
	playlist, err := s.client.GetPlaylist(context.Background(), id)
	if err != nil {
		log.Errorf("Couldn't get playlist: %v", err)
		return nil, err
	}
	return playlist, nil
}

func (s *Service) GetPlaylistTracks(id spotify.ID) ([]spotify.PlaylistTrack, error) {
	tracks, err := s.client.GetPlaylistTracks(context.Background(), id)
	if err != nil {
		log.Errorf("Couldn't get playlist tracks: %v", err)
		return nil, err
	}

	allPlaylistTracks := []spotify.PlaylistTrack{}

	for page := 1; ; page++ {
		{
			// Apend tracks
			for _, track := range tracks.Tracks {
				allPlaylistTracks = append(allPlaylistTracks, track)
			}
			err = s.client.NextPage(context.Background(), tracks)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Errorf("Couldn't get playlist tracks: %v", err)
			}
		}
	}
	return allPlaylistTracks, nil
}
