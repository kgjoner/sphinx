package server

import (
	"context"
	"fmt"
	"html/template"

	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/cornucopia/v2/utils/httputil"
	"github.com/kgjoner/hermes/pkg/hermes"
	"github.com/kgjoner/sphinx/internal/assets/style"
	"github.com/kgjoner/sphinx/internal/config"
)

func (s Server) runJobs(ctx context.Context) {
	if config.Env.APP_STYLE_URL != "" {
		go runPeriodicTask(ctx, 24*60*60, func() {
			err := updateHermesStyle(s.mailSvc)
			if err != nil {
				fmt.Printf("Failed to update Hermes style: %v\n", err)
			}
		})
	}
}

// Retrieves app style accordingly to configuration and applies it to Hermes service.
func updateHermesStyle(hms *hermes.Service) error {
	logoURL := config.Env.APP_LOGO_URL
	if logoURL == "" {
		logoURL = config.Env.HOST + "/root/logo.svg"
	}

	if config.Env.APP_STYLE_URL == "" {
		// DEFAULT STYLE
		hms.UpdateDefaultOptions(hermes.Options{
			PrimaryColor:      style.Root.Colors.PrimaryPure,
			PrimaryHoverColor: style.Root.Colors.PrimaryDark,
			Header: struct {
				Logo      string       "json:\"logo\""
				Title     string       "json:\"title\""
				Style     template.CSS "json:\"style\""
				LogoStyle template.CSS "json:\"logoStyle\""
			}{
				Logo:  logoURL,
				Title: config.Env.APP_NAME,
				Style: template.CSS(fmt.Sprintf("background-color: %v;", style.Root.Colors.BackgroundLight)),
			},
			Footer: struct {
				Text  string       "json:\"text\""
				Style template.CSS "json:\"style\""
			}{
				Style: template.CSS(fmt.Sprintf("background-color: %v;", style.Root.Colors.BackgroundDark)),
			},
			Alias: struct {
				Address htypes.Email "json:\"address\""
				Name    string       "json:\"name\""
			}{
				Name: config.Env.APP_NAME,
			},
		})
		return nil
	}

	// CUSTOM STYLE
	resp, err := httputil.Get[presenter.Success[style.AppStyle]](config.Env.APP_STYLE_URL)
	if err != nil {
		return err
	}
	appStyle := resp.Data

	hms.UpdateDefaultOptions(hermes.Options{
		PrimaryColor:      appStyle.Colors.PrimaryPure,
		PrimaryHoverColor: appStyle.Colors.PrimaryLight,
		Header: struct {
			Logo      string       "json:\"logo\""
			Title     string       "json:\"title\""
			Style     template.CSS "json:\"style\""
			LogoStyle template.CSS "json:\"logoStyle\""
		}{
			Logo:      logoURL,
			Title:     config.Env.APP_NAME,
			Style:     appStyle.Mail.Header,
			LogoStyle: appStyle.Mail.Logo,
		},
		Footer: struct {
			Text  string       "json:\"text\""
			Style template.CSS "json:\"style\""
		}{
			Style: appStyle.Mail.Footer,
		},
		Alias: struct {
			Address htypes.Email "json:\"address\""
			Name    string       "json:\"name\""
		}{
			Name: config.Env.APP_NAME,
		},
	})

	return nil
}
