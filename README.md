# IsItStreamableBot

This discord bot utilizes the Discord, Spotify and Twitch API to detect posted Spotify-Links from a Discord channel and checks them against the Twitch DJ catalogue.

The found results are then written back as a reply to the posted message in Discord.

## Usage

You can either build the bot locally from source with:

```sh
go build .
```

or use the provided Dockerfile to build a ready to go Docker image:

```sh
docker buildx build -t isitstreamablebot:latest .
```

The Docker container can be started with the command:

```sh
docker run --rm -it --env-file=.env -p 8080:8080 isitstreamablebot:latest
```

### Remarks

The bot requires some environment variables in order to work properly.

#### Twitch client id

This is the weak point of this application.
There is no official documentation for the usage of the ["DJ music catalog"](https://www.twitch.tv/dj-signup#dj-music-catalog) so I just reverse engineered this part.
The TWITCH_CLIENT_ID is extracted from the gql-Request Headers (`client-id`) after a search request...


#### Spotify Login
Please visit the https://developer.spotify.com/dashboard to create yourself an application which will be used to connect to the Spotify API.

##### SPOTIFY_ID

The ID used for login.

##### SPOTIFY_SECRET

The secret used for the login.

##### SPOTIFY_REDIRECT_URI

Spotify needs a callback url where the user is redirected to.
Spotify does not allow HTTP in combination with `localhost`, but the loopback address (127.0.0.1) does still work.
All other FQDN need HTTPS, therefore the URI can define this here.

#### Discord

##### DISCORD_BOT_TOKEN

The login secret which is generated with the discord bot.

##### DISCORD_LISTENING_CHANNEL_IDS

In order to limit the bot to a specific channel, this variable can have multiple channel IDs seperated by comma.
