package apple

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseMF_iBeacon(t *testing.T) {
	cases := [][]byte{
		{76, 0, 2, 21, 226, 197, 109, 181, 223, 251, 72, 210, 176, 96, 208, 245, 167, 16, 150, 224, 0, 0, 0, 0, 197},
		{76, 0, 2, 21, 80, 118, 92, 183, 217, 234, 78, 33, 153, 164, 250, 135, 150, 19, 164, 146, 141, 154, 239, 255, 255},
	}
	for i, data := range cases {
		got := ParseMF(data)
		if !strings.Contains(got, `"ibeacon"`) {
			t.Errorf("case %d: parsed JSON missing ibeacon block: %s", i, got)
		}
	}
}

func TestParseMF_ShortPacket_NoPanic(t *testing.T) {
	got := ParseMF([]byte{76, 0, 2, 21}) // claims length 21 but has no payload
	if got != "" {
		t.Errorf("expected empty result for truncated packet, got %q", got)
	}
}

// TestParseMF_FindMy_Nearby covers the 2-byte short variant we actually
// observed in the wild (status=0x00, public_key_bits=0x03).
func TestParseMF_FindMy_Nearby(t *testing.T) {
	// [vendor_id_lo, vendor_id_hi, action=0x12, length=0x02, status, pkbits]
	data := []byte{0x4c, 0x00, 0x12, 0x02, 0x00, 0x03}
	got := ParseMF(data)
	var v struct {
		Apple struct {
			FindMy struct {
				Variant       string `json:"variant"`
				PublicKeyBits int    `json:"public_key_bits"`
				Maintained    bool   `json:"maintained"`
			} `json:"findmy"`
		} `json:"apple"`
	}
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Fatalf("invalid JSON: %v (%s)", err, got)
	}
	if v.Apple.FindMy.Variant != "nearby" {
		t.Errorf("variant: got %q want nearby", v.Apple.FindMy.Variant)
	}
	if v.Apple.FindMy.PublicKeyBits != 3 {
		t.Errorf("public_key_bits: got %d want 3", v.Apple.FindMy.PublicKeyBits)
	}
	if v.Apple.FindMy.Maintained {
		t.Errorf("maintained: got true want false")
	}
}

// TestParseMF_FindMy_Separated covers the 25-byte lost-mode variant.
func TestParseMF_FindMy_Separated(t *testing.T) {
	pubKey := make([]byte, 22)
	for i := range pubKey {
		pubKey[i] = byte(i + 1)
	}
	data := []byte{0x4c, 0x00, 0x12, 0x19, 0x04} // status: maintained bit set
	data = append(data, pubKey...)
	data = append(data, 0x02, 0xAB) // pkbits, hint
	got := ParseMF(data)
	if !strings.Contains(got, `"variant":"separated"`) {
		t.Errorf("missing separated variant in %s", got)
	}
	if !strings.Contains(got, `"maintained":true`) {
		t.Errorf("maintained flag not set in %s", got)
	}
	if !strings.Contains(got, `"hint":171`) {
		t.Errorf("hint not 0xAB(171) in %s", got)
	}
}

// TestParseMF_MultiAction: a single Apple manufacturer block can carry
// multiple Continuity subtypes concatenated. Verify both surface.
func TestParseMF_MultiAction(t *testing.T) {
	// [vendor lo, vendor hi, findmy(0x12) length 2 status pkbits, nearby(0x10) length 2 state mask]
	data := []byte{0x4c, 0x00, 0x12, 0x02, 0x00, 0x03, 0x10, 0x02, 0x03, 0x18}
	got := ParseMF(data)
	if !strings.Contains(got, `"findmy"`) {
		t.Errorf("findmy missing from %s", got)
	}
	if !strings.Contains(got, `"nearby"`) {
		t.Errorf("nearby missing from %s", got)
	}
}

// TestParseMF_UnknownSubtype: undocumented subtypes should appear under a
// stable unknown_0xNN key, preserving the raw bytes.
func TestParseMF_UnknownSubtype(t *testing.T) {
	data := []byte{0x4c, 0x00, 0x16, 0x0a, 4, 3, 66, 85, 166, 182, 254, 206, 109, 145}
	got := ParseMF(data)
	if !strings.Contains(got, `"unknown_0x16"`) {
		t.Errorf("unknown_0x16 missing from %s", got)
	}
	if !strings.Contains(got, `"action":22`) {
		t.Errorf("action=22 missing from %s", got)
	}
}

// TestParseMF_Coverage spot-checks each newly-added parser emits its JSON key.
func TestParseMF_Coverage(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		key  string
	}{
		{"nearby_action", []byte{0x4c, 0x00, 0x0f, 0x05, 0xc0, 0x08, 0x01, 0x02, 0x03}, `"nearby_action"`},
		{"airdrop", append([]byte{0x4c, 0x00, 0x05, 0x12}, make([]byte, 18)...), `"airdrop"`},
		{"homekit", append([]byte{0x4c, 0x00, 0x06, 0x0d}, make([]byte, 13)...), `"homekit"`},
		{"hey_siri", append([]byte{0x4c, 0x00, 0x08, 0x07}, make([]byte, 7)...), `"hey_siri"`},
		{"airplay_target", []byte{0x4c, 0x00, 0x09, 0x06, 0x00, 0x00, 192, 168, 1, 1}, `"airplay_target"`},
		{"airplay_source", []byte{0x4c, 0x00, 0x0a, 0x01, 0x00}, `"airplay_source"`},
		{"magic_switch", []byte{0x4c, 0x00, 0x0b, 0x03, 0x00, 0x00, 0x3F}, `"magic_switch"`},
		{"tethering_target", []byte{0x4c, 0x00, 0x0d, 0x04, 0xde, 0xad, 0xbe, 0xef}, `"tethering_target"`},
		{"tethering_source", []byte{0x4c, 0x00, 0x0e, 0x06, 0x01, 0x00, 75, 0x07, 0x00, 4}, `"tethering_source"`},
		{"airprint", append([]byte{0x4c, 0x00, 0x03, 0x16}, make([]byte, 22)...), `"airprint"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseMF(tc.data)
			if !strings.Contains(got, tc.key) {
				t.Errorf("%s missing in: %s", tc.key, got)
			}
		})
	}
}
