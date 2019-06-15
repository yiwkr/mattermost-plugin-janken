package janken

import (
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

type configuration struct {
	Trigger         string
	DefaultLanguage string
}

func (c *configuration) GetDefaultLanguageTag() language.Tag {
	defaultLanguage := language.English
	if c == nil {
		return defaultLanguage
	}

	t, err := language.Parse(c.DefaultLanguage)
	if err != nil {
		return defaultLanguage
	}
	return t
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

	if bundle, err := p.InitBundle(); err != nil {
		return err
	} else {
		p.bundle = bundle
	}

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
