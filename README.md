# thebiggame/BIGbot
<img align="right" alt="BIGbot logo" src="assets/logo.png" width="200"/>

A simple discord bot written in [Go](https://golang.org/).

_BIGbot_ does many things. It is:
* responsible for the creation and assignment of special "team" roles to those who request them 
  * (this helps massively when you want to mention an entire team quickly and easily (or give them their own chat channel!))
* a helper for tBG Crew actions, such as announcements / AV switching

## Install

[Install Go](https://golang.org/doc/install#install)
```sh
go get github.com/thebiggame/bigbot
go install github.com/thebiggame/bigbot
```
(Installs to `~/go/bin/`)

## Running

You will need a bot token from the [Discord developers site](https://discordapp.com/developers/applications/me)

Usage:

Via docker:
```
docker run -d --restart=always --name bigbot thebiggame/bigbot /app/bigbot run --serve-wan
```

Or local install:
```
~$ bigbot run
Usage: bigbot run --serve-lan --serve-wan --discord.token=STRING [flags]

Run BIGbot.

Flags:
  -h, --help                      Show context-sensitive help.
      --config=CONFIG-FLAG        Location of config ($BIGBOT_CONFIG)
  -l, --log-level="info"          Set the logging level
                                  (debug|info|warn|error|fatal)
                                  ($BIGBOT_LOG_LEVEL)

  -t, --discord.token=STRING      Discord bot token ($BIGBOT_DISCORD_TOKEN)
      --discord.guild-id=""       Discord guild ID to monitor
                                  ($BIGBOT_DISCORD_GUILD)
      --av.obs.hostname=""        OBS Host ($BIGBOT_AV_OBS_HOST)
      --av.obs.password=""        OBS password ($BIGBOT_AV_OBS_PASSWORD)
      --teams.max-user-teams=5    Maximum number of teams a User can join
                                  ($BIGBOT_MAX_USER_ROLES)
      --remove-commands           Remove commands on shutdown
                                  ($BIGBOT_COMMANDS_REMOVE)

serve
  --serve-lan    Serve the LAN portion of the bot.
  --serve-wan    Serve the WAN portion of the bot.
```
Example:
```sh
bigbot run --serve-wan -t=[discord-token]
```
```
2018/03/18 18:09:18 Running on servers:
2018/03/18 18:09:18 	test (removed)
2018/03/18 18:09:18 Join URL: https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=(removed)
2018/03/18 18:09:18 Bot running as (removed). CTRL-C to exit.
```
paste the link into a web browser to add the bot to your discord server (you will need the Manage Server permission)

## Command Usage

### Register
Usage: `/team create (team name)`

This command takes any string team name and creates a team role with that name.

Example:
`/team join iBUYJEFFS`

### Join
Usage: `/team join (team)`

This command requires a team role (enforced in Discord).

### Leave
Usage: `/team leave (team)`

This command requires a team role (enforced in Discord).

You may only Leave teams that you are a member of (and not other roles).