package handler

import (
	"encoding/json"
	"net/http"
)

// decodeJSON decodes the JSON body of r into dst.
// It is a thin wrapper around json.NewDecoder that disallows unknown fields,
// which would otherwise be silently ignored — a common source of subtle bugs
// when a client sends a misspelled field and wonders why it has no effect.
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
