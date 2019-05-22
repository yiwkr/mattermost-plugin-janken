package janken

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func (p *Plugin) initBundle() (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	exPath, err := os.Executable()
	serverDistDir := filepath.Dir(exPath)
	serverDir := filepath.Dir(serverDistDir)
	pluginDir := filepath.Dir(serverDir)
	assetsDir := filepath.Join(pluginDir, "assets")

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
	}
	return bundle, nil
}

func (p *Plugin) Localize(l *i18n.Localizer, messageId string, templateData map[string]interface{}) string {
	m := l.MustLocalize(&i18n.LocalizeConfig{
		MessageID: messageId,
		TemplateData: templateData,
	})
	return m
}

func (p *Plugin) GetLocalizer(tag string) *i18n.Localizer {
	return i18n.NewLocalizer(p.bundle, tag)
}
