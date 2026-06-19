package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/mohclips/BLEAS2/internal/logging"
)

// RunningAsRoot reports whether the process is running with uid 0.
func RunningAsRoot() bool {
	return os.Geteuid() == 0
}

// Exists reports whether the named file or directory exists.
// https://stackoverflow.com/a/12527546/7396553
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func setUsbDevPower(p string, power string) {
	log.Info("Setting %s to %q\n", p, power)

	data := []byte(power + "\n")
	err := os.WriteFile(p, data, 0644)
	if err != nil {
		panic(err)
	}
}
func findDevPower(p string) {
	splitPath := strings.Split(p, "/")

	hubStart := false
	startPos := 0
	for i := range splitPath {
		if strings.Contains(splitPath[i], "usb") {
			hubStart = true
			startPos = i + 1
		}
		if hubStart && i > startPos {
			currentPath := strings.Join(splitPath[:i], "/")
			controlFile := currentPath + "/power/control"
			if !exists(controlFile) {
				continue
			}
			data, _ := os.ReadFile(controlFile)
			if strings.TrimSpace(string(data)) != "on" {
				setUsbDevPower(controlFile, "on")
			}
		}
	}
}

// CheckBtUsbPower - check that USB port BT device is on is not allowed to suspend power
func CheckBtUsbPower() {
	const bluetoothPath = "/sys/class/bluetooth/"

	// get all hci device names
	files, err := os.ReadDir(bluetoothPath)
	if err != nil {
		log.Fatal("No HCI devices found, odd: ", err)
	}

	// get paths to each device
	for _, f := range files {
		btDev, _ := os.Readlink(bluetoothPath + f.Name())
		absBtDev, _ := filepath.Abs(bluetoothPath + btDev)

		if exists(absBtDev) {
			findDevPower(absBtDev)
		}
	}
}

// FormatHex - reformat string for proper display of hex
func FormatHex(instr string) (outstr string) {
	outstr = ""
	for i := range instr {
		if i%2 == 0 {
			outstr += instr[i:i+2] + " "
		}
	}
	// last := len(outstr) - 1
	// outstr = outstr[:last]
	return
}

// FormatHexComma - reformat string for proper display of hex
func FormatHexComma(instr string) (outstr string) {
	outstr = ""
	if len(instr) == 0 {
		return
	}
	for i := range instr {
		if i%2 == 0 {

			hex := instr[i : i+2]
			outstr += "0x" + hex + ", "
		}
	}
	last := len(outstr) - 2
	outstr = outstr[:last]
	return
}

// FormatDecComma - reformat string for proper display of dec
func FormatDecComma(instr string) (outstr string) {
	outstr = ""
	if len(instr) == 0 {
		return
	}
	for i := range instr {
		if i%2 == 0 {

			value, _ := strconv.ParseInt(instr[i:i+2], 16, 64)
			outstr += fmt.Sprintf("%d", value) + ", "
		}
	}
	last := len(outstr) - 2
	outstr = outstr[:last]
	return
}

// BitmaskToNames - return an array of strings from a map dependant on bitmask input
func BitmaskToNames(k int, m map[int]string) []string {
	var result []string
	if k == 0 {
		result = append(result, m[k])
	} else {
		for i := 0; i < len(m); i++ {
			pos := k & (1 << i)
			if pos != 0 {
				result = append(result, m[pos])
			}
		}
	}
	return result
}
