package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultJwtExpirationSeconds int = 60 * 60
const defaultRefreshExpirationSeconds int = 60 * 24 * 60 * 60

func buildExpiration(expirationSeconds int) time.Time {
	return time.Now().Add(time.Duration(expirationSeconds) * time.Second)
}

func createJwt(userId int, secret string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(buildExpiration(defaultJwtExpirationSeconds)),
		Subject:   fmt.Sprint(userId),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

func getUserIdFromJwt(value, secret string) (int, error) {
	token, err := jwt.ParseWithClaims(
		value,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) { return []byte(secret), nil },
	)
	if err != nil {
		return 0, err
	}
	userIdStr, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func buildRandomToken() (string, error) {
	bytesLen := 32
	bToken := make([]byte, bytesLen)
	_, err := rand.Read(bToken)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(bToken)
	return token, nil
}
