package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

//unit test to check add function (takes transaction, updates sorted transaction array and balance map)
func TestAddPoints(t *testing.T) {
	var testTransactions [5]Transaction = [5]Transaction{
		{"DANNON", 1000, "2020-11-02T14:00:00Z"},
		{"UNILEVER", 200, "2020-10-31T11:00:00Z"},
		{"DANNON", -200, "2020-10-31T15:00:00Z"},
		{"MILLER COORS", 10000, "2020-11-01T14:00:00Z"},
		{"DANNON", 300, "2020-10-31T10:00:00Z"}}
	for tnsactn := range testTransactions {
		add(testTransactions[tnsactn])
		if len(sortedTransactions) != (tnsactn + 1) {
			t.Errorf("Add function not accepting valid transactions: %d", len(sortedTransactions))
		}
	}
	sort.Sort(sortedTransactions)
}

//unit test to check deductTransactions function (takes point value, returns list of deductions in the form {"payer": _, "points": _})
func TestSpendPoints(t *testing.T) {
	balancesList, err := deductTransactions(5000)
	if err != nil {
		t.Error(err.Error())
	}
	for balance := range balancesList {
		payer := balancesList[balance].Payer
		if payer == "DANNON" && balancesList[balance].Points != -100 {
			t.Errorf("deductTransactions function not taking proper points from \"DANNON\": %d", balancesList[balance].Points)
		} else if payer == "UNILEVER" && balancesList[balance].Points != -200 {
			t.Errorf("deductTransactions function not taking proper points from \"UNILEVER\": %d", balancesList[balance].Points)
		} else if payer == "MILLER COORS" && balancesList[balance].Points != -4700 {
			t.Errorf("deductTransactions function not taking proper points from \"MILLER COORS\": %d", balancesList[balance].Points)
		}
	}
}

//unit test to check the getBalances function (takes no input, returns the balance map)
func TestGetPoints(t *testing.T) {
	currentBalances := getBalances()
	if currentBalances["DANNON"] != 1000 {
		t.Error("Invalid balance for \"DANNON\"")
	}
	if currentBalances["UNILEVER"] != 0 {
		t.Error("Invalid balance for \"UNILEVER\"")
	}
	if currentBalances["MILLER COORS"] != 5300 {
		t.Error("Invalid balance for \"MILLER COORS\"")
	}
	t.Cleanup(cleanTests)
}

//reset the global data structures between tests
func cleanTests() {
	sortedTransactions = nil
	for balance := range Balances {
		delete(Balances, balance)
	}
}

func TestAddPointsHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/add", addTransaction)
	badTransaction := Transaction{"TESLA", int(0), "2020-11-01T14:00:00Z"}
	postBody, _ := json.Marshal(badTransaction)
	reqBody := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "http://localhost:8080/add", reqBody)
	if err != nil {
		t.Errorf("Failed to create mock HTTP add request: %s" + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("Failed add route")
	}
	t.Cleanup(cleanTests)
}

//route test to check if the /spend route is working as intended with valid input
func TestSpendPointsHTTPValid(t *testing.T) {
	TestAddPoints(t)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/spend", spendPoints)
	pointMap := make(map[string]int)
	pointMap["points"] = 5000
	postBody, _ := json.Marshal(pointMap)
	reqBody := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "http://localhost:8080/spend", reqBody)
	if err != nil {
		t.Errorf("Failed to create mock HTTP spend request: %s" + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Error("Failed spend route")
	}
	t.Cleanup(cleanTests)
}

//route test to check if the /spend route is working as intended with invalid input
func TestSpendPointsHTTPInvalid(t *testing.T) {
	TestAddPoints(t)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/spend", spendPoints)
	pointMap := make(map[string]int)
	pointMap["points"] = 9999999
	postBody, _ := json.Marshal(pointMap)
	reqBody := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "http://localhost:8080/spend", reqBody)
	if err != nil {
		t.Errorf("Failed to create mock HTTP spend request: %s" + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Error("Failed to reject invalid spend request (more points than available)")
	}
	t.Cleanup(cleanTests)
}

//route test to check if the /points route is returning properly
func TestGetPointsHTTP(t *testing.T) {
	TestAddPoints(t)
	r := gin.Default()
	r.GET("/points", getPoints)
	req, err := http.NewRequest("GET", "http://localhost:8080/points", nil)
	if err != nil {
		t.Errorf("Failed to create mock HTTP points request: %s" + err.Error())
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Error("Failed to reject invalid spend request (spent more points than available)")
	}
	t.Cleanup(cleanTests)
}
