package manufacturers

import (
	"encoding/json"

	log "github.com/mohclips/BLEAS2/internal/logging"
)

// MarshalOrEmpty returns json.Marshal(v) as raw bytes, or []byte(`{}`) on error.
// Used by manufacturer parsers that don't want to surface marshal errors past
// the parse boundary.
func MarshalOrEmpty(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Error("marshal: %s", err)
		return []byte("{}")
	}
	return b
}
