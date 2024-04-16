// app.go

package main

import (
	"database/sql"
	// tom: for Initialize
	"fmt"
	"log"

	"encoding/json"
	// tom: for route handlers
	"net/http"
	"strconv"

	// tom: go get required
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s port=5416 dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()

	// tom: this line is added after initializeRoutes is created later on
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}
	if p, err := getP(w, id, a); err == nil {
		respondWithJSON(w, http.StatusOK, p)
	}
}
func getP(w http.ResponseWriter, id int, a *App) (product, error) {
	p := product{ID: id}
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return p, err
	}
	return p, nil
}
func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if u, err := getU(w, id, a); err == nil {
		respondWithJSON(w, http.StatusOK, u)
	}
}
func getU(w http.ResponseWriter, id int, a *App) (user, error) {
	u := user{ID: id}
	if err := u.getUser(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return u, err
	}
	return u, nil
}
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}
func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	if start < 0 {
		start = 0
	}
	users, err := getUsers(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}
func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}
func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := u.createUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}
func (a *App) addToCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	pid, pErr := strconv.Atoi(vars["pid"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID "+strconv.Itoa(id))
		return
	}
	if pErr != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID "+strconv.Itoa(pid))
		return
	}
	user, _ := getU(w, id, a)
	c := cart{ID: user.CartId, UserID: user.ID}
	if err := c.getCart(a.DB); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get cart with userID: "+strconv.Itoa(user.ID))
		return
	}
	p, err := getP(w, pid, a)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not find product "+strconv.Itoa(p.ID))
		return
	}
	if err := c.addToCart(a.DB, &p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not add to cart")
		return
	}
	respondWithJSON(w, http.StatusAccepted, c)
}
func (a *App) removeFromCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	pid, pErr := strconv.Atoi(vars["pid"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID "+strconv.Itoa(id))
		return
	}
	if pErr != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID "+strconv.Itoa(pid))
		return
	}
	user, _ := getU(w, id, a)
	c := cart{ID: user.CartId, UserID: user.ID}
	if err := c.getCart(a.DB); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get cart with userID: "+strconv.Itoa(user.ID))
		return
	}
	p, err := getP(w, pid, a)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not find product "+strconv.Itoa(p.ID))
		return
	}
	if err := c.deleteFromCart(a.DB, &p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not delete product from cart")
		return
	}
	respondWithJSON(w, http.StatusAccepted, c)
}
func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"]) //conversion str to int
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	p := product{ID: id}
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	u := user{ID: id}
	if err := u.deleteUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) initializeRoutes() {
	// products
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
	// users
	a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
	a.Router.HandleFunc("/users", a.createUser).Methods("POST")
	a.Router.HandleFunc("/users/{id:[0-9]*}&{pid:[0-9]+}", a.addToCart).Methods("POST")
	a.Router.HandleFunc("/users/del/{id:[0-9]*}&{pid:[0-9]+}", a.removeFromCart).Methods("POST")
	a.Router.HandleFunc("/users/{id:[0-9]*}", a.getUser).Methods("GET")
	a.Router.HandleFunc("/users/{id:[0-9]*}", a.deleteUser).Methods("DELETE")
}
