package utils

import (
	"fmt"
	"io/ioutil"

	//"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/mohclips/BLEAS2/internal/logging"
)

// #######################################################################################

// RunningAsRoot - check if running as root
func RunningAsRoot() bool {
	// https://www.socketloop.com/tutorials/golang-force-your-program-to-run-with-root-permissions
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		log.Fatal("%s", err)
	}

	// output has trailing \n
	// need to remove the \n
	// otherwise it will cause error for strconv.Atoi
	// log.Println(output[:len(output)-1])

	// 0 = root, 501 = non-root user
	i, err := strconv.Atoi(string(output[:len(output)-1]))

	if err != nil {
		log.Fatal("%s", err)
	}

	if i != 0 {
		return false
	}

	return true
}

// #######################################################################################

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
	err := ioutil.WriteFile(p, data, 644)
	if err != nil {
		panic(err)
	}
}
func findDevPower(p string) {

	splitPath := strings.Split(p, "/")

	//fmt.Printf("%s\n", splitPath)

	// path position to start in
	var hubStart bool = false
	var startPos int = 0
	for i := range splitPath {
		if strings.Contains(splitPath[i], "usb") {
			hubStart = true
			startPos = i + 1
		}
		if hubStart && (i > startPos) {
			currentPath := strings.Join(splitPath[:i], "/")
			controlFile := currentPath + "/power/control"
			if exists(controlFile) {
				// if debug {
				// 	fmt.Printf("found: %s\n", controlFile)
				// }

				data, _ := ioutil.ReadFile(controlFile)
				//fmt.Print(string(data))

				controlValue := strings.TrimSpace(string(data))
				if controlValue != "on" {
					// if debug {
					// 	fmt.Printf("Warning: USB device not set correctly [Power=%s]\n", controlValue)
					// }
					setUsbDevPower(controlFile, "on")
				}

			}
		}
	}

}

// CheckBtUsbPower - check that USB port BT device is on is not allowed to suspend power
func CheckBtUsbPower() {
	const bluetoothPath = "/sys/class/bluetooth/"

	// get all hci device names
	files, err := ioutil.ReadDir(bluetoothPath)
	if err != nil {
		log.Fatal("No HCI devices found, odd: ", err)
	}

	// get paths to each device
	for _, f := range files {
		//fmt.Println(f.Name())
		btDev, _ := os.Readlink(bluetoothPath + f.Name())
		//fmt.Println(btDev)

		absBtDev, _ := filepath.Abs(bluetoothPath + btDev)
		//fmt.Println(absBtDev)

		if exists(absBtDev) {
			// if debug {
			// 	fmt.Printf("Device path exists: %s\n", absBtDev)
			// }

			findDevPower(absBtDev)
		}

	}
}

// #######################################################################################

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

// #######################################################################################

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
