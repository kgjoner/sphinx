package sphinx

import (
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/httputil"
	"github.com/kgjoner/cornucopia/utils/structop"
)

type LoginOutput struct {
	UserID       uuid.UUID `json:"userID"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int       `json:"expiresIn"`
}

type ExternalAuthBody struct {
	ProviderName    string `validate:"required"`
	Params          map[string]string
	Body            map[string]string
	ConsentRelation bool
	ConsentCreation bool
	Email           string
}

func (s Service) ExternalAuth(authorization string, body ExternalAuthBody, languages ...string) (*LoginOutput, error) {
	mapBody := structop.New(body).Map()

	var respData presenter.Success[LoginOutput]
	_, err := s.httpApi.Post("/auth/external", mapBody, &httputil.Options{
		Headers: map[string]string{
			"Authorization":   authorization,
			"Accept-Language": strings.Join(languages, ","),
		},
	})(&respData)

	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}
