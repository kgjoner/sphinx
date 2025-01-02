package common

import (
	"time"

	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/cornucopia/utils/httputil"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/i18n"
	"github.com/kgjoner/sphinx/internal/assets/style"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Mail struct {
	MailService hermes.MailService
	CacheRepo   cache.DAO
}

type MailInput struct {
	TemplateKey string
	Target      auth.Account
	Application auth.Application
	Links       []i18n.CustomLink
	Languages   []string
}

func (i Mail) Execute(input MailInput) (bool, error) {
	appName := config.Env.APP_NAME
	opt := []hermes.Options{}
	if input.Application.Brand.IsValidOnEmail {
		appName = input.Application.Name

		style, err := cache.RunWithCache[style.AppStyle](
			i.CacheRepo,
			7*24*time.Hour,
			httputil.Get[style.AppStyle],
		)(input.Application.Brand.StyleUrl)
		if err != nil {
			return false, err
		}

		opt = append(opt, hermes.Options{
			PrimaryColor:      style.Colors.PrimaryPure,
			PrimaryHoverColor: style.Colors.PrimaryLight,
			Header: struct {
				Logo            string "json:\"logo\""
				Title           string "json:\"title\""
				BackgroundColor string "json:\"backgroundColor\""
				Height          string "json:\"height\""
				Align           string "json:\"align\" validate:\"oneof=flex-end flex-start center space-between space-around\""
			}{
				Logo:            input.Application.Brand.LogoUrl,
				Title:           input.Application.Name,
				BackgroundColor: style.Colors.BackgroundLight,
			},
			Footer: struct {
				BackgroundColor string "json:\"backgroundColor\""
			}{
				BackgroundColor: style.Colors.BackgroundDark,
			},
		})
	}

	t := i18n.Resource(input.Languages).Mails[input.TemplateKey]
	t.ParseContent(i18n.ResourceParams{
		UserName:     input.Target.Name(),
		AppName:      appName,
		SupportEmail: config.Env.SUPPORT_EMAIL,
	})

	err := i.MailService.SendCustomEmail(input.Target.Email, t.Subject.Content, t.FormatBody(input.Links...), opt...)
	if err != nil {
		return false, err
	}

	return true, nil
}
