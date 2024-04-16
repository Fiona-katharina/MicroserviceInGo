package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func clearUserTable() {
	a.DB.Exec("DELETE FROM users")
	a.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
}

func clearCartsTable() {
	a.DB.Exec("DELETE FROM carts")
	a.DB.Exec("ALTER SEQUENCE carts_id_seq RESTART WITH 1")
}
func addUsers(count int) {
	if count < 1 {
		count = 1
	}
}
func TestGetNonExistentUser(t *testing.T) {
	clearUserTable()

	req, _ := http.NewRequest("GET", "/users/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "User not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["error"])
	}
}
func TestCreateUser(t *testing.T) {
	clearUserTable()
	var jsonStr = []byte(`{"name":"example user"}`)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "example user" {
		t.Errorf("Expected user name to be 'example user'. Got '%v'", m["name"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected user ID to be '1'. Got '%v'", m["id"])
	}
}
func TestDeleteUser(t *testing.T) {
	clearUserTable()
	var jsonStr = []byte(`{"name":"example user"}`)
	req1, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonStr))
	req1.Header.Set("Content-Type", "application/json")

	response1 := executeRequest(req1)
	checkResponseCode(t, http.StatusCreated, response1.Code)

	req, _ := http.NewRequest("GET", "/users/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/users/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/users/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
func TestAddToCart(t *testing.T) {
	clearUserTable()
	clearCartsTable()
	var jsonStr = []byte(`{"name":"example user"}`)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	// Then create product
	clearProductTable()
	var jsonStr2 = []byte(`{"name":"test product", "price": 11.22}`)
	req2, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr2))
	req2.Header.Set("Content-Type", "application/json")
	response2 := executeRequest(req2)
	checkResponseCode(t, http.StatusCreated, response2.Code)

	req3, _ := http.NewRequest("POST", "/users/1&1", nil)
	response3 := executeRequest(req3)
	checkResponseCode(t, http.StatusAccepted, response3.Code)

	b := response3.Body
	if b.String() != "{\"id\":1,\"userID\":1,\"items\":[1],\"balance\":11.22}" {
		t.Errorf("Expected {\"id\":1,\"userID\":1,\"items\":[1],\"balance\":11.22}. Got '%v'", b.String())
	}
}
func TestRemoveFromCart(t *testing.T) {
	clearUserTable()
	clearCartsTable()
	var jsonStr = []byte(`{"name":"example user"}`)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	// Then create product
	clearProductTable()
	var jsonStr2 = []byte(`{"name":"test product", "price": 11.22}`)
	req2, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr2))
	req2.Header.Set("Content-Type", "application/json")
	response2 := executeRequest(req2)
	checkResponseCode(t, http.StatusCreated, response2.Code)

	req3, _ := http.NewRequest("POST", "/users/1&1", nil)
	response3 := executeRequest(req3)
	checkResponseCode(t, http.StatusAccepted, response3.Code)

	req4, _ := http.NewRequest("POST", "/users/del/1&1", nil)
	response4 := executeRequest(req4)
	checkResponseCode(t, http.StatusAccepted, response4.Code)

	b := response4.Body
	if b.String() != "{\"id\":1,\"userID\":1,\"items\":[],\"balance\":0}" {
		t.Errorf("Expected {\"id\":1,\"userID\":1,\"items\":[],\"balance\":0}. Got '%v'", b.String())
	}
}
