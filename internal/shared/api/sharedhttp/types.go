package sharedhttp

/* ==============================================================================
	Context Keys
============================================================================== */

type ctxKey string

const (
	ActorCtxKey  ctxKey = "sphinx_actor"
	TokenCtxKey  ctxKey = "sphinx_token"
	TargetIDCtxKey ctxKey = "sphinx_target_id"
)