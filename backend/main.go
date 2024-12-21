package main

import (
	_ "context"
	"fmt"
	"log"
	"net/http"

	"github.com/hossein-nas/analytics_aggregator/router"
)

func main() {
	r := router.Router()
	fmt.Println("Starting server on the port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
