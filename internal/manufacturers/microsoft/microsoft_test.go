package microsoft

import (
	"encoding/json"
	"testing"
)

func TestParseMF_ValidPackets(t *testing.T) {
	// Real captures from a Microsoft 0x0006 manufacturer advertisement.
	cases := [][]byte{
		{6, 0, 1, 9, 32, 2, 100, 249, 61, 127, 54, 229, 151, 142, 235, 115, 40, 106, 108, 208, 176, 132, 121, 119, 51, 228, 127, 15, 12},
		{6, 0, 1, 9, 32, 2, 109, 117, 23, 172, 181, 200, 183, 22, 7, 230, 58, 127, 90, 197, 25, 34, 246, 177, 114, 84, 191, 35, 241},
		{6, 0, 1, 9, 32, 2, 27, 0, 48, 153, 31, 185, 154, 70, 216, 81, 19, 90, 46, 45, 59, 60, 82, 104, 36, 76, 0, 104, 62},
	}
	for i, data := range cases {
		got := ParseMF(data)
		// got is `{"microsoft": <object>}` — verify it parses as JSON.
		var v map[string]json.RawMessage
		if err := json.Unmarshal([]byte(got), &v); err != nil {
			t.Errorf("case %d: result is not valid JSON: %v (%q)", i, err, got)
		}
		if _, ok := v["microsoft"]; !ok {
			t.Errorf("case %d: missing microsoft key in %q", i, got)
		}
	}
}
