package main

import (
	"errors"
	"math"
	"sort"
	"time"
)

//the type passed back in an array from a spendPoints call for each payer and their deduction
type Balance struct {
	Payer  string `json:"payer"`
	Points int    `json:"points"`
}

//the type used to store transactions in the sorted transactions array
type Transaction struct {
	Payer     string `json:"payer" binding:"required"`
	Points    int    `json:"points" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
}

type transactionArray []Transaction

//implementing the necessary functions for a Transaction array to allow for sorting with built in sort
func (th transactionArray) Len() int { return len(th) }

func (th transactionArray) Less(i, j int) bool {
	iTime, _ := time.Parse(time.RFC3339, th[i].Timestamp)
	jTime, _ := time.Parse(time.RFC3339, th[j].Timestamp)

	return iTime.Before(jTime)
}

func (th transactionArray) Swap(i, j int) {
	th[i], th[j] = th[j], th[i]
}

//map to keep track of each payer's current points (stored in a payer:points pattern), updated upon each transaction
var Balances map[string]int = make(map[string]int)

//sorted array to keep track of transactions in order of timestamp, updated after add and spend calls
var sortedTransactions transactionArray

//function to add a transaction to both the sorted transaction array and the balances map
func add(a Transaction) {
	//add transaction to sorted array (we sort in batches after deducting points, so sorting takes place outside of this function)
	sortedTransactions = append(sortedTransactions, a)
	//add balance to balances map
	_, inBalances := Balances[a.Payer]
	if inBalances {
		Balances[a.Payer] += a.Points
	} else {
		Balances[a.Payer] = a.Points
	}
}

//function to spend points, takes in the desired spend amount and returns the corresponding deductions (or an error)
func deductTransactions(p int) ([]Balance, error) {
	//First go through transactions newest -> oldest, in order to set how many points are available
	// at each transaction (if a future transaction is negative, then the current transaction
	// cannot have more points available than the future transaction amount)
	returnList := []Balance{}
	var deductions map[string]Balance = make(map[string]Balance)
	var futureDeductions map[string]int = make(map[string]int)
	var availableHere []int = make([]int, len(sortedTransactions))
	for i := len(sortedTransactions) - 1; i >= 0; i-- {
		cur := sortedTransactions[i]
		_, prs := futureDeductions[cur.Payer]
		//if the payer has a pending future deduction
		if prs {
			if cur.Points < 0 { //if this transaction is also a deduction
				futureDeductions[cur.Payer] += cur.Points
				availableHere[i] = 0
				continue
			} else { //if this transaction can reduce the future deduction amount
				futureDeductions[cur.Payer] += cur.Points
				if futureDeductions[cur.Payer] >= 0 { //this transaction is more positive than future deductions
					availableHere[i] = futureDeductions[cur.Payer]
					delete(futureDeductions, cur.Payer)
				} else { //this transaction is not enough to offset future deductions
					availableHere[i] = 0
					continue
				}
			}
		} else { //if the payer currently has no pending future deductions
			if cur.Points < 0 { //if this is a deduction
				futureDeductions[cur.Payer] = cur.Points
				availableHere[i] = 0
				continue
			} else { //this is a positive amount and the player has no current deductions pending
				availableHere[i] = cur.Points
			}
		}
	}

	//Next iterate through transactions oldest -> newest, keeping track of how much to deduct from each payer
	// with a *positive* "availableHere" value. Update the current payer's temporary deduction balance's points
	// with the negative of min(remaining balance, availableHere value)
	//Continue this for each transaction until the remaining balance is 0. if we go through all the transactions
	// and we are still > 0, return an error and say that there are not enough points.
	// Otherwise, the remaining balance hit 0 and we can add each deduction transaction and return the balances to the user
	remainingBalance := p
	for b := range sortedTransactions {
		//skip deductions
		if sortedTransactions[b].Points <= 0 {
			continue
		}
		//establish current payer
		curPayer := sortedTransactions[b].Payer
		//exit to return output
		if remainingBalance == 0 {
			break
		} else if availableHere[b] > 0 { //this payer has points available
			deductions[curPayer] = Balance{curPayer, -1 * int(math.Min(float64(availableHere[b]), float64(remainingBalance)))}
			remainingBalance += deductions[curPayer].Points
		}

	}
	if remainingBalance == 0 {
		//create transactions for every payer's deductions and populate balances return list
		for pyr, pnts := range deductions {
			add(Transaction{pyr, pnts.Points, time.Now().Format(time.RFC3339)})
			returnList = append(returnList, pnts)
		}
		sort.Sort(sortedTransactions)
		return returnList, nil

	} else {
		return returnList, errors.New("Insufficent points to spend")
	}

}

//function to return all current payer balances
//returns the map key-value pairs
func getBalances() map[string]int {
	return Balances
}
