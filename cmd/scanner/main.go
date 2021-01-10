package main

// https://towardsdatascience.com/spelunking-bluetooth-le-with-go-c2cff65a7aca
// https://gist.github.com/sausheong/16afa3c4018a22a737f08416768cab90#file-adscanhandler-go

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/sausheong/ble"
	"github.com/sausheong/ble/examples/lib/dev"

	// "github.com/gernest/wow"

	"github.com/mohclips/BLEAS2/internal/manufacturers"
	"github.com/mohclips/BLEAS2/internal/utils"
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
	if !utils.RunningAsRoot() {
		log.Fatal("This program must be run as root! (sudo)")
	}

	// make sure USB power is 'on' not 'auto' otherwise we have s suspend issue
	utils.CheckBtUsbPower()

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

		//binary.LittleEndian.Uint16(b[0:])
		//m := binary.LittleEndian.Uint16((a.ManufacturerData())[0:2])

		mID := manufacturers.GetID(a.ManufacturerData())
		mName := manufacturers.GetName(mID)

		fmt.Printf(" Manufacturer Id: 0x%04x %s\n", mID, mName)
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
