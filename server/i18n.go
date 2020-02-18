package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (p *Plugin) getAssetsDir() (string, error) {
	pluginDir, err := p.API.GetBundlePath()
	if err != nil {
		return "", err
	}
	p.API.LogDebug("pluginDir: " + pluginDir)
	assetsDir := filepath.Join(pluginDir, "assets")
	p.API.LogDebug("assetsDir: " + assetsDir)
	return assetsDir, nil
}

// InitBundle initialize i18n.Bundle
func (p *Plugin) InitBundle() (*i18n.Bundle, error) {
	t := p.configuration.GetDefaultLanguageTag()
	p.API.LogDebug("DefaultLanguage: " + t.String())
	bundle := i18n.NewBundle(t)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	assetsDir, err := p.getAssetsDir()
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(assetsDir)
	if err != nil {
		return nil, err
	}

	r := regexp.MustCompile(`^active\..*\.toml$`)
	for _, file := range files {
		if !r.MatchString(file.Name()) {
			continue
		}

		if _, err = bundle.LoadMessageFile(filepath.Join(assetsDir, file.Name())); err != nil {
			return nil, err
		}
		p.API.LogDebug(fmt.Sprintf("loaded language file: %s", file.Name()))
	}
	return bundle, nil
}

func (p *Plugin) getLocalizer(tag string) *i18n.Localizer {
	return i18n.NewLocalizer(p.bundle, tag)
}

// Localize localize message
func Localize(l *i18n.Localizer, defaultMessage *i18n.Message, templateData map[string]interface{}) string {
	m := l.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: defaultMessage,
		TemplateData:   templateData,
	})
	return m
}
