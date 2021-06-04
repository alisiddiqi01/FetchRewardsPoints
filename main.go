// main

package main

import (
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	gin.SetMode(gin.ReleaseMode)
	//setting up the router
	router = gin.Default()
	//initializing the 3 routes as specified
	initializeRoutes()
	//beginnning the service
	router.Run()
}
