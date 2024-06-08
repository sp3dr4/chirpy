package database

import "github.com/sp3dr4/chirpy/internal/entities"

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(userId int, body string) (*entities.Chirp, error) {
	db.chirpIdMux.Lock()
	db.chirpLastId += 1
	db.chirpIdMux.Unlock()
	chirp := entities.Chirp{Id: db.chirpLastId, Body: body, UserId: userId}
	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	dbObj.Chirps[chirp.Id] = chirp
	if err = db.writeDB(*dbObj); err != nil {
		return nil, err
	}
	return &chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]entities.Chirp, error) {
	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirps := make([]entities.Chirp, 0, len(dbObj.Chirps))
	for _, value := range dbObj.Chirps {
		chirps = append(chirps, value)
	}
	return chirps, nil
}
