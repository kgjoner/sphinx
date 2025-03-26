package common

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	// "time"

	"github.com/kgjoner/cornucopia/helpers/presenter"
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

		resp, err := cache.RunWithCache[presenter.Success[style.AppStyle]](
			i.CacheRepo,
			7*24*time.Hour,
			httputil.Get[presenter.Success[style.AppStyle]],
		)(input.Application.Brand.StyleUrl)
		if err != nil {
			return false, err
		}
		appStyle := resp.Data

		opt = append(opt, hermes.Options{
			PrimaryColor:      appStyle.Colors.PrimaryPure,
			PrimaryHoverColor: appStyle.Colors.PrimaryLight,
			Header: struct {
				Logo      string       "json:\"logo\""
				Title     string       "json:\"title\""
				Style     template.CSS "json:\"style\""
				LogoStyle template.CSS "json:\"logoStyle\""
			}{
				Logo:      input.Application.Brand.LogoUrl,
				Title:     input.Application.Name,
				Style:     appStyle.Mail.Header,
				LogoStyle: appStyle.Mail.Logo,
			},
			Footer: struct {
				Text  string       "json:\"text\""
				Style template.CSS "json:\"style\""
			}{
				Style: appStyle.Mail.Footer,
			},
		})
	}

	
	for i, link := range input.Links {
		if strings.HasPrefix(link.Link, "/") {
			if input.Application.Brand.IsValidOnEmail {
				path := ""
				parts := strings.Split(link.Link, "?")
				if len(parts) > 1 {
					path = fmt.Sprintf("/%s?path=%s&%s", input.Application.Id, parts[0], parts[1])
				} else {
					path = fmt.Sprintf("/%s?path=%s", input.Application.Id, parts[0])
				}
				input.Links[i].Link = config.Env.CLIENT.BASE_URL + path
			} else {
				input.Links[i].Link = config.Env.CLIENT.BASE_URL + link.Link
			}
		}
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
