package main

// https://towardsdatascience.com/spelunking-bluetooth-le-with-go-c2cff65a7aca
// https://gist.github.com/sausheong/16afa3c4018a22a737f08416768cab90#file-adscanhandler-go

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sausheong/ble"
	"github.com/sausheong/ble/examples/lib/dev"
	// "github.com/gernest/wow"
)

var (
	device = flag.String("device", "default", "implementation of ble")
	du     = flag.Duration("du", 5*time.Second, "scanning duration")
	dup    = flag.Bool("dup", true, "allow duplicate reported")
)

var isRunning bool = false

const debug bool = false

func main() {

	// needed to access Bluetooth and set USB power
	if !runningAsRoot() {
		log.Fatal("This program must be run as root! (sudo)")
	}

	// make sure USB power is 'on' not 'auto' otherwise we have s suspend issue
	checkBtUsbPower()

	flag.Parse()

	fmt.Printf("Running...\n")

	d, err := dev.NewDevice(*device)
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	// best struct pprint
	//spew.Dump(d)

	isRunning = true

	// Scan for specified duration, AND until interrupted by user.
	for isRunning {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
		chkErr(ble.Scan(ctx, *dup, advHandler, nil))
	}
}

// #######################################################################################

func advHandler(a ble.Advertisement) {

	if (len(a.LocalName()) > 0) &&
		(a.LocalName() == "ID152") &&
		(len(a.ManufacturerData()) == 0) {
		return
	}

	if len(a.Services()) > 0 {
		if a.Services()[0].String() == "fd6f" {
			return
		}
	}

	//debug
	// if len(a.Services()) > 0 {
	// 	if a.Services()[0].String() != "fd6f" {
	// 		fmt.Printf("\n\n%s", spew.Sdump(a))
	// 	}
	// }
	// if len(a.ManufacturerData()) > 0 {
	// 	spew.Dump(a)
	// }
	//spew.Printf("\n\nSPEW %#v", a)
	//fmt.Printf("\n\nFMT %#v", a)

	fmt.Printf("%s ", time.Now().Format(time.RFC3339))

	if a.Connectable() {
		fmt.Printf("[%s] C %3d:", a.Addr(), a.RSSI())
	} else {
		fmt.Printf("[%s] N %3d:", a.Addr(), a.RSSI())
	}
	comma := ""
	if len(a.LocalName()) > 0 {
		fmt.Printf(" Name: %s", a.LocalName())
		comma = ","
	}
	if len(a.Services()) > 0 {
		fmt.Printf("%s Svcs: %v", comma, a.Services())
		comma = ","
	}
	if len(a.ManufacturerData()) > 0 {
		fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
	}

	fmt.Printf("\n%s RAW: %s", comma, formatHex(hex.EncodeToString(a.LEAdvertisingReportRaw())))

	fmt.Printf("\n%s SR: %s", comma, formatHex(hex.EncodeToString(a.ScanResponseRaw())))

	fmt.Printf("\n\n")
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("%s ", time.Now().Format(time.RFC3339))
		fmt.Printf("---\n") // no data seen?
	case context.Canceled:
		fmt.Printf("cancelled\n")
		isRunning = false
	default:
		log.Fatalf(err.Error())
	}
}

// reformat string for proper display of hex
func formatHex(instr string) (outstr string) {
	outstr = ""
	for i := range instr {
		if i%2 == 0 {
			outstr += instr[i:i+2] + " "
		}
	}
	return
}

// #######################################################################################

func runningAsRoot() bool {
	// https://www.socketloop.com/tutorials/golang-force-your-program-to-run-with-root-permissions
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	// output has trailing \n
	// need to remove the \n
	// otherwise it will cause error for strconv.Atoi
	// log.Println(output[:len(output)-1])

	// 0 = root, 501 = non-root user
	i, err := strconv.Atoi(string(output[:len(output)-1]))

	if err != nil {
		log.Fatal(err)
	}

	if i != 0 {
		return false
	}

	return true
}

// #######################################################################################

// Exists reports whether the named file or directory exists.
// https://stackoverflow.com/a/12527546/7396553
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func setUsbDevPower(p string, power string) {
	fmt.Printf("Setting %s to %q\n", p, power)

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
			if Exists(controlFile) {
				if debug {
					fmt.Printf("found: %s\n", controlFile)
				}

				data, _ := ioutil.ReadFile(controlFile)
				//fmt.Print(string(data))

				controlValue := strings.TrimSpace(string(data))
				if controlValue != "on" {
					if debug {
						fmt.Printf("Warning: USB device not set correctly [Power=%s]\n", controlValue)
					}
					setUsbDevPower(controlFile, "on")
				}

			}
		}
	}

}

func checkBtUsbPower() {
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

		if Exists(absBtDev) {
			if debug {
				fmt.Printf("Device path exists: %s\n", absBtDev)
			}

			findDevPower(absBtDev)
		}

	}
}

// #######################################################################################
