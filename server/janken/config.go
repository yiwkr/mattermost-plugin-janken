package janken

import (
	"github.com/pkg/errors"
)

type configuration struct {
	Trigger string
}

/*
OnConfigurationChange Hookをオーバーライド．
Mattermostサーバーで設定が変更された場合に実行される．
*/
func (p *Plugin) OnConfigurationChange() error {
	configuration := new(configuration)

	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.ServerConfig = p.API.GetConfig()
	p.setConfiguration(configuration)
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
