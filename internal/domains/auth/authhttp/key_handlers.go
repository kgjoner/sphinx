package authhttp

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
)

// keyHandlers mounts admin key management routes
func (g *gateway) keyHandlers(r chi.Router) {
	r.Use(g.AuthenticateApp)
	r.Post("/rotate", g.rotateKeys)
	r.Get("/status", g.getKeysStatus)
}

// GetJWKS godoc
//
//	@Summary		Get JSON Web Key Set (JWKS)
//	@Description	Retrieve the public keys used for verifying JWTs issued by the system
//	@Router			/.well-known/jwks.json [get]
//	@Tags			Well-known
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	JWKSResponse
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		403	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
func (g *gateway) getJWKS(w http.ResponseWriter, r *http.Request) {
	i := authcase.ListActiveSigningKeys{
		AuthRepo: g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection()),
	}

	output, err := i.ExecutePublic()
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	jwksPresenter(output, g.KeyProvisioner, w)
}

// RotateKeys godoc
//
//	@Summary		Rotate signing keys
//	@Description	Manually trigger immediate JWT signing key rotation (admin only)
//	@Router			/admin/keys/rotate [post]
//	@Tags			Admin
//	@Security		BasicApp
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		403	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
func (g *gateway) rotateKeys(w http.ResponseWriter, r *http.Request) {
	c := httpserver.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input authcase.RotateKeysInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	_, err = g.PGPool.WithTx(r.Context(), nil, func(tx *sql.Tx) (any, error) {
		i := authcase.RotateKeys{
			AuthRepo:       g.AuthFactory.NewDAO(r.Context(), tx),
			KeyProvisioner: g.KeyProvisioner,
			Encryptor:      g.Encryptor,
		}

		return i.Execute(input)
	})
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(nil, w, r, http.StatusNoContent)
}

// GetKeysStatus godoc
//
//	@Summary		Get signing key status
//	@Description	Get information about active signing keys and rotation schedule (admin only)
//	@Router			/admin/keys/status [get]
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Security		BasicApp
//	@Success		200	{object}	KeyStatusResponse
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		403	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
func (g *gateway) getKeysStatus(w http.ResponseWriter, r *http.Request) {
	c := httpserver.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input authcase.ListActiveSigningKeysInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	i := authcase.ListActiveSigningKeys{
		AuthRepo: g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection()),
	}

	keys, err := i.Execute(input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	resp := KeyStatusResponse{
		ActiveKeysCount: len(keys),
		Keys:            keys,
	}

	httpserver.HTTPSuccess(resp, w, r)
}

// Response types
type KeyStatusResponse struct {
	ActiveKeysCount int                       `json:"activeKeysCount"`
	Keys            []auth.SigningKeyStatView `json:"keys"`
}
