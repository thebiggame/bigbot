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
docker run -d --restart=always --name bigbot thebiggame/bigbot /app/bigbot run
```

Local installations work identically; just call the binary directly.

### run
```
~$ bigbot run
Usage: bigbot run --discord.token=SECRET-STRING [flags]

Run BIGbot (the main Discord bot).

Flags:
  -h, --help                                   Show context-sensitive help.
      --config=CONFIG-FLAG                     Location of config ($BIGBOT_CONFIG)
  -l, --log-level="INFO"                       Set the logging level (TRACE|DEBUG|INFO|WARN|ERROR|FATAL) ($BIGBOT_LOG_LEVEL)

      --bridge.enabled                         Enable the BIGbot -> Bridge Server ($BIGBOT_BRIDGE_ENABLED)
      --bridge.address="localhost:8080"        Listen address and port ($BIGBOT_BRIDGE_LISTEN)
      --bridge.key=SECRET-STRING               BIGbot authentication key ($BIGBOT_BRIDGE_KEY)
  -t, --discord.token=SECRET-STRING            Discord bot token ($BIGBOT_DISCORD_TOKEN)
      --discord.guild-id=""                    Discord guild ID to monitor ($BIGBOT_DISCORD_GUILD)
      --discord.announcements.channel-id=""    Channel ID ($BIGBOT_DISCORD_ANNOUNCEMENTS_CHANNEL)
      --discord.permissions.crew-role=""       If a user is a member of this role ID, treat them as Crew ($BIGBOT_DISCORD_PERMISSIONS_ROLE_CREW).
      --discord.shoutbox.channel-id=""         Channel ID ($BIGBOT_DISCORD_SHOUTBOX_CHANNEL)
      --av.nodecg.bundle-name="thebiggame"     NodeCG bundle name ($BIGBOT_AV_NODECG_BUNDLE)
      --teams.max-user-teams=5                 Maximum number of teams a User can join ($BIGBOT_MAX_USER_ROLES)
      --remove-commands                        Remove commands on shutdown ($BIGBOT_COMMANDS_REMOVE)
```
Example:
```sh
bigbot run
```
```
2018/03/18 18:09:18 Running on servers:
2018/03/18 18:09:18 	test (removed)
2018/03/18 18:09:18 Join URL: https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=(removed)
2018/03/18 18:09:18 Bot running as (removed). CTRL-C to exit.
```
Paste the link into a web browser to add the bot to your discord server (you will need the Manage Server permission)

### Bridge
```
Usage: bigbot bridge --key=SECRET-STRING [flags]

Run BIGbridge (the event client).

Flags:
  -h, --help                                   Show context-sensitive help.
      --config=CONFIG-FLAG                     Location of config ($BIGBOT_CONFIG)
  -l, --log-level="INFO"                       Set the logging level (TRACE|DEBUG|INFO|WARN|ERROR|FATAL) ($BIGBOT_LOG_LEVEL)

      --ws-address="ws://localhost:8080/ws"    BIGbot address and port ($BIGBRIDGE_ADDR)
      --key=SECRET-STRING                      BIGbot authentication key ($BIGBRIDGE_KEY)
      --av.obs.hostname=""                     OBS Host ($BIGBRIDGE_AV_OBS_HOST)
      --av.obs.password=""                     OBS password ($BIGBRIDGE_AV_OBS_PASSWORD)
      --av.nodecg.hostname=""                  NodeCG Host ($BIGBRIDGE_AV_NODECG_HOST)
      --av.nodecg.bundle-name="thebiggame"     NodeCG bundle name ($BIGBRIDGE_AV_NODECG_BUNDLE)
      --av.nodecg.authentication-key=""        Authentication key ($BIGBRIDGE_AV_NODECG_AUTHKEY)
```
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