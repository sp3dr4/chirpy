package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/sp3dr4/chirpy/internal/entities"
)

var ErrDuplicateUser = fmt.Errorf("user with email already exists")

type DB struct {
	debug bool

	mux  *sync.RWMutex
	path string

	chirpIdMux  *sync.RWMutex
	chirpLastId int

	userIdMux  *sync.RWMutex
	userLastId int
}

type DBStructure struct {
	Chirps map[int]entities.Chirp `json:"chirps"`
	Users  map[int]entities.User  `json:"users"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string, debug bool) (*DB, error) {
	db := &DB{
		debug:       debug,
		mux:         &sync.RWMutex{},
		path:        path,
		chirpIdMux:  &sync.RWMutex{},
		chirpLastId: 0,
		userIdMux:   &sync.RWMutex{},
		userLastId:  0,
	}
	if db.debug {
		if err := os.Remove(db.path); err != nil {
			return nil, err
		}
	}
	if err := db.ensureDB(); err != nil {
		return nil, err
	}

	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	for cid := range dbObj.Chirps {
		db.chirpLastId = max(db.chirpLastId, cid)
	}

	for uid := range dbObj.Users {
		db.userLastId = max(db.userLastId, uid)
	}

	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (*entities.Chirp, error) {
	db.chirpIdMux.Lock()
	db.chirpLastId += 1
	db.chirpIdMux.Unlock()
	chirp := entities.Chirp{Id: db.chirpLastId, Body: body}
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

func findUserByEmail(users map[int]entities.User, email string) (*entities.User, bool) {
	for _, u := range users {
		if u.Email == email {
			return &u, true
		}
	}
	return nil, false
}

// CreateUser creates a new user and saves it to disk
func (db *DB) CreateUser(email, password string) (*entities.User, error) {
	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	if _, exists := findUserByEmail(dbObj.Users, email); exists {
		return nil, ErrDuplicateUser
	}

	db.userIdMux.Lock()
	db.userLastId += 1
	db.userIdMux.Unlock()
	user := entities.User{
		Id:       db.userLastId,
		Email:    email,
		Password: password,
	}

	dbObj.Users[user.Id] = user
	if err = db.writeDB(*dbObj); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers returns all users in the database
func (db *DB) GetUsers() ([]entities.User, error) {
	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	users := make([]entities.User, 0, len(dbObj.Users))
	for _, value := range dbObj.Users {
		users = append(users, value)
	}
	return users, nil
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
	dbObj := DBStructure{
		Chirps: map[int]entities.Chirp{},
		Users:  map[int]entities.User{},
	}
	if err = db.writeDB(dbObj); err != nil {
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
