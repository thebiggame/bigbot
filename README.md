# thebiggame/bigbot
A simple discord bot written in [Go](https://golang.org/).

_bigbot_ does many things. It is:
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
docker run -d --restart=always --name bigbot thebiggame/bigbot /app/main run --wan
```

Or local install:
```
~$ bigbot run
Usage:
  bigbot run [flags]

Flags:
  -h, --help   help for run
      --lan    Serve the LAN portion of the bot.
      --wan    Serve the WAN portion of the bot.

Global Flags:
      --config string      config file (default is ./bigbot.yaml)
      --log.level string   Log level (debug, info, warn, error, fatal) (default "info")
```
Example:
```sh
bigbot run --wan
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