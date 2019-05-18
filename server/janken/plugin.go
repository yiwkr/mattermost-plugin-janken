package janken

import (
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin
	router *mux.Router

	configurationLock sync.RWMutex

	configuration *configuration
	ServerConfig  *model.Config

	store *Store
}

const (
	PluginId = "com.github.yiwkr.mattermost-plugin-janken"
	trigger = "janken"
)

// OnAcrivate registers the plugin command
func (p *Plugin) OnActivate() error {
	p.router = p.InitAPI()
	p.store, _ = NewStore(p.API)
	return p.API.RegisterCommand(&model.Command{
		Trigger: trigger,
		DisplayName: "Janken",
		Description: "Playing janken",
		AutoComplete: true,
		AutoCompleteDesc: "Create a janken",
	})
	return nil
}

// OnDeactivate unregister the plugin command
func (p *Plugin) OnDeactivate() error {
	err := p.API.UnregisterCommand("", trigger)
	return err
}

/*
HasPermission checks if a given user has the permission to end or delete a given poll
ほぼ下記のコピペ
https://github.com/matterpoll/matterpoll/blob/v1.1.0/server/plugin/plugin.go#L109
*/
func (p *Plugin) HasPermission(game *JankenGame, userId string) (bool, *model.AppError) {
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
