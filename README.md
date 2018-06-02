A simple discord bot written in [Go](https://golang.org/). It assigns a role to a user when they send a specific message.

<img src="screenshots/1.jpg" width="300"> <img src="screenshots/2.jpg" width="300">

## Install

[Install Go](https://golang.org/doc/install#install)
```sh
go get github.com/luaduck/rolebot
go install github.com/luaduck/rolebot
```
(Installs to `~/go/bin/`)

## Running
You will need a bot token from the [Discord developers site](https://discordapp.com/developers/applications/me)

Usage:
```
~$ rolebot
Usage of rolebot:
  -chan name
    	Channel name to use (default "roles")
  -char string
        Command character to previx all commands with (default "!")
  -t token
    	Bot token (required)
  -v	Verbose logging
```
Example:
```sh
rolebot -t YOURTOKENHERE
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
Usage: `!jointeam (team name)`

Command only works in a channel named `roles` (or other supplied with `-c`). It can be used by anyone.
Example:

`jointeam iBUYJEFFS`

