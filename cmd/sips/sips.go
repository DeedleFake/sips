package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/DeedleFake/sips"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	PinsApiService := sips.NewPinsApiService()
	PinsApiController := sips.NewPinsApiController(PinsApiService)

	router := sips.NewRouter(PinsApiController)

	log.Printf("Listening on %v...", *addr)
	log.Fatal(http.ListenAndServe(*addr, router))
}
