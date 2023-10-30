# Go TwitchAlerts
A go package to allow the user to detect when a streamer is live and trigger a custom event. Requests are currently rate limited at 80ms since that that is the fastest the Twitch api allows. Each streamer has a thirty-second cooldown between checks, delay between check runs is customisable in the config file.<br><br>If you find any bugs or have a feature request, please report them on GitHub and any improvements or additions are welcome through pull requests.
## Features
- Customisable Events
- Configurable
- Easy Error Handling
## Setup

### Install
Run the command - 
```bash
go get github.com/concelaredev/gotwitchalerts
```

In your go.mod
```go.mod
module example

go 1.21

require github.com/concelaredev/gotwitchalerts v1.0
```
Import into file
```go
package main

import "github.com/concelaredev/gotwitchalerts"

func main() {
}
```

### Configuration

The first run will create a config file, this should contain the client-id, token, delay and list of streamers to monitor.

To get your OAuth token for twitch go to https://dev.twitch.tv/console create an application and use the Client ID & Secret with the command below
```http request
curl -X POST 'https://id.twitch.tv/oauth2/token' \
-H 'Content-Type: application/x-www-form-urlencoded' \
-d 'client_id=<your client id goes here>&client_secret=<your client secret goes here>&grant_type=client_credentials'
```
#### Example Config
```json
{
  "user_id":"your_user_id",
  "token":"your_token",
  "streamers":["streamer1", "streamer2"],
  "delay":80
}
```
## Dependencies
- encoding/json
- io
- log
- net/http
- os
- slices
- strings
- time
## Contributors
- [@ConcelareDev](https://www.github.com/ConcelareDev)