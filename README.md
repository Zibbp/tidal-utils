# Tidal-Utils

### Features

- Convert Spotify playlists to Tidal playlists
- Archive Spotify and Tidal playlists in JSON format
- More coming soon

### Setup

1. Grab a copy of the `docker-compose.yml` and run it for the first time `docker compose up`.
2. The container will provide a tidal oauth URL. Follow that and grant access to the application.
    1. Once authenticaed with Tidal the container will error due to missing Spotify credentials.
3. Create a [Spotify Application](https://developer.spotify.com/dashboard/applications) and fill in the Spotify client ID and secret in `config.json` located under the `data` volume.
5. In the `data` volume, create a folder named `spotify` and within create a file named `playlists.json`. Inside playlists.json enter Spotify playlist IDs in the format below.

```json
{
  "playlists": [
    "3cqRGdypS8yS0TvlwepZVR"
  ]
}
```

5. Bring the container up and it should start converting Spotify playlists to Tidal playlists while archiving each platform's playlists in JSON format.
