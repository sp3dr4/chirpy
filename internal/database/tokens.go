package database

import (
	"errors"
	"time"

	"github.com/sp3dr4/chirpy/internal/entities"
)

func findRefreshTokenByToken(refreshTokens map[int]entities.RefreshToken, token string) (*entities.RefreshToken, bool) {
	for _, t := range refreshTokens {
		if t.Token == token {
			return &t, true
		}
	}
	return nil, false
}

func (db *DB) SaveRefreshToken(userId int, token string, expiresAt time.Time) (*entities.RefreshToken, error) {
	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	tokenObj := entities.RefreshToken{
		UserId:    userId,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	dbObj.RefreshTokens[tokenObj.UserId] = tokenObj
	if err = db.writeDB(*dbObj); err != nil {
		return nil, err
	}
	return &tokenObj, nil
}

func (db *DB) GetRefreshToken(token string) (*entities.RefreshToken, error) {
	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	tokenObj, found := findRefreshTokenByToken(dbObj.RefreshTokens, token)
	if !found {
		return nil, errors.New("refresh token not found")
	}

	return tokenObj, nil
}

func (db *DB) DeleteRefreshToken(token string) error {
	dbObj, err := db.loadDB()
	if err != nil {
		return err
	}
	tokenObj, found := findRefreshTokenByToken(dbObj.RefreshTokens, token)
	if !found {
		return errors.New("refresh token not found")
	}
	delete(dbObj.RefreshTokens, tokenObj.UserId)
	if err = db.writeDB(*dbObj); err != nil {
		return err
	}

	return nil
}
