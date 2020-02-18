package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin
	router *mux.Router

	configurationLock sync.RWMutex

	configuration *pluginConfig
	ServerConfig  *model.Config

	store *Store

	bundle *i18n.Bundle
}

const (
	// PluginID is a mattermost plugin id
	PluginID = "com.github.yiwkr.mattermost-plugin-janken"
)

// OnActivate registers the plugin command
func (p *Plugin) OnActivate() error {
	p.router = p.initAPI()
	p.store = NewStore(p.API)

	rand.Seed(time.Now().UnixNano())

	return nil
}

// OnDeactivate unregister the plugin command
func (p *Plugin) OnDeactivate() error {
	if err := p.API.UnregisterCommand("", p.getConfiguration().Trigger); err != nil {
		return errors.Wrap(err, "failed to deactivate command")
	}
	return nil
}

func getCommand(siteURL, trigger string) *model.Command {
	return &model.Command{
		Trigger:          trigger,
		DisplayName:      "Janken",
		Description:      "Playing janken",
		AutoComplete:     true,
		AutoCompleteDesc: "Create a janken",
		IconURL:          fmt.Sprintf("%s/plugins/%s/%s", siteURL, PluginID, iconFilename),
	}
}

// HasPermission checks if a given user has the permission
func (p *Plugin) HasPermission(game *game, userID string) (bool, error) {
	if userID == game.Creator {
		return true, nil
	}
	user, err := p.API.GetUser(userID)
	if err != nil {
		return false, err
	}
	if user.IsInRole(model.SYSTEM_ADMIN_ROLE_ID) {
		return true, nil
	}
	return false, nil
}
