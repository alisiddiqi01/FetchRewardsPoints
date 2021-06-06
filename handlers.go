package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

//getPoints returns the map of {"payer":points} balance objects
func getPoints(c *gin.Context) {
	var balances = getBalances()
	c.JSON(http.StatusOK, balances)
}

//addTransaction adds a transaction to the sorted list of transactions
func addTransaction(c *gin.Context) {
	var transaction Transaction
	err := c.ShouldBindJSON(&transaction)
	//input doesn't properly bind with transaction object
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	//time input is not properly formatted to layout "RFC3339"
	_, err = time.Parse(time.RFC3339, transaction.Timestamp)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	//transaction is formatted correctly with a valid timestamp
	add(transaction)
	sort.Sort(sortedTransactions)
	c.JSON(http.StatusOK, "Transaction added successfully")
}

//spendPoints extracts the desired spend amount from the request and calls deductTransactions to spend them
func spendPoints(c *gin.Context) {
	//map to store the desired amount of points to spend
	var toSpend map[string]int
	p, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	json.Unmarshal(p, &toSpend)
	_, valid := toSpend["points"]
	//input formatted incorrectly
	if !valid {
		c.AbortWithStatus(400)
		return
	} else if toSpend["points"] < 0 { //a user cannot spend negative points
		c.AbortWithStatus(400)
		return
	}
	deductions, err := deductTransactions(toSpend["points"])
	//deductTransactions returns an error if there are insufficient points to spend
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, deductions)
}
