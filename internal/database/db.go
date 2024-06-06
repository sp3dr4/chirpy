package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/sp3dr4/chirpy/internal/entities"
)

type DB struct {
	mux  *sync.RWMutex
	path string

	idMux  *sync.RWMutex
	lastId int
}

type DBStructure struct {
	Chirps map[int]entities.Chirp `json:"chirps"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		mux:    &sync.RWMutex{},
		path:   path,
		idMux:  &sync.RWMutex{},
		lastId: 0,
	}
	if err := db.ensureDB(); err != nil {
		return nil, err
	}
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (*entities.Chirp, error) {
	db.idMux.Lock()
	db.lastId += 1
	db.idMux.Unlock()
	chirp := entities.Chirp{Id: db.lastId, Body: body}
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
	chirps := []entities.Chirp{}
	for _, c := range dbObj.Chirps {
		chirps = append(chirps, c)
	}
	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err = db.writeDB(DBStructure{Chirps: map[int]entities.Chirp{}}); err != nil {
		return err
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (*DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dat, err := os.ReadFile(db.path)
	if err != nil {
		return nil, err
	}
	dbstruct := &DBStructure{}
	err = json.Unmarshal(dat, dbstruct)
	if err != nil {
		return nil, err
	}
	return dbstruct, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, dat, 0640)
	if err != nil {
		fmt.Printf("WriteFile err: %v\n", err)
		return err
	}
	return nil
}
