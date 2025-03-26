package style

import (
	"embed"
	"encoding/json"
	"html/template"
)

type AppStyle struct {
	Colors struct {
		PrimaryPure     string `json:"primaryPure" validate:"required"`
		PrimaryLight    string `json:"primaryLight"`
		PrimaryLightest string `json:"primaryLightest"`
		PrimaryDark     string `json:"primaryDark"`
		PrimaryDarkest  string `json:"primaryDarkest"`

		SecondaryPure     string `json:"secondaryPure" validate:"required"`
		SecondaryLight    string `json:"secondaryLight"`
		SecondaryLightest string `json:"secondaryLightest"`
		SecondaryDark     string `json:"secondaryDark"`
		SecondaryDarkest  string `json:"secondaryDarkest"`

		BackgroundLight string `json:"backgroundLight"`
		BackgroundDark  string `json:"backgroundDark"`

		PositivePure  string `json:"positivePure"`
		PositiveLight string `json:"positiveLight"`
		PositiveDark  string `json:"positiveDark"`

		DangerPure  string `json:"dangerPure"`
		DangerLight string `json:"dangerLight"`
		DangerDark  string `json:"dangerDark"`

		WarningPure  string `json:"warningPure"`
		WarningLight string `json:"warningLight"`
		WarningDark  string `json:"warningDark"`

		InfoPure  string `json:"infoPure"`
		InfoLight string `json:"infoLight"`
		InfoDark  string `json:"infoDark"`

		NeutralPure string `json:"neutralPure"`
		Neutral50   string `json:"neutral50"`
		Neutral100  string `json:"neutral100"`
		Neutral200  string `json:"neutral200"`
		Neutral300  string `json:"neutral300"`
		Neutral400  string `json:"neutral400"`
		Neutral500  string `json:"neutral500"`
		Neutral600  string `json:"neutral600"`
		Neutral700  string `json:"neutral700"`
		Neutral800  string `json:"neutral800"`
		Neutral900  string `json:"neutral900"`
	} `json:"colors"`

	Fonts struct {
		HeadlineLarge string `json:"headlineLarge"`
		Headline      string `json:"headline"`
		HeadlineSmall string `json:"headlineSmall"`

		SubtitleLarge string `json:"subtitleLarge"`
		Subtitle      string `json:"subtitle"`
		SubtitleSmall string `json:"subtitleSmall"`

		BodyLarge string `json:"bodyLarge"`
		Body      string `json:"body"`
		BodySmall string `json:"bodySmall"`

		CaptionLarge string `json:"captionLarge"`
		Caption      string `json:"caption"`
		CaptionSmall string `json:"captionSmall"`
	} `json:"fonts"`

	Outline string `json:"outline"`

	Mail struct {
		Header template.CSS `json:"header"`
		Logo   template.CSS `json:"logo"`
		Footer template.CSS `json:"footer"`
	}
}

//go:embed *.json
var files embed.FS

var Root AppStyle

func init() {
	rawFile, _ := files.ReadFile("root.json")
	json.Unmarshal(rawFile, &Root)
}
