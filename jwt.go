package main

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type ClaimBody struct {
	*jwt.StandardClaims
	TokenType string
	Email string
	ID uint16
	Name string
	Doctor bool
	Avatar string
}

func (v *ViewData) CreateToken() (token string, err error) {

	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Claims = &ClaimBody{
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 5).Unix(),
		},
		"level1",
		v.Email,
		v.ID,
		v.Name,
		v.Doctor,
		v.AvatarURI,
	}

	return t.SignedString(signKey)
}