// go-coronanet - Coronavirus social distancing network
// Copyright (c) 2020 Péter Szilágyi. All rights reserved.

package rest

import (
	"encoding/json"
	"net/http"

	"github.com/coronanet/go-coronanet"
)

// servePairing serves API calls concerning the contact pairing.
func (api *api) servePairing(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Creates a pairing session for contact establishment
		switch secret, address, err := api.backend.InitPairing(); err {
		case coronanet.ErrNetworkDisabled:
			http.Error(w, "Cannot pair while offline", http.StatusForbidden)
		case nil:
			w.Header().Add("Content-Type", "application/json")
			json.NewEncoder(w).Encode(append(secret, address...))
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "GET":
		// Waits for a pairing session to complete
		switch id, err := api.backend.WaitPairing(); err {
		case coronanet.ErrNotPairing:
			http.Error(w, "No pairing session in progress", http.StatusForbidden)
		case nil:
			// Pairing succeeded, try to inject the contact into the backend
			switch uid, err := api.backend.AddContact(id); err {
			case coronanet.ErrContactExists:
				http.Error(w, "Remote contact already paired", http.StatusConflict)
			case nil:
				w.Header().Add("Content-Type", "application/json")
				json.NewEncoder(w).Encode(uid)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "PUT":
		// Waits for a pairing session to complete

		// Read the pairing secret from the command line
		var blob []byte
		if err := json.NewDecoder(r.Body).Decode(&blob); err != nil { // Bit unorthodox, but we don't want callers to interpret the data
			http.Error(w, "Provided pairing secret is invalid: "+err.Error(), http.StatusBadRequest)
			return
		}
		if len(blob) != 64 {
			http.Error(w, "Provided pairing secret is invalid: not 64 bytes", http.StatusBadRequest)
			return
		}
		switch id, err := api.backend.JoinPairing(blob[:32], blob[32:]); err {
		case coronanet.ErrNetworkDisabled:
			http.Error(w, "Cannot pair while offline", http.StatusForbidden)
		case nil:
			// Pairing succeeded, try to inject the contact into the backend
			switch uid, err := api.backend.AddContact(id); err {
			case coronanet.ErrContactExists:
				http.Error(w, "Remote contact already paired", http.StatusConflict)
			case nil:
				w.Header().Add("Content-Type", "application/json")
				json.NewEncoder(w).Encode(uid)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
