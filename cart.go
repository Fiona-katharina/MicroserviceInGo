package main

import (
	"database/sql"
	"github.com/lib/pq"
)

type cart struct {
	ID      int           `json:"id"`
	UserID  int           `json:"userID"`
	Items   pq.Int64Array `json:"items"`
	Balance float64       `json:"balance"`
}

func (c *cart) getCart(db *sql.DB) error {
	return db.QueryRow("SELECT balance, ID, products FROM carts WHERE userID=$1",
		c.UserID).Scan(&c.Balance, &c.ID, &c.Items)
}

func (c *cart) deleteCart(db *sql.DB) error {
	_, err := db.Exec("UPDATE carts SET products=nil, balance=0 WHERE id=$1", c.ID)

	return err
}

func (c *cart) deleteFromCart(db *sql.DB, p *product) error {
	indx := implContains(c.Items, p)
	if indx > -1 {
		c.Items = append(c.Items[:indx], c.Items[indx+1:]...)
		c.Balance -= p.Price
	}
	_, err2 :=
		db.Exec("UPDATE carts SET products=$1, balance=$2 WHERE id=$3",
			pq.Array(c.Items), c.Balance, c.ID)
	return err2
}

func (c *cart) addToCart(db *sql.DB, p *product) error {
	c.Items = append(c.Items, (int64)(p.ID))
	c.Balance += p.Price
	_, err2 :=
		db.Exec("UPDATE carts SET balance=$1, products=$2 WHERE id=$3", c.Balance, pq.Array(c.Items), c.ID)
	return err2
}

// fuction to check given string is in array or not
func implContains(sl pq.Int64Array, name *product) int {
	// iterate over the array and compare given string to each element

	for index, value := range sl {
		if value == (int64)(name.ID) {
			// return index when the element is found
			return index
		}
	}
	// if not found given string, return -1
	return -1
}
