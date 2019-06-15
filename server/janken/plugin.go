package janken

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

type Plugin struct {
	plugin.MattermostPlugin
	router *mux.Router

	configurationLock sync.RWMutex

	configuration *configuration
	ServerConfig  *model.Config

	store *Store

	bundle *i18n.Bundle
}

const (
	PluginId = "com.github.yiwkr.mattermost-plugin-janken"
)

// OnAcrivate registers the plugin command
func (p *Plugin) OnActivate() error {
	p.router = p.InitAPI()
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

func getCommand(trigger string) *model.Command {
	return &model.Command{
		Trigger:          trigger,
		DisplayName:      "Janken",
		Description:      "Playing janken",
		AutoComplete:     true,
		AutoCompleteDesc: "Create a janken",
	}
}

/*
HasPermission checks if a given user has the permission
*/
func (p *Plugin) HasPermission(game *JankenGame, userId string) (bool, error) {
	if userId == game.Creator {
		return true, nil
	}
	user, err := p.API.GetUser(userId)
	if err != nil {
		return false, err
	}
	if user.IsInRole(model.SYSTEM_ADMIN_ROLE_ID) {
		return true, nil
	}
	return false, nil
}
