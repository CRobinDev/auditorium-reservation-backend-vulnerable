package middleware

import "github.com/nathakusuma/auditorium-reservation-backend/pkg/jwt"

type Middleware struct {
	jwt jwt.IJwt
}

func NewMiddleware(
	jwt jwt.IJwt,
) *Middleware {
	return &Middleware{
		jwt: jwt,
	}
}
