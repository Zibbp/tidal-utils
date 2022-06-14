# Tidal-Utils

### Features

- Convert Spotify playlists to Tidal playlists
- Archive Spotify and Tidal playlists in JSON format
- Save Tidal playlists in JSON format for [navidrome-utils](https://github.com/Zibbp/navidrome-utils)

### Getting Started

1. Grab a copy of the `docker-compose.yml` and run it to generate the config file `docker compose up`.
2. Follow the supplied link to authorize your Tidal account. The program will now exit asking for a Spotify cliend ID and secret.
3. Create a [Spotify Application](https://developer.spotify.com/dashboard/applications) adding a valid redirect URI of `http://HOST:28542/callback` and fill in the Spotify client ID and secret in `./data/config.json`.
4. Modify the spotify redirect URI in `./data/config.json` to match the one in your Spotify application, ensuring the IP/host is correct.
4. Run `docker-compose up` again and follow the supplied link to authorize your Spotify account.
5. Your created and followed Spotify playlists should begin to be processed.

### Notes

#### Manual Mode

By default the program will process all created or liked Spotify playlists by the authorized account. If you would like to switch to manual mode allowing you to enter specific IDs in a JSON file, you can do so by setting the `manual` flag in the config file to `true`. Once you have set this flag, create a `playlists.json` file under `./data/spotify` with the following format.

```json
{
  "playlists": [
    "3cqRGdypS8yS0TvlwepZVR"
  ]
}
```

#### Spotify Token

The Spotify access token lasts for an hour but the refresh is unlimited. At appplication run the access token is refreshed meaning it is unlikely you will have to re-authorize your Spotify account.

#### Directory Structure

Within the data folder there are four sub-folders.

- `missing_tracks`: Contains tracks that are missing for each Playlist during that Spotify to Tidal conversion.
- `spotify`: Contains the Spotify playlists in JSON format.
- `tidal`: Contains the Tidal playlists in JSON format.
- `navidrome`: Contains Tidal playlists in a format that [navidrome-utils](https://github.com/Zibbp/navidrome-utils) can read. This is for creating Navidrome playlists from Tidal playlists if you have the songs downloaded.

