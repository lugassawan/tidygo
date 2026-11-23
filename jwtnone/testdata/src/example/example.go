package example

import (
	legacy "github.com/dgrijalva/jwt-go"
	. "github.com/golang-jwt/jwt/v4"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func sign() jwtv5.SigningMethod {
	return jwtv5.SigningMethodNone // want `jwt.SigningMethodNone produces unsigned tokens`
}

func verify() any {
	return jwtv5.UnsafeAllowNoneSignatureType // want `jwt.UnsafeAllowNoneSignatureType lets unsigned tokens verify`
}

func aliased() legacy.SigningMethod {
	return legacy.SigningMethodNone // want `jwt.SigningMethodNone produces unsigned tokens`
}

func dotImport() SigningMethod {
	_ = UnsafeAllowNoneSignatureType // want `jwt.UnsafeAllowNoneSignatureType lets unsigned tokens verify`
	return SigningMethodNone         // want `jwt.SigningMethodNone produces unsigned tokens`
}

func hmac() jwtv5.SigningMethod {
	return jwtv5.SigningMethodHS256
}

func shadowed() int {
	SigningMethodNone := 1
	return SigningMethodNone
}

type local struct{ SigningMethodNone int }

func field(l local) int {
	return l.SigningMethodNone
}
