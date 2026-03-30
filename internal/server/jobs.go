package server

import (
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/kgjoner/cornucopia/v3/httpclient"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/assets/style"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/pkg/mailer"
)

func (s Server) runJobs(ctx context.Context) {
	if config.Env.APP_STYLE_URL != "" {
		go runPeriodicTask(ctx, 24*60*60, func() {
			err := updateMailStyle(s.mailer)
			if err != nil {
				fmt.Printf("Failed to update Hermes style: %v\n", err)
			}
		})
	}

	// Key rotation job (if RS256 is enabled)
	// Check every 6 hours if rotation is needed based on actual key age.
	// This ensures rotation happens even when pods restart before the rotation period.
	if config.Env.JWT.ALGORITHM == string(auth.RS256) && config.Env.JWT.KEY_ROTATION_INTERVAL_HOURS > 0 {
		checkInterval := 6 * 3600 // 6 hours in seconds
		go runPeriodicTask(ctx, checkInterval, func() {
			shouldRotate, err := s.authIntGtw.ShouldRotate()
			if err != nil {
				fmt.Printf("Failed to check key rotation status: %v\n", err)
				return
			}

			if shouldRotate {
				err := s.authIntGtw.RotateKeys()
				if err != nil {
					fmt.Printf("Failed to rotate signing keys: %v\n", err)
				} else {
					fmt.Println("Signing keys rotated successfully")
				}
			} else {
				fmt.Println("Key rotation check: rotation not needed yet")
			}
		})
	}
}

// Retrieves app style accordingly to configuration and applies it to Hermes service.
func updateMailStyle(m *mailer.Mailer) error {
	logoURL := config.Env.APP_LOGO_URL
	if logoURL == "" {
		logoURL = config.Env.SCHEME + "://" + config.Env.HOST + config.BASE_PATH + "/assets/logo.svg"
	}

	if config.Env.APP_STYLE_URL == "" {
		// DEFAULT STYLE
		m.UpdateData(mailer.BaseData{
			PrimaryColor:           style.Root.Colors.PrimaryPure,
			PrimaryHoverColor:      style.Root.Colors.PrimaryDark,
			ContentBackgroundColor: style.Root.Colors.Neutral50,
			BodyBackgroundColor:    style.Root.Colors.BackgroundLight,
			LinkColor:              style.Root.Colors.PrimaryPure,
			DividerColor:           style.Root.Colors.Neutral200,
			Header: struct {
				Logo      string       "json:\"logo\""
				Title     string       "json:\"title\""
				Style     template.CSS "json:\"style\""
				LogoStyle template.CSS "json:\"logoStyle\""
			}{
				Logo:  logoURL,
				Title: config.Env.APP_NAME,
			},
			Footer: struct {
				Text  string       "json:\"text\""
				Style template.CSS "json:\"style\""
			}{
				Text:  fmt.Sprintf("© %v %v", time.Now().Year(), config.Env.APP_NAME),
				Style: template.CSS(fmt.Sprintf("background-color: %v; color: #fff", style.Root.Colors.BackgroundDark)),
			},
		}, mailer.InvariantData{
			AppName:              config.Env.APP_NAME,
			SupportEmail:         config.Env.SUPPORT_EMAIL,
			ClientBaseURL:        config.Env.CLIENT.BASE_URL,
			DataVerificationPath: config.Env.CLIENT.DATA_VERIFICATION,
			PasswordResetPath:    config.Env.CLIENT.PASSWORD_RESET,
			AliasName:            config.Env.APP_NAME,
		})
		return nil
	}

	// CUSTOM STYLE
	resp, err := httpclient.Get[httpserver.SuccessResponse[style.AppStyle]](config.Env.APP_STYLE_URL)
	if err != nil {
		return err
	}
	appStyle := resp.Data

	m.UpdateData(mailer.BaseData{
		PrimaryColor:           appStyle.Colors.PrimaryPure,
		PrimaryHoverColor:      appStyle.Colors.PrimaryLight,
		LinkColor:              appStyle.Colors.PrimaryPure,
		ContentBackgroundColor: appStyle.Colors.Neutral50,
		BodyBackgroundColor:    appStyle.Colors.BackgroundLight,
		DividerColor:           appStyle.Colors.Neutral200,
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
			Text:  fmt.Sprintf("© %v %v", time.Now().Year(), config.Env.APP_NAME),
			Style: appStyle.Mail.Footer,
		},
	}, mailer.InvariantData{
		AppName:              config.Env.APP_NAME,
		SupportEmail:         config.Env.SUPPORT_EMAIL,
		ClientBaseURL:        config.Env.CLIENT.BASE_URL,
		DataVerificationPath: config.Env.CLIENT.DATA_VERIFICATION,
		PasswordResetPath:    config.Env.CLIENT.PASSWORD_RESET,
		AliasName:            config.Env.APP_NAME,
	})

	return nil
}
