package util

import (
	"encoding/json"
	"fmt"

	"github.com/flytam/filenamify"
	"github.com/zibbp/tidal-utils/internal/file"
	"github.com/zibbp/tidal-utils/internal/navidrome"
	"github.com/zibbp/tidal-utils/internal/tidal"
	spotifyPkg "github.com/zmb3/spotify/v2"
)

type MissingTracks struct {
	PlaylistName string         `json:"playlist_name"`
	Tracks       []MissingTrack `json:"tracks"`
}

type MissingTrack struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	ISRC   string `json:"isrc"`
}

func SpotifyPlaylistOnTidal(a string, list []tidal.Playlist) (bool, int) {
	for i, b := range list {
		if b.Title == a {
			return true, i
		}
	}
	return false, 0
}

func ProcessMissingTracks(tracks []spotifyPkg.PlaylistTrack, playlistName string) error {

	var missingTracks []MissingTrack
	for _, track := range tracks {
		misTrack := MissingTrack{
			ID:     string(track.Track.ID),
			Name:   track.Track.Name,
			Artist: track.Track.Artists[0].Name,
			Album:  track.Track.Album.Name,
			ISRC:   track.Track.ExternalIDs["isrc"],
		}
		missingTracks = append(missingTracks, misTrack)
	}
	missing := MissingTracks{
		PlaylistName: playlistName,
		Tracks:       missingTracks,
	}

	json, err := json.Marshal(missing)
	if err != nil {
		return err
	}
	fileName, err := filenamify.Filenamify(fmt.Sprintf("%s.json", playlistName), filenamify.Options{Replacement: "-"})
	if err != nil {
		return err
	}
	err = file.WriteJson(json, "/data/missing_tracks", fileName)
	if err != nil {
		return err
	}
	return nil

}

func TidalPlaylistToNavidromePlaylist(tidalPlaylist tidal.Playlist) error {
	var navidromePlaylist navidrome.Playlist
	for _, track := range tidalPlaylist.Tracks {
		navidromeTrack := navidrome.Track{
			Title:  track.Title,
			Artist: track.Artist.Name,
			Album:  track.Album.Title,
		}
		navidromePlaylist.Tracks = append(navidromePlaylist.Tracks, navidromeTrack)
	}
	navidromePlaylist.Name = tidalPlaylist.Title
	navidromePlaylist.ID = tidalPlaylist.UUID
	err := file.WriteNavidromePlaylist(navidromePlaylist)
	if err != nil {
		return err
	}
	return nil
}
