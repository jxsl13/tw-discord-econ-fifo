# tw-discord-econ-fifo

Small fifo bridge for discord.
Allows to execute econ/rcon commands via discord.


# building

```shell
go build .

# or

go install .

# or

go install github.com/jxsl13/tw-discord-econ-fifo@latest
```

# running

Check out the `Makefile` for docker instructions.

```shell
$ ./tw-discord-econ-fifo --help
Environment variables:
  DISCORD_TOKEN      discord bot token from the discord developer website
  DISCORD_CHANNEL    channel id that the bot works on
  ECON_ADDRESS       ip:port
  ECON_PASSWORD      econ server password

Usage:
  tw-discord-econ-fifo [flags]

Flags:
  -c, --config string            .env config file path (or via env variable CONFIG)
      --discord-channel string   channel id that the bot works on
      --discord-token string     discord bot token from the discord developer website
      --econ-address string      ip:port
      --econ-password string     econ server password
  -h, --help                     help for tw-discord-econ-fifo
```


example `.env` config file:

```dotenv
DISCORD_TOKEN="MT...."
DISCORD_CHANNEL="1196208959123456789"
ECON_PASSWORD="secret_password"

# use this hostname in order to access the host system's ports from within the docker container
# otherwise you may preferrably use localhost:9303 (or any other port)
ECON_ADDRESS="host.docker.internal:9303"
```

