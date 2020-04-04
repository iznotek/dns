package roles

import (
	"encoding/json"
	"github.com/iznotek/dns/db"
	"github.com/iznotek/dns/util"
	bolt "go.etcd.io/bbolt"
	"net/http"
)

func update(w http.ResponseWriter, r *http.Request, path string, database *bolt.DB) {
	// Validate initial request with type, body exists, and headers
	if r.Method != "PUT" {
		util.Responses.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	} else if r.Body == nil {
		util.Responses.Error(w, http.StatusBadRequest, "body must be present")
		return
	} else if r.Header.Get("Content-Type") != "application/json" {
		util.Responses.Error(w, http.StatusBadRequest, "record must be specified in path")
		return
	} else if r.Header.Get("Authorization") == "" {
		util.Responses.Error(w, http.StatusUnauthorized, "header 'Authorization' must be present")
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

	// Check user role
	if u.Role != "admin" {
		util.Responses.Error(w, http.StatusForbidden, "user must be of role 'admin'")
		return
	}

	// Validate body by decoding json, checking fields exist, and checking field type
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		util.Responses.Error(w, http.StatusBadRequest, "failed to decode body: "+err.Error())
		return
	}
	validationErr, valid := util.ValidateBody(body, []string{"description", "allow", "deny"}, map[string]map[string]string{
		"description": {"type": "string", "required": "true"},
		"allow": {"type": "string", "required": "true"},
		"deny": {"type": "string", "required": "true"},
	})
	if validationErr != "" {
		util.Responses.Error(w, http.StatusBadRequest, validationErr)
		return
	}

	// Get role from database
	role, err := db.GetRole(r.URL.Path[len(path):], database)
	if role == nil {
		util.Responses.Error(w, http.StatusBadRequest, "specified role does not exist")
		return
	}

	// Update values if they exist in the body
	if valid["description"] {
		role.Description = body["description"].(string)
	}
	if valid["allow"] {
		role.Allow = body["allow"].(string)
	}
	if valid["deny"] {
		role.Deny = body["deny"].(string)
	}

	// Save to database
	if err := db.CreateRole(role.Name, role.Description, role.Allow, role.Deny, database); err != nil {
		util.Responses.Error(w, http.StatusInternalServerError, "failed to write role to database: "+err.Error())
		return
	}

	util.Responses.Success(w)
}
