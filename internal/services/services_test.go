package services

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestName_Known(t *testing.T) {
	for uuid, want := range map[string]string{
		"180f": "Battery",
		"180d": "Heart Rate",
		"feaa": "Eddystone (Google)",
		"fef3": "Google Fast Pair",
		"fd6f": "Google Exposure Notification",
	} {
		if got := Name(uuid); got != want {
			t.Errorf("Name(%q)=%q want %q", uuid, got, want)
		}
	}
}

func TestName_Unknown(t *testing.T) {
	if got := Name("dead"); got != "" {
		t.Errorf("Name(dead) = %q, want empty", got)
	}
}

func TestParse_Eddystone_URL(t *testing.T) {
	// frame_type=URL(0x10), tx_power=-50(0xCE), scheme=https://, body="google" + .com(0x07)
	data := []byte{0x10, 0xCE, 0x01, 'g', 'o', 'o', 'g', 'l', 'e', 0x07}
	name, body := Parse("feaa", data)
	if name != "eddystone" {
		t.Errorf("name=%q want eddystone", name)
	}
	var v map[string]interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if v["frame_type"] != "url" {
		t.Errorf("frame_type=%v want url", v["frame_type"])
	}
	if got, want := v["url"], "https://www.google.com"; got != want {
		t.Errorf("url=%v want %v", got, want)
	}
}

func TestParse_FastPair_Discoverable(t *testing.T) {
	// 24-bit model ID 0x123456
	name, body := Parse("fef3", []byte{0x12, 0x34, 0x56})
	if name != "fast_pair" {
		t.Errorf("name=%q want fast_pair", name)
	}
	if !strings.Contains(string(body), `"model_id":"123456"`) {
		t.Errorf("model_id missing in %s", body)
	}
	if !strings.Contains(string(body), `"mode":"discoverable"`) {
		t.Errorf("mode missing in %s", body)
	}
}

func TestParse_ExposureNotification(t *testing.T) {
	data := make([]byte, 20)
	for i := range data {
		data[i] = byte(i)
	}
	name, body := Parse("fd6f", data)
	if name != "exposure_notification" {
		t.Errorf("name=%q want exposure_notification", name)
	}
	if !strings.Contains(string(body), `"rpi":"000102030405060708090a0b0c0d0e0f"`) {
		t.Errorf("rpi missing/wrong in %s", body)
	}
}

func TestParse_Unknown_PreservesBytes(t *testing.T) {
	name, body := Parse("aabb", []byte{0xde, 0xad, 0xbe, 0xef})
	if name != "service_aabb" {
		t.Errorf("name=%q want service_aabb", name)
	}
	if !strings.Contains(string(body), `"uuid":"aabb"`) {
		t.Errorf("uuid missing in %s", body)
	}
}
