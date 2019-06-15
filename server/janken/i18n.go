package janken

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func getAssetsDir() (string, error) {
	exPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	serverDistDir := filepath.Dir(exPath)
	serverDir := filepath.Dir(serverDistDir)
	pluginDir := filepath.Dir(serverDir)
	assetsDir := filepath.Join(pluginDir, "assets")
	return assetsDir, nil
}

func (p *Plugin) InitBundle() (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(p.configuration.GetDefaultLanguageTag())
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	assetsDir, err := getAssetsDir()
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
	}
	return bundle, nil
}

func (p *Plugin) GetLocalizer(tag string) *i18n.Localizer {
	return i18n.NewLocalizer(p.bundle, tag)
}

func Localize(l *i18n.Localizer, defaultMessage *i18n.Message, templateData map[string]interface{}) string {
	m := l.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: defaultMessage,
		TemplateData:   templateData,
	})
	return m
}
