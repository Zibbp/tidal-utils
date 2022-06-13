package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zibbp/tidal-utils/internal/config"
	"github.com/zibbp/tidal-utils/internal/file"
	"github.com/zibbp/tidal-utils/internal/spotify"
	"github.com/zibbp/tidal-utils/internal/tidal"
	"github.com/zibbp/tidal-utils/internal/util"
	spotifyPkg "github.com/zmb3/spotify/v2"
)

type PlatformPlaylists struct {
	Playlists []PlatformPlaylist
}

type PlatformPlaylist struct {
	Spotify spotifyPkg.FullPlaylist
	Tidal   tidal.Playlist
}

func main() {

	config.NewConfig()

	// Tidal auth service
	tidalService := tidal.NewService()

	// Spotify auth service
	spotifyService := spotify.NewService()

	if viper.Get("debug").(bool) == true {
		log.SetLevel(log.DebugLevel)
	}

	file.Initialize()

	spotifyToTidal(spotifyService, tidalService)

}

func spotifyToTidal(spotifyService *spotify.Service, tidalSerivce *tidal.Service) {

	manual := viper.Get("manual").(bool)

	var spotPlaylists file.UserSpotifyPlaylists

	// Check if manual mode is set to true.
	// Manual mode: User manually adds Spotify playlist IDs to playlists.json
	// User mode: User's own/liked Spotify playlists are auto added
	if manual == true {
		log.Info("Manual mode enabled.")
		// Get local user playlists
		spotifyPlaylistsJson, err := file.GetUserSpotifyPlaylists()
		if err != nil {
			log.Errorf("Error reading local Spotify playlists: %w", err)
		}
		spotPlaylists = spotifyPlaylistsJson
	} else {
		log.Info("User mode enabled.")
		var userSpotifyPlaylists file.UserSpotifyPlaylists
		usrPlaylists, err := spotifyService.GetUsersPlaylists()
		if err != nil {
			log.Errorf("Couldn't get user playlists: %v", err)
			return
		}
		for _, playlist := range usrPlaylists {
			userSpotifyPlaylists.Playlists = append(userSpotifyPlaylists.Playlists, string(playlist.ID))
		}
		spotPlaylists = userSpotifyPlaylists
	}

	// Fetch Spotify playlist information
	spotifyPlaylists, err := spotifyService.GetPlaylists(spotPlaylists)
	if err != nil {
		log.Errorf("Error fetching Spotify playlists: %w", err)
	}
	// Fetch user tidal playlists
	tidalPlaylists, err := tidalSerivce.GetUserPlaylists(tidalSerivce.UserID)
	if err != nil {
		log.Errorf("Error fetching Tidal playlists: %w", err)
	}
	var platformPlaylists PlatformPlaylists
	// Check if Spotify playlist exists on Tidal
	for _, playlist := range spotifyPlaylists {
		// Write spotify playlist locally
		log.Debugf("Writing Spotify playlist %s to local file", playlist.Name)
		err = file.WriteUserSpotifyPlaylist(playlist)
		if err != nil {
			log.Errorf("Error writing Spotify playlist %s to local file: %w", playlist.Name, err)
		}
		var platformPlaylist PlatformPlaylist
		// Add playlist to platformPlaylists
		platformPlaylist.Spotify = playlist
		// Check if playlist exists on Tidal
		playlistExists, i := util.SpotifyPlaylistOnTidal(playlist.Name, tidalPlaylists.Items)
		if !playlistExists {
			// Playlist does not exist on Tidal - create it
			createdTidalPlaylist, err := tidalSerivce.CreatePlaylist(playlist.Name, playlist.Description)
			if err != nil {
				log.Errorf("Error creating playlist %s on Tidal: %w", playlist.Name, err)
			}
			log.Debugf("Created playlist %s on Tidal - ID: %v", createdTidalPlaylist.Title, createdTidalPlaylist.UUID)
			// Add playlist to platformPlaylists
			platformPlaylist.Tidal = createdTidalPlaylist
		} else {
			// Add playlist to platformPlaylists
			platformPlaylist.Tidal = tidalPlaylists.Items[i]
		}

		platformPlaylists.Playlists = append(platformPlaylists.Playlists, platformPlaylist)
	}

	// Match spotify tracks to deezer tracks
	for _, playlist := range platformPlaylists.Playlists {
		log.Infof("Processing playlist %s", playlist.Spotify.Name)
		// Get playlist tracks
		spotifyTracks, err := spotifyService.GetPlaylistTracks(playlist.Spotify.ID)
		if err != nil {
			log.Errorf("Error getting playlist tracks: %w", err)
		}
		var missingTracks []*spotifyPkg.PlaylistTrack
		// Loop through Spotify playlist tracks and attempt to search on Tidal
		for _, track := range spotifyTracks {
			searchSpotifyToTidal(track, playlist.Tidal, &missingTracks, tidalSerivce)
		}
		// Missing tracks
		if len(missingTracks) > 0 {
			var newMissingTracks []spotifyPkg.PlaylistTrack
			for _, track := range missingTracks {
				newMissingTracks = append(newMissingTracks, *track)
			}
			util.ProcessMissingTracks(newMissingTracks, playlist.Tidal.Title)
		}
		log.Infof("Finished processing playlists %s", playlist.Spotify.Name)
		tidalPlaylistTracks, err := tidalSerivce.GetPlaylistTracks(playlist.Tidal.UUID)
		if err != nil {
			log.Errorf("Error getting playlist tracks: %w", err)
		}
		playlist.Tidal.Tracks = tidalPlaylistTracks.Items
		// Write tidal playlist locally
		log.Debugf("Writing Tidal playlist %s to local file", playlist.Tidal.Title)
		err = file.WriteUserTidalPlaylist(playlist.Tidal)
		if err != nil {
			log.Errorf("Error writing Tidal playlist %s to local file: %w", playlist.Tidal.Title, err)
		}
		// Write tidal playlist to Navidrome playlist
		err = util.TidalPlaylistToNavidromePlaylist(playlist.Tidal)
		if err != nil {
			log.Errorf("Error writing Tidal playlist %s to Navidrome playlist: %w", playlist.Tidal.Title, err)
		}
	}

}

func searchSpotifyToTidal(track spotifyPkg.PlaylistTrack, tidalPlaylist tidal.Playlist, missingTracks *[]*spotifyPkg.PlaylistTrack, tidalSerivce *tidal.Service) {
	// Search for track on Tidal
	search, err := tidalSerivce.SearchTracks(fmt.Sprintf("%s %s", track.Track.Name, track.Track.Artists[0].Name), 1)
	if err != nil {
		log.Error(err)
	}
	if search.TopHit.Value.Title != "" {
		// Track found on Tidal
		// Add to playlist
		err := tidalSerivce.AddTrackToPlaylist(tidalPlaylist.UUID, search.TopHit.Value.ID)
		if err != nil {
			log.Errorf("Error adding track %s to playlist %s: %w", search.TopHit.Value.Title, tidalPlaylist.Title, err)
		}
	} else {
		// Track not found - attempt to search again
		// Remove any [ or ( from the track name
		splitBrackets := strings.Split(track.Track.Name, "[")[0]
		newTrackTitle := strings.Split(splitBrackets, "(")[0]

		// Check if there is a second artist
		var newArtist string
		if len(track.Track.Artists) > 1 {
			newArtist = track.Track.Artists[1].Name
		} else {
			newArtist = track.Track.Artists[0].Name
		}

		// Search by new track title and album
		search, err = tidalSerivce.SearchTracks(fmt.Sprintf("%s %s", newTrackTitle, newArtist), 1)

		if err != nil {
			log.Error(err)
		}
		if search.TopHit.Value.Title != "" {
			// Track found on Tidal
			// Add to playlist
			err := tidalSerivce.AddTrackToPlaylist(tidalPlaylist.UUID, search.TopHit.Value.ID)
			if err != nil {
				log.Errorf("Error adding track %s to playlist %s: %w", search.TopHit.Value.Title, tidalPlaylist.Title, err)
			}
		} else {
			// Track is missing :(
			*missingTracks = append(*missingTracks, &track)
		}
	}

}
