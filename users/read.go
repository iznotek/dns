package users

import (
	"encoding/json"
	"github.com/iznotek/dns/db"
	"github.com/iznotek/dns/util"
	bolt "go.etcd.io/bbolt"
	"net/http"
)

func read(w http.ResponseWriter, r *http.Request, database *bolt.DB) {
	// Validate initial request with request type
	if r.Method != "GET" {
		util.Responses.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	} else if r.Header.Get("Authorization") == "" {
		util.Responses.Error(w, http.StatusUnauthorized, "header 'Authorization' is required")
		return
	}

	// Verify JWT in headers
	token, err := db.TokenFromString(r.Header.Get("Authorization"), database)
	if err != nil {
		util.Responses.Error(w, http.StatusUnauthorized, "failed to authenticate: "+err.Error())
		return
	}

	// Get user from database
	u, err := db.UserFromToken(token, database)
	if err != nil {
		util.Responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Operate differently if admin
	username := u.Username
	if u.Role == "admin" && r.URL.Query().Get("user") != "" {
		// Allow operating on different user if admin
		username = r.URL.Query().Get("user")
	}

	// Get list of all users if admin
	if username == "*" && u.Role == "admin" {
		var users []map[string]interface{}

		if err := database.View(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte("users")).ForEach(func(k, v []byte) error {
				// Decode user JSON
				var u db.User
				if err := json.Unmarshal(v, &u); err != nil {
					return err
				}

				// Remove password hash from user data
				userData := map[string]interface{}{}
				userData["name"] = u.Name
				userData["username"] = u.Username
				userData["role"] = u.Role
				userData["logins"] = u.Tokens

				users = append(users, userData)

				return nil
			})
		}); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to retrieve all users: "+err.Error())
			return
		}

		util.Responses.SuccessWithData(w, users)
		return
	}

	// Retrieve user from database
	rawUser, err := db.UserFromDatabase(username, database)
	if err != nil {
		util.Responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Remove password hash from user data
	userData := map[string]interface{}{}
	userData["name"] = rawUser.Name
	userData["username"] = rawUser.Username
	userData["role"] = rawUser.Role
	userData["logins"] = rawUser.Tokens

	// Return user data
	util.Responses.SuccessWithData(w, userData)
}
