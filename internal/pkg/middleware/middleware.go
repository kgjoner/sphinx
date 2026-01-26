package middleware

type Middleware struct {
	authorizer Authorizer
}

func New(authorizer Authorizer) *Middleware {
	return &Middleware{
		authorizer: authorizer,
	}
}