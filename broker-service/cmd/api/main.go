package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "8020"

type Config struct{}

func main() {
	app := Config{}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}
