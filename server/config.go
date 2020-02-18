package main

import (
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

type pluginConfig struct {
	Trigger         string
	DefaultLanguage string
}

func (c *pluginConfig) GetDefaultLanguageTag() language.Tag {
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
	p.ServerConfig = p.API.GetConfig()
	c := new(pluginConfig)

	if err := p.API.LoadPluginConfiguration(c); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	if old := p.getConfiguration(); old.Trigger != "" {
		if err := p.API.UnregisterCommand("", old.Trigger); err != nil {
			return errors.Wrap(err, "failed to unregister old command")
		}
	}

	if err := p.API.RegisterCommand(getCommand(*p.ServerConfig.ServiceSettings.SiteURL, c.Trigger)); err != nil {
		return errors.Wrap(err, "failed to register new command")
	}

	p.setConfiguration(c)

	bundle, err := p.InitBundle()
	if err != nil {
		return err
	}
	p.bundle = bundle

	return nil
}

// getConfiguration はプラグインのコンフィグを取得する
func (p *Plugin) getConfiguration() *pluginConfig {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &pluginConfig{}
	}
	return p.configuration
}

// getConfiguration はプラグインのコンフィグを設定する
func (p *Plugin) setConfiguration(configuration *pluginConfig) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		panic("setConfiguration called with the existing configuration")
	}
	p.configuration = configuration
}
