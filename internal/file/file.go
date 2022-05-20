package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zibbp/tidal-utils/internal/navidrome"
	"github.com/zibbp/tidal-utils/internal/tidal"

	spotifyPkg "github.com/zmb3/spotify/v2"
)

type UserSpotifyPlaylists struct {
	Playlists []string `json:"playlists"`
}

func Initialize() {
	// Create missing tracks directory if it doesn't exist
	missingTracks := "/data/missing_tracks"
	if _, err := os.Stat(missingTracks); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(missingTracks, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	spotify := "/data/spotify"
	if _, err := os.Stat(spotify); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(spotify, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	tidal := "/data/tidal"
	if _, err := os.Stat(tidal); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(tidal, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	navidrome := "/data/navidrome"
	if _, err := os.Stat(navidrome); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(navidrome, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	spotifyPlaylists := "/data/spotify/playlists"
	if _, err := os.Stat(spotifyPlaylists); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(spotifyPlaylists, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	tidalPlaylists := "/data/tidal/playlists"
	if _, err := os.Stat(tidalPlaylists); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(tidalPlaylists, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	navidromePlaylists := "/data/navidrome/playlists"
	if _, err := os.Stat(navidromePlaylists); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(navidromePlaylists, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
}

func GetUserSpotifyPlaylists() (UserSpotifyPlaylists, error) {
	jsonFile, err := os.Open("/data/spotify/playlists.json")
	if err != nil {
		log.Fatalf("Error opening user spotify playlists.json %w", err)
		return UserSpotifyPlaylists{}, err
	}
	defer jsonFile.Close()

	var userSpotifyPlaylists UserSpotifyPlaylists
	err = json.NewDecoder(jsonFile).Decode(&userSpotifyPlaylists)
	if err != nil {
		log.Fatalf("Error decoding user spotify playlists.json %w", err)
		return UserSpotifyPlaylists{}, err
	}

	return userSpotifyPlaylists, nil
}

func WriteJson(data []byte, path string, fileName string) error {
	err := ioutil.WriteFile(fmt.Sprintf("%s/%s", path, fileName), []byte(data), 0644)
	if err != nil {
		log.Fatalf("Error writing json %s %w", path, err)
		return err
	}
	return nil
}
func WriteUserSpotifyPlaylist(playlist spotifyPkg.FullPlaylist) error {
	data, err := json.Marshal(playlist)
	if err != nil {
		log.Fatalf("Error marshalling playlist %w", err)
		return err
	}
	err = WriteJson(data, "/data/spotify/playlists", fmt.Sprintf("%s.json", playlist.ID))
	if err != nil {
		log.Fatalf("Error writing playlist to file %w", err)
		return err
	}
	return nil
}

func WriteUserTidalPlaylist(playlist tidal.Playlist) error {
	data, err := json.Marshal(playlist)
	if err != nil {
		log.Fatalf("Error marshalling playlist %w", err)
		return err
	}
	err = WriteJson(data, "/data/tidal/playlists", fmt.Sprintf("%s.json", playlist.UUID))
	if err != nil {
		log.Fatalf("Error writing playlist to file %w", err)
		return err
	}
	return nil
}

func WriteNavidromePlaylist(playlist navidrome.Playlist) error {
	data, err := json.Marshal(playlist)
	if err != nil {
		log.Fatalf("Error marshalling playlist %w", err)
		return err
	}
	err = WriteJson(data, "/data/navidrome/playlists", fmt.Sprintf("%s.json", playlist.ID))
	if err != nil {
		log.Fatalf("Error writing playlist to file %w", err)
		return err
	}
	return nil
}
