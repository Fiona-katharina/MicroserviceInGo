package main

import (
	"database/sql"
	"github.com/lib/pq"
)

type user struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	CartId int    `json:"cartID"`
}

func (u *user) getUser(db *sql.DB) error {
	return db.QueryRow("SELECT name FROM users WHERE id=$1",
		u.ID).Scan(&u.Name)
}

func (u *user) updateUser(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE users SET name=$1 WHERE id=$3",
			u.Name, u.ID)

	return err
}

func (u *user) deleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE id=$1", u.ID)

	return err
}

func (u *user) createUser(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO users(name, cartID) VALUES($1, $2) RETURNING id",
		u.Name, -1).Scan(&u.ID)
	if err != nil {
		return err
	}
	//var c = new(cart)
	var itms []int
	err = db.QueryRow(
		"INSERT INTO carts(userID, products) VALUES($1, $2) RETURNING id",
		u.ID, pq.Array(itms)).Scan(&u.CartId)
	if err != nil {
		return err
	}
	return nil
}

func getUsers(db *sql.DB, start, count int) ([]user, error) {
	rows, err := db.Query(
		"SELECT id, name FROM users LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []user{}

	for rows.Next() {
		var u user
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
