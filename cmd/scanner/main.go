package main

// https://towardsdatascience.com/spelunking-bluetooth-le-with-go-c2cff65a7aca
// https://gist.github.com/sausheong/16afa3c4018a22a737f08416768cab90#file-adscanhandler-go

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"time"

	// BLE
	"github.com/pkg/errors"
	"github.com/sausheong/ble"
	"github.com/sausheong/ble/examples/lib/dev"

	// Logging
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"

	// My stuff
	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
	apple "github.com/mohclips/BLEAS2/internal/manufacturers/apple"
	"github.com/mohclips/BLEAS2/internal/utils"
)

var (
	device = flag.String("device", "default", "implementation of ble")
	du     = flag.Duration("du", 5*time.Second, "scanning duration")
	dup    = flag.Bool("dup", true, "allow duplicate reported")
	id152  = flag.Bool("id152", false, "show ID152 devices")
	nhs    = flag.Bool("nhs", false, "show NHS advertisements")
)

var loglevel = logrus.TraceLevel

var isRunning bool = false

const debug bool = false

var log = logrus.New()

// ParsedManufacturerData - The parsed data we are after
type ParsedManufacturerData struct {
	ID   int
	Name string
	// we add more here once its Marshalled as json
}

// Device - represents a BLE device, with our parsed data tacked on
type Device struct {
	Address          string    `json:"address"`
	Detected         time.Time `json:"detected"`
	Since            string    `json:"since"`
	Name             string    `json:"name"`
	RSSI             int       `json:"rssi"`
	Advertisement    string    `json:"advertisement"`
	ScanResponse     string    `json:"scanresponse"`
	ManufacturerData ParsedManufacturerData
}

func main() {

	log.SetFormatter(&nested.Formatter{
		HideKeys: false,
		//FieldsOrder: []string{"component", "category"},
		TimestampFormat: "2006-01-02 15:04:06",
		ShowFullLevel:   true,
	})

	// log to stdout not stderr
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(loglevel)

	// needed to access Bluetooth and set USB power
	if !utils.RunningAsRoot() {
		log.Fatal("This program must be run as root! (sudo)")
	}

	// make sure USB power is 'on' not 'auto' otherwise we have a suspend issue
	utils.CheckBtUsbPower()

	flag.Parse()

	log.Info("Running...")

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
	/*
	   	t := reflect.TypeOf(struct{ ble.Advertisement }{})
	   	for i := 0; i < t.NumMethod(); i++ {
	   		fmt.Println(t.Method(i).Name)
	   	}
	   Addr
	   Connectable
	   LEAdvertisingReportRaw
	   LocalName
	   ManufacturerData
	   OverflowService
	   RSSI
	   ScanResponseRaw
	   ServiceData
	   Services
	   SolicitedService
	   TxPowerLevel
	*/

	// very chatty device at home
	if a.LocalName() == "ID152" && *id152 == false {
		return
	}

	// only display pkts with MD in them
	if (len(a.LocalName()) > 0) &&
		(a.LocalName() == "ID152") &&
		(len(a.ManufacturerData()) == 0) {
		return
	}

	// do not display NHS app
	if len(a.Services()) > 0 {
		if a.Services()[0].String() == "fd6f" && *nhs == false {
			return
		}
	}

	log.WithFields(logrus.Fields{
		"time": time.Now().Format(time.RFC3339),
		"addr": a.Addr(),
		"rssi": a.RSSI(),
	}).Info()

	if len(a.LocalName()) > 0 {
		log.WithFields(logrus.Fields{
			"Name": a.LocalName(),
		}).Debug()
	}
	if len(a.Services()) > 0 {
		log.WithFields(logrus.Fields{
			"Services": a.Services(),
		}).Debug()
	}

	if len(a.ManufacturerData()) > 0 {
		mID := mf.GetID(a.ManufacturerData())
		mName := mf.GetName(mID)

		log.WithFields(logrus.Fields{
			"ManufacturerID":   fmt.Sprintf("0x%04x", mID),
			"Manufacturer":     mName,
			"ManufacturerData": a.ManufacturerData(),
		}).Debug()

		// do each known Manufacturer
		if mName == "Apple, Inc." {
			ret := apple.ParseMF(a.ManufacturerData())
			log.Debug(ret)
		}
	}

	log.WithFields(logrus.Fields{
		//"test": a.EventType(),
		"RAW":              utils.FormatHex(hex.EncodeToString(a.LEAdvertisingReportRaw())),
		"SR":               utils.FormatHex(hex.EncodeToString(a.ScanResponseRaw())),
		"TxPowerLevel":     a.TxPowerLevel,
		"OverflowService":  a.OverflowService,
		"SolicitedService": a.SolicitedService,
		"Connectable":      a.Connectable,
	}).Trace()

	//log.Error("%s",utils.FormatHex(hex.EncodeToString(   LEAdvertisingReport())))

	//NOTE: when saving to Elastic, make sure you save the RAW packet as well
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		//log.Info("%s ---\n", time.Now().Format(time.RFC3339))
		//fmt.Printf("---\n") // no data seen?
		log.Info("---")
	case context.Canceled:
		log.Warn("Cancelled\n")
		isRunning = false
	default:
		log.Fatalf(err.Error())
	}
}
