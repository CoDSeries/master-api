package api

import (
	"encoding/json"
	"fmt"
	"master-api/app/lobby"
	"net/http"
)

// Router initialise le routeur et les routes
func Router(lobbyManager *lobby.LobbyManager) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/create_lobby", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			LobbyID string `json:"lobby_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		l := lobbyManager.CreateLobby(req.LobbyID)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Lobby created with ID: %s", l.ID)
	})

	mux.HandleFunc("/get_lobby", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("test get lobby")
		lobbyID := r.URL.Query().Get("lobby_id")
		if lobbyID == "" {
			http.Error(w, "Lobby ID is required", http.StatusBadRequest)
			return
		}
		l, exists := lobbyManager.GetLobby(lobbyID)
		if !exists {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(l)
	})

	mux.HandleFunc("/delete_lobby", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		lobbyID := r.URL.Query().Get("lobby_id")
		if lobbyID == "" {
			http.Error(w, "Lobby ID is required", http.StatusBadRequest)
			return
		}
		lobbyManager.DeleteLobby(lobbyID)
		fmt.Fprintf(w, "Lobby with ID %s deleted", lobbyID)
	})

	mux.HandleFunc("/list_lobbies", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("test get lobby")
		lobbies := lobbyManager.ListLobbies()
		if err := json.NewEncoder(w).Encode(lobbies); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return mux
}
