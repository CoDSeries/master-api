package main

import (
	"fmt"
	"log"
	"master-api/api"
	"master-api/app/lobby"
	"net/http"
)

func main() {
	fmt.Println("Hello CoDSeries !")

	lobbyManager := lobby.NewLobbyManager()

	mux := api.Router(lobbyManager)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
