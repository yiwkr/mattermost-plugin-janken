package main

import (
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/yiwkr/mattermost-plugin-janken/server/janken"
)

func main() {
	plugin.ClientMain(&janken.Plugin{})
}
