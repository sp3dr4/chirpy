package database

import "github.com/sp3dr4/chirpy/internal/entities"

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

// UpdateUser updates a user attributes and returns it
func (db *DB) UpdateUser(user *entities.User) (*entities.User, error) {
	dbObj, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	for i, u := range dbObj.Users {
		if u.Id == user.Id {
			dbObj.Users[i] = *user
		}
	}
	if err = db.writeDB(*dbObj); err != nil {
		return nil, err
	}
	return user, nil
}
