package records

import (
	"encoding/json"
	"github.com/iznotek/dns/db"
	"github.com/iznotek/dns/util"
	bolt "go.etcd.io/bbolt"
	"net"
	"net/http"
	"strings"
)

// Handle the updating of records
func update(w http.ResponseWriter, r *http.Request, path string, database *bolt.DB) {
	// Set database into operations
	db.Get.Db = database
	db.Set.Db = database
	db.Delete.Db = database

	// Validate initial request with request type, body exists, and content type
	if r.Method != "PUT" {
		util.Responses.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	} else if r.Body == nil {
		util.Responses.Error(w, http.StatusBadRequest, "body must be present")
		return
	} else if r.Header.Get("Content-Type") != "application/json" {
		util.Responses.Error(w, http.StatusBadRequest, "body must be of type JSON")
		return
	} else if len(r.URL.Path[len(path):]) == 0 {
		util.Responses.Error(w, http.StatusBadRequest, "record must be specified in path")
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

	// Get user from token
	user, err := db.UserFromToken(token, database)
	if err != nil {
		util.Responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	recordName := strings.ToLower(r.URL.Path[len(path):])

	// Check if allowed
	if allowed, err := db.EvaluateRole(user.Role, recordName, database); err != nil {
		util.Responses.Error(w, http.StatusInternalServerError, "failed to evaluate the role: "+err.Error())
		return
	} else if !allowed {
		util.Responses.Error(w, http.StatusForbidden, "role '"+user.Role+"' is not allowed to create record")
		return
	}

	// Validate body by decoding json, checking fields exists, and checking field type
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		util.Responses.Error(w, http.StatusBadRequest, "failed to decode body: "+err.Error())
		return
	} else if err, _ := util.ValidateBody(body, []string{"type"}, map[string]map[string]string{"type": {"type": "string", "required": "true"}}); err != "" {
		util.Responses.Error(w, http.StatusBadRequest, err)
		return
	}

	// Parse out body by type
	switch strings.ToUpper(body["type"].(string)) {
	case "A":
		// Get original record from database
		record := db.Get.A(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"host"}, map[string]map[string]string{"host": {"type": "ipv4", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in the body
		if valid["host"] {
			record.Address = net.ParseIP(body["host"].(string))
		}

		// Write updated values to the database
		if err := db.Set.A(recordName, record.Address.String()); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "AAAA":
		// Get original record from database
		record := db.Get.AAAA(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"host"}, map[string]map[string]string{"host": {"type": "ipv6", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in the body
		if valid["host"] {
			record.Address = net.ParseIP(body["host"].(string))
		}

		// Write updated values to the database
		if err := db.Set.AAAA(recordName, record.Address.String()); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "CNAME":
		// Get original record from database
		record := db.Get.CNAME(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"target"}, map[string]map[string]string{"target": {"type": "string", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in the body
		if valid["target"] {
			record.Target = body["target"].(string)
		}

		// Write updated values to the database
		if err := db.Set.CNAME(recordName, record.Target); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "MX":
		// Get original record from database
		record := db.Get.MX(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"host", "priority"}, map[string]map[string]string{"host": {"type": "string", "required": "false"}, "priority": {"type": "uint16", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in the body
		if valid["host"] {
			record.Host = body["host"].(string)
		}
		if valid["priority"] {
			record.Priority = uint16(body["priority"].(float64))
		}

		// Write updated values to the database
		if err := db.Set.MX(recordName, record.Priority, record.Host); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "LOC":
		// Get original record from database
		record := db.Get.LOC(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"version", "size", "horizontal-precision", "vertical-precision", "altitude", "lat-degrees", "lat-minutes", "lat-seconds", "lat-direction", "long-degrees", "long-minutes", "long-seconds", "long-direction"}, map[string]map[string]string{
			"version": {"type": "uint8", "required": "true"},
			"size": {"type": "uint8", "required": "true"},
			"horizontal-precision": {"type": "uint8", "required": "true"},
			"vertical-precision": {"type": "uint8", "required": "true"},
			"altitude": {"type": "uint32", "required": "true"},
			"lat-degrees": {"type": "uint8", "required": "true", "min": "0", "max": "90"},
			"lat-minutes": {"type": "uint8", "required": "true", "min": "0", "max": "60"},
			"lat-seconds": {"type": "uint8", "required": "true", "min": "0", "max": "60"},
			"lat-direction": {"type": "string", "required": "true", "oneOf": "N,S"},
			"long-degrees": {"type": "uint8", "required": "true", "min": "0", "max": "180"},
			"long-minutes": {"type": "uint8", "required": "true", "min": "0", "max": "60"},
			"long-seconds": {"type": "uint8", "required": "true", "min": "0", "max": "60"},
			"long-direction": {"type": "string", "required": "true", "oneOf": "E,W"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in body
		if valid["version"] {
			record.Version = uint8(body["version"].(float64))
		}
		if valid["size"] {
			record.Size = uint8(body["size"].(float64))
		}
		if valid["horizontal-precision"] {
			record.HorizontalPrecision = uint8(body["horizontal-precision"].(float64))
		}
		if valid["vertical-precision"] {
			record.VerticalPrecision = uint8(body["vertical-precision"].(float64))
		}
		if valid["altitude"] {
			record.Altitude = uint32(body["altitude"].(float64))
		}
		if valid["lat-degrees"] {
			record.LatDegrees = uint8(body["lat-degrees"].(float64))
		}
		if valid["lat-minutes"] {
			record.LatMinutes = uint8(body["lat-minutes"].(float64))
		}
		if valid["lat-seconds"] {
			record.LatSeconds = uint8(body["lat-seconds"].(float64))
		}
		if valid["lat-direction"] {
			record.LatDirection = body["lat-direction"].(string)
		}
		if valid["long-degrees"] {
			record.LongDegrees = uint8(body["long-degrees"].(float64))
		}
		if valid["long-minutes"] {
			record.LongMinutes = uint8(body["long-minutes"].(float64))
		}
		if valid["long-seconds"] {
			record.LongSeconds = uint8(body["long-seconds"].(float64))
		}
		if valid["long-direction"] {
			record.LongDirection = body["long-direction"].(string)
		}

		// Write updated values to database
		if err := db.Set.LOC(recordName, record.Version, record.Size, record.HorizontalPrecision, record.VerticalPrecision, record.Altitude, record.LatDegrees, record.LatMinutes, record.LatSeconds, record.LatDirection, record.LongDegrees, record.LongMinutes, record.LongSeconds, record.LongDirection); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "SRV":
		// Get original record from database
		record := db.Get.SRV(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"priority", "weight", "port", "target"}, map[string]map[string]string{
			"priority": {"type": "uint16", "required": "false"},
			"weight": {"type": "uint16", "required": "false"},
			"port": {"type": "uint16", "required": "false"},
			"target": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in body
		if valid["priority"] {
			record.Priority = uint16(body["priority"].(float64))
		}
		if valid["weight"] {
			record.Weight = uint16(body["weight"].(float64))
		}
		if valid["port"] {
			record.Port = uint16(body["port"].(float64))
		}
		if valid["target"] {
			record.Target = body["target"].(string)
		}

		// Write updated values to database
		if err := db.Set.SRV(recordName, record.Priority, record.Weight, record.Port, record.Target); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "SPF":
		// Get original record from database
		record := db.Get.SPF(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"text"}, map[string]map[string]string{"text": {"type": "stringarray", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in body
		if valid["text"] {
			text, _ := util.ConvertArrayToString(body["text"].([]interface{}))
			record.Text = text
		}

		// Write updated values to database
		if err := db.Set.SPF(recordName, record.Text); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "TXT":
		// Get original record from database
		record := db.Get.TXT(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"text"}, map[string]map[string]string{"text": {"type": "stringarray", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in body
		if valid["text"] {
			text, _ := util.ConvertArrayToString(body["text"].([]interface{}))
			record.Text = text
		}

		// Write updated values to database
		if err := db.Set.TXT(recordName, record.Text); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "NS":
		// Get original record from database
		record := db.Get.NS(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"nameserver"}, map[string]map[string]string{"nameserver": {"type": "string", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
			return
		}

		// Update values if they exist in body
		if valid["nameserver"] {
			record.Nameserver = body["nameserver"].(string)
		}

		// Write updated values to database
		if err := db.Set.NS(recordName, record.Nameserver); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+ err.Error())
			return
		}

	case "CAA":
		// Get original record from database
		record := db.Get.CAA(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"tag", "content"}, map[string]map[string]string{"tag": {"type": "string", "required": "false"}, "content": {"type": "string", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["tag"] {
			record.Tag = body["tag"].(string)
		}
		if valid["content"] {
			record.Tag = body["content"].(string)
		}

		// Write updated values to database
		if err := db.Set.CAA(recordName, record.Tag, record.Content); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+ err.Error())
			return
		}

	case "PTR":
		// Get original record from database
		record := db.Get.PTR(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"domain"}, map[string]map[string]string{"domain": {"type": "string", "required": "false"}})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["domain"] {
			record.Domain = body["domain"].(string)
		}

		// Write updated values to database
		if err := db.Set.PTR(recordName, record.Domain); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+ err.Error())
			return
		}

	case "CERT":
		// Get original record from database
		record := db.Get.CERT(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"c-type", "key-tag", "algorithm", "certificate"}, map[string]map[string]string{
			"c-type": {"type": "uint16", "requried": "false"},
			"key-tag": {"type": "uint16", "required": "false"},
			"algorithm": {"type": "uint8", "required": "false"},
			"certificate": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["c-type"] {
			record.Type = uint16(body["c-type"].(float64))
		}
		if valid["key-tag"] {
			record.KeyTag = uint16(body["key-tag"].(float64))
		}
		if valid["algorithm"] {
			record.Algorithm = uint8(body["algorithm"].(float64))
		}
		if valid["certificate"] {
			record.Certificate = body["certificate"].(string)
		}

		// Write updated values to database
		if err := db.Set.CERT(recordName, record.Type, record.KeyTag, record.Algorithm, record.Certificate); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "DNSKEY":
		// Get original record from database
		record := db.Get.DNSKEY(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"flags", "protocol", "algorithm", "public-key"}, map[string]map[string]string{
			"flags": {"type": "uint16", "required": "false"},
			"protocol": {"type": "uint8", "required": "false"},
			"algorithm": {"type": "uint8", "required": "false"},
			"public-key": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["flags"] {
			record.Flags = uint16(body["flags"].(float64))
		}
		if valid["protocol"] {
			record.Protocol = uint8(body["protocol"].(float64))
		}
		if valid["algorithm"] {
			record.Algorithm = uint8(body["algorithm"].(float64))
		}
		if valid["public-key"] {
			record.PublicKey = body["public-key"].(string)
		}

		// Write updated values to database
		if err := db.Set.DNSKEY(recordName, record.Flags, record.Protocol, record.Algorithm, record.PublicKey); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "DS":
		// Get original record from database
		record := db.Get.DS(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"key-tag", "algorithm", "digest-type", "digest"}, map[string]map[string]string{
			"key-tag": {"type": "uint16", "required": "false"},
			"algorithm": {"type": "uint8", "required": "false"},
			"digest-type": {"type": "uint8", "required": "false"},
			"digest": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["key-tag"] {
			record.KeyTag = uint16(body["key-tag"].(float64))
		}
		if valid["algorithm"] {
			record.Algorithm = uint8(body["algorithm"].(float64))
		}
		if valid["digest-type"] {
			record.DigestType = uint8(body["digest-type"].(float64))
		}
		if valid["digest"] {
			record.Digest = body["digest"].(string)
		}

		// Write updated values to database
		if err := db.Set.DS(recordName, record.KeyTag, record.Algorithm, record.DigestType, record.Digest); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "NAPTR":
		// Get original record from database
		record := db.Get.NAPTR(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"order", "preference", "flags", "service", "regexp", "replacement"}, map[string]map[string]string{
			"order": {"type": "uint16", "required": "false"},
			"preference": {"type": "uint16", "required": "false"},
			"flags": {"type": "string", "required": "false"},
			"service": {"type": "string", "required": "false"},
			"regexp": {"type": "string", "required": "false"},
			"replacement": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["order"] {
			record.Order = uint16(body["order"].(float64))
		}
		if valid["preference"] {
			record.Preference = uint16(body["preference"].(float64))
		}
		if valid["flags"] {
			record.Flags = body["flags"].(string)
		}
		if valid["service"] {
			record.Service = body["service"].(string)
		}
		if valid["regexp"] {
			record.Regexp = body["regexp"].(string)
		}
		if valid["replacement"] {
			record.Replacement = body["replacement"].(string)
		}

		// Write updated values to database
		if err := db.Set.NAPTR(recordName, record.Order, record.Preference, record.Flags, record.Service, record.Regexp, record.Replacement); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "SMIMEA":
		// Get original record from database
		record := db.Get.SMIMEA(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"usage", "selector", "matching-type", "certificate"}, map[string]map[string]string{
			"usage": {"type": "uint8", "required": "false"},
			"selector": {"type": "uint8", "required": "false"},
			"matching-type": {"type": "uint8", "required": "false"},
			"certificate": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["usage"] {
			record.Usage = uint8(body["usage"].(float64))
		}
		if valid["selector"] {
			record.Selector = uint8(body["selector"].(float64))
		}
		if valid["matching-type"] {
			record.MatchingType = uint8(body["matching-type"].(float64))
		}
		if valid["certificate"] {
			record.Certificate = body["certificate"].(string)
		}

		// Write updated values to database
		if err := db.Set.SMIMEA(recordName, record.Usage, record.Selector, record.MatchingType, record.Certificate); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "SSHFP":
		// Get original record from database
		record := db.Get.SSHFP(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"algorithm", "s-type", "fingerprint"}, map[string]map[string]string{
			"algorithm": {"type": "uint8", "required": "false"},
			"s-type": {"type": "uint8", "required": "false"},
			"fingerprint": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["algorithm"] {
			record.Algorithm = uint8(body["algorithm"].(float64))
		}
		if valid["s-type"] {
			record.Type = uint8(body["s-type"].(float64))
		}
		if valid["fingerprint"] {
			record.Fingerprint = body["fingerprint"].(string)
		}

		// Write updated values to database
		if err := db.Set.SSHFP(recordName, record.Algorithm, record.Type, record.Fingerprint); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "TLSA":
		// Get original record from database
		record := db.Get.TLSA(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"usage", "selector", "matching-type", "certificate"}, map[string]map[string]string{
			"usage": {"type": "uint8", "required": "false"},
			"selector": {"type": "uint8", "required": "false"},
			"matching-type": {"type": "uint8", "required": "false"},
			"certificate": {"type": "string", "required": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["usage"] {
			record.Usage = uint8(body["usage"].(float64))
		}
		if valid["selector"] {
			record.Selector = uint8(body["selector"].(float64))
		}
		if valid["matching-type"] {
			record.MatchingType = uint8(body["matching-type"].(float64))
		}
		if valid["certificate"] {
			record.Certificate = body["certificate"].(string)
		}

		// Write updated values to database
		if err := db.Set.TLSA(recordName, record.Usage, record.Selector, record.MatchingType, record.Certificate); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}

	case "URI":
		// Get original record from database
		record := db.Get.URI(recordName + ".")
		if util.RecordDoesNotExist(record) {
			util.Responses.Error(w, http.StatusBadRequest, "specified record does not exist")
			return
		}

		// Get valid values in body
		err, valid := util.ValidateBody(body, []string{"priority", "weight", "target"}, map[string]map[string]string{
			"priority": {"type": "uint16", "required": "false"},
			"weight": {"type": "uint16", "required": "false"},
			"target": {"type": "string", "requried": "false"},
		})
		if err != "" {
			util.Responses.Error(w, http.StatusBadRequest, err)
		}

		// Update values if they exist in body
		if valid["priority"] {
			record.Priority = uint16(body["priority"].(float64))
		}
		if valid["weight"] {
			record.Weight = uint16(body["weight"].(float64))
		}
		if valid["target"] {
			record.Target = body["target"].(string)
		}

		// Write updated values to database
		if err := db.Set.URI(recordName, record.Priority, record.Weight, record.Target); err != nil {
			util.Responses.Error(w, http.StatusInternalServerError, "failed to write record to database: "+err.Error())
			return
		}
	default:
		util.Responses.Error(w, http.StatusBadRequest, "field 'type' must be on of: A, AAAA, CNAME, MX, LOC, SRV, SPF, TXT, NS, CAA, PTR, CERT, DNSKEY, DS, NAPTR, SMIMEA, SSHFP, TLSA, URI")
		return
	}

	util.Responses.Success(w)
}
