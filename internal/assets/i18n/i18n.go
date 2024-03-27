package i18n

import (
	"embed"
	"encoding/json"

	"github.com/kgjoner/cornucopia/helpers/i18n"
	"github.com/kgjoner/sphinx/internal/config"
)

type resource struct {
	Mail struct {
		Welcome struct {
			Subject string
			Title   string
			Button  string
			P1      string
			P2      string
		}
	}
}

type resourceMap map[i18n.Language]resource

//go:embed *.json
var files embed.FS

var AcceptedLanguages = []i18n.Language{i18n.LanguageValues.PT_BR, i18n.LanguageValues.EN_US}
var Resources resourceMap

func init() {
	for _, lng := range AcceptedLanguages {
		rawFile, _ := files.ReadFile(string(lng) + ".json")
		var res resource
		json.Unmarshal(rawFile, &res)

		Resources[lng] = res
	}
}

func Resource(lngs []string) resource {
	for _, lng := range lngs {
		res, exists := Resources[i18n.Language(lng)]
		if exists {
			return res
		}
	}

	return Resources[i18n.Language(config.Environment.FALLBACK_LANGUAGE)]
}
