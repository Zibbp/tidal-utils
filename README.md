# Tidal-Utils

### Features

- Convert Spotify playlists to Tidal playlists
- Archive Spotify and Tidal playlists in JSON format
- More coming soon

### Setup

1. Grab a copy of the `docker-compose.yml` and run it for the first time `docker compose up`.
2. The container will provide a tidal oauth URL. Follow that and grant access to the application.
    1. Once authenticaed with Tidal the container will error due to missing Spotify credentials.
3. Create a [Spotify Application](https://developer.spotify.com/dashboard/applications) and fill in the spotify section in the config file.
4. In the `/data` folder create a file named `playlists.json` in the `/data/playlists` folder and enter Spotify playlists in the following format

```json
{
  "playlists": [
    "3cqRGdypS8yS0TvlwepZVR"
  ]
}
```

5. Bring the container up again and it should start converting the Spotify playlists to Tidal playlists.