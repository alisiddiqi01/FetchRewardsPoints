package main

func initializeRoutes() {
	//route to return the current payers' balances
	router.GET("/points", getPoints)
	//route to add points
	router.POST("/add", addTransaction)
	//route to spend points
	router.POST("/spend", spendPoints)
}
