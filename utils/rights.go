package utils

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

/*
Token JWT claims struct
*/
type Token struct {
	UserId     uint
	UserRights RightBits
	jwt.StandardClaims
}

//RightBits set a right
type RightBits uint32

const (
	NoRight RightBits = 0
)

//Set a right
func Set(b, flag RightBits) RightBits { return b | flag }

//Clear a right
func Clear(b, flag RightBits) RightBits { return b &^ flag }

//Toggle a Right
func Toggle(b, flag RightBits) RightBits { return b ^ flag }

//Has a Right
func Has(b, flag RightBits) bool { return b&flag == flag }

//Set a right
func (b RightBits) Set(flag RightBits) RightBits { return b | flag }

//Clear a right
func (b RightBits) Clear(flag RightBits) RightBits { return b &^ flag }

//Toggle a right
func (b RightBits) Toggle(flag RightBits) RightBits { return b ^ flag }

//Has a right
func (b RightBits) Has(flag RightBits) bool { return b&flag == flag }

//HasRightsRequest check for correct role
func HasRightsRequest(r *http.Request, rights RightBits) bool {
	user, ok := GetAuthenticatedToken(r)
	return ok && user.UserRights.Has(rights)
}

//HasRights check for correct role
func HasRights(r *http.Request, rights RightBits, f func(r *http.Request, rights RightBits) bool) bool {
	auth, ok := GetAuthenticatedToken(r)
	//If has no right and no auth then OK
	if !ok && !rights.Has(NoRight) {
		return false
	}
	//If user is ok, then check right and function
	ret := f(r, rights)
	return auth.UserRights.Has(rights) && ret
}

//GetAuthenticatedToken return the Authentication token if exists in context
func GetAuthenticatedToken(r *http.Request) (Token, bool) {
	user, ok := r.Context().Value("user").(Token)
	return user, ok
}
