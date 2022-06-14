package spotify

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/zibbp/tidal-utils/internal/file"
	"github.com/zmb3/spotify/v2"
)

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

func (s *Service) GetUsersPlaylists() ([]spotify.SimplePlaylist, error) {

	playlists, err := s.client.CurrentUsersPlaylists(context.Background())
	if err != nil {
		log.Errorf("Couldn't get user playlists: %v", err)
		return nil, err
	}

	allUsersPlaylists := []spotify.SimplePlaylist{}

	for page := 1; ; page++ {
		{
			// Apend playlists
			for _, playlist := range playlists.Playlists {
				allUsersPlaylists = append(allUsersPlaylists, playlist)
			}
			err = s.client.NextPage(context.Background(), playlists)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Errorf("Couldn't get users playlists: %v", err)
			}
		}
	}
	return allUsersPlaylists, nil
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
