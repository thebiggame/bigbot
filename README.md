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
docker run -d --restart=always --name bigbot thebiggame/bigbot /app/main -chan <channel> -token <token>
```

Or local install:
```
~$ bigbot
Usage of bigbot:
  -chan name
    	Channel name to use (default "roles")
  -char string
        Command character to prefix all commands with (default "!")
  -token token
    	Bot token (required)
  -user_maxroles int
        The maximum number of teams a User is allowed to join (default 5)
  -v	Verbose logging
```
Example:
```sh
bigbot -token YOURTOKENHERE
```
```
2018/03/18 18:09:18 Running on servers:
2018/03/18 18:09:18 	test (272429559406919681)
2018/03/18 18:09:18 channel name: roles
2018/03/18 18:09:18 Join URL:
2018/03/18 18:09:18 https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=(removed)
2018/03/18 18:09:18 Bot running as (removed). CTRL-C to exit.
```
paste the link into a web browser to add the bot to your discord server (you will need the Manage Server permission)

## Command Usage

### Register
Usage: `/team join (team name)`

Command only works in a channel named `roles` (or other supplied with `-c`). It can be used by anyone.
Example:

`/team join iBUYJEFFS`

