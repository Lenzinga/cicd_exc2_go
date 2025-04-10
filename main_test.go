package main_test

import (
	"log"
	"os"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"

	"strings"

	git "github.com/Lenzinga/cicd_exc2_go" // Adjust the import path as necessary
)

var a git.App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

type product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {

	clearTable()

	var jsonStr = []byte(`{"name":"test product", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

// main_test.go

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func TestUpdateProduct(t *testing.T) {

	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestSearchProductsByName(t *testing.T) {
	clearTable()

	// Add some products
	a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Running Shoes", 59.99)
	a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Basketball Shoes", 89.99)
	a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Backpack", 39.99)

	// Search for "shoes"
	req, _ := http.NewRequest("GET", "/products/search?name=shoes", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var products []product
	err := json.Unmarshal(response.Body.Bytes(), &products)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}

	for _, p := range products {
		if !strings.Contains(strings.ToLower(p.Name), "shoes") {
			t.Errorf("Product name does not contain 'shoes': %s", p.Name)
		}
	}
}

func TestSortProducts(t *testing.T) {
	clearTable()

	// Insert in unsorted order
	a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Zebra", 30.0)
	a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Apple", 10.0)
	a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Mango", 20.0)

	// --- Test sort by name ascending ---
	req, _ := http.NewRequest("GET", "/products?sort=name&order=asc", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var productsByName []product
	json.Unmarshal(response.Body.Bytes(), &productsByName)

	if productsByName[0].Name != "Apple" || productsByName[2].Name != "Zebra" {
		t.Errorf("Products not sorted by name ascending: %+v", productsByName)
	}

	// --- Test sort by price descending ---
	req, _ = http.NewRequest("GET", "/products?sort=price&order=desc", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var productsByPrice []product
	json.Unmarshal(response.Body.Bytes(), &productsByPrice)

	if productsByPrice[0].Price != 30.0 || productsByPrice[2].Price != 10.0 {
		t.Errorf("Products not sorted by price descending: %+v", productsByPrice)
	}
}
