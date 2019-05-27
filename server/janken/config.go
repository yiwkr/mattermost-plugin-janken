package janken

import (
	"github.com/pkg/errors"
)

type configuration struct {
	Trigger string
}

// OnConfigurationChange loads the plugin configuration
func (p *Plugin) OnConfigurationChange() error {
	configuration := new(configuration)

	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	if old := p.getConfiguration(); old.Trigger != "" {
		if err := p.API.UnregisterCommand("", old.Trigger); err != nil {
			return errors.Wrap(err, "failed to unregister old command")
		}
	}

	if err := p.API.RegisterCommand(getCommand(configuration.Trigger)); err != nil {
		return errors.Wrap(err, "failed to register new command")
	}

	p.setConfiguration(configuration)
	p.ServerConfig = p.API.GetConfig()
	return nil
}

// getConfigurationはプラグインのコンフィグを取得する
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}
	return p.configuration
}

// getConfigurationはプラグインのコンフィグを設定する
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		panic("setConfiguration called with the existing configuration")
	}
	p.configuration = configuration
}
