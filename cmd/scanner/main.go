package main

// https://towardsdatascience.com/spelunking-bluetooth-le-with-go-c2cff65a7aca
// https://gist.github.com/sausheong/16afa3c4018a22a737f08416768cab90#file-adscanhandler-go

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	// BLE
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/sausheong/ble"
	"github.com/sausheong/ble/examples/lib/dev"
	"github.com/sirupsen/logrus"

	// My stuff
	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
	apple "github.com/mohclips/BLEAS2/internal/manufacturers/apple"
	"github.com/mohclips/BLEAS2/internal/utils"

	//
	"github.com/olivere/elastic/v7"
	//
	"gopkg.in/yaml.v2"
	//
	"github.com/tidwall/sjson"

	log "github.com/mohclips/BLEAS2/internal/logging"
)

var (
	configPath = flag.String("configPath", "./config.yml", "path to config file")
)

// Global config struct
var cfg *Config

var esctx context.Context
var esclient *elastic.Client

var isRunning bool = false

const debug bool = false

// Config struct for webapp config
type Config struct {
	// Host is the local machine IP Address to bind the HTTP Server to
	Host string `yaml:"host"`

	// default BLE device
	Device string `yaml:"device"`
	// duration of scan
	Duration time.Duration `yaml:"duration"`
	// allow recording fof duplicate events
	AllowDuplicates bool `yaml:"allow_duplicates"`
	// allow ID152 events
	AllowID152 bool `yaml:"allow_id152"`
	// allow NHS Covid19 app events
	AllowNHS bool `yaml:"allow_nhs"`
	// record (stdout) no events seen in scan duration
	LogBlanks bool   `yaml:"log_blanks"`
	LogLevel  string `yaml:"log_level"`
	Elastic   struct {
		// elastic stack destination urls
		Servers []string
		// api key id
		APIKeyID string `yaml:"api_key_id"`
		// api key secret
		APIKeySecret string `yaml:"api_id_secret"`
		// index to save event to
		Index string `yaml:"index"`
	} `yaml:"elastic"`
}

func main() {

	// needed to access Bluetooth and set USB power
	if !utils.RunningAsRoot() {
		log.Fatal("This program must be run as root! (sudo)")
	}

	// make sure USB power is 'on' not 'auto' otherwise we have a suspend issue
	utils.CheckBtUsbPower()

	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(*configPath); err != nil {
		log.Fatal("%s", err)
	}

	var err error
	cfg, err = NewConfig(*configPath)
	if err != nil {
		log.Fatal("%s", err)
	}

	//log.Info("%+v", cfg)

	//spew.Dump(cfg)

	loglevel, _ := logrus.ParseLevel(cfg.LogLevel)
	log.SetLevel(loglevel)

	// FIXME: direct access needed
	es := cfg.Elastic.Servers
	index := cfg.Elastic.Index

	// initialise ElasticStack

	// split es flag into array
	if len(es) > 0 && index != "" {

		esctx = context.Background()
		var err error
		esclient, err = getESClient()
		if err != nil {
			log.Error("Error initializing : %s", err)
			panic("Client fail ")
		}

	}
	log.Info("Running...")

	d, err := dev.NewDevice(cfg.Device)
	if err != nil {
		log.Fatal("can't 'new' device : %s", err)
	}
	ble.SetDefaultDevice(d)

	// best struct pprint
	//spew.Dump(d)

	isRunning = true

	// Scan for specified duration, AND until interrupted by user.
	for isRunning {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), cfg.Duration))
		chkErr(ble.Scan(ctx, cfg.AllowDuplicates, advHandler, nil))
	}
}

// #######################################################################################

func advHandler(a ble.Advertisement) {

	// parsed - data ready for elastic
	var parsed string
	// error
	var err error

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
	if a.LocalName() == "ID152" && cfg.AllowID152 == false {
		return
	}

	// only display ID152 pkts with MD in them
	if (len(a.LocalName()) > 0) &&
		(a.LocalName() == "ID152") &&
		(len(a.ManufacturerData()) == 0) {
		return
	}

	// do not display NHS app
	if len(a.Services()) > 0 {
		if a.Services()[0].String() == "fd6f" && cfg.AllowNHS == false {
			return
		}
	}

	// log basics to console
	log.Info("Addr:%s rssi:%d", a.Addr(), a.RSSI())

	if len(a.LocalName()) > 0 {
		log.Debug("LocalName: %s", a.LocalName())
	}

	// Trace level debug
	log.Trace("RAW: %s SR:%s TxPowerLevel:%d OverflowService:%s SolicitedService:%s Connectable:%T",
		utils.FormatHex(hex.EncodeToString(a.LEAdvertisingReportRaw())),
		utils.FormatHex(hex.EncodeToString(a.ScanResponseRaw())),
		a.TxPowerLevel(),
		a.OverflowService(),
		a.SolicitedService(),
		a.Connectable(),
	)

	//
	// Services data provided
	//
	//TODO: parse services
	if len(a.Services()) > 0 {
		log.Debug("Services: %s", a.Services())
	}

	//
	// Manufacturer data provided
	//
	var mID uint16
	var mName string
	if len(a.ManufacturerData()) > 0 {
		mID = mf.GetID(a.ManufacturerData())
		mName = mf.GetName(mID)

		log.Info("Manufacturer: 0x%04x : %s", mID, mName)
		//
		// do each known Manufacturer
		//
		parsedOk := false

		// list known manufacturer parsers here
		if mName == "Apple, Inc." {
			parsed = apple.ParseMF(a.ManufacturerData())
			parsedOk = true
		}

		// if not parsed then Debug
		if !parsedOk {
			log.Debug("ManufacturerID: %s Manufacturer: %s ManufacturerData: %s",
				fmt.Sprintf("0x%04x", mID),
				mName,
				utils.FormatDecComma(hex.EncodeToString(a.ManufacturerData())),
			)

			// this has issues and i dont like it.
			parsed = fmt.Sprintf("{ %q: [%s] }", "unparsed", utils.FormatDecComma(hex.EncodeToString(a.ManufacturerData())))
		}

	}

	// ParsedManufacturerData - The parsed data we are after
	type ParsedManufacturerData struct {
		ID   uint16 `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		// we add more here once its Marshalled as json
		Details string `json:"details,omitempty"`
	}

	// Device - represents a BLE device, with our parsed data tacked on
	type Device struct {
		Timestamp        string                 `json:"@timestamp"`
		Address          string                 `json:"address"`
		Detected         string                 `json:"detected"`
		Since            string                 `json:"since,omitempty"`
		Name             string                 `json:"name,omitempty"`
		RSSI             int                    `json:"rssi"`
		Advertisement    string                 `json:"advertisement,omitempty"`
		ScanResponse     string                 `json:"scanresponse,omitempty"`
		ManufacturerData ParsedManufacturerData `json:"manufacturerdata,omitempty"`
	}

	device := Device{
		Timestamp:     time.Now().Format(time.RFC3339),
		Address:       fmt.Sprintf("%s", a.Addr()),
		Detected:      time.Now().Format(time.RFC3339),
		Since:         "",
		Name:          a.LocalName(),
		RSSI:          a.RSSI(),
		Advertisement: utils.FormatHex(hex.EncodeToString(a.LEAdvertisingReportRaw())),
		ScanResponse:  utils.FormatHex(hex.EncodeToString(a.ScanResponseRaw())),
		ManufacturerData: ParsedManufacturerData{
			ID:      mID,
			Name:    mName,
			Details: "replaced by sjson", // replaced by sjson
		},
	}

	//fmt.Printf("DEVICE: %+v\n", device)
	//spew.Dump(device)

	//
	// append "parsed" to struct containing BLE data
	//
	// convert to json
	var dpkt []byte
	dpkt, err = json.Marshal(device)
	if err != nil {
		log.Error("%s", err)
	}

	var rjson string = ""
	if len(parsed) > 0 {
		// now replace the parsed data
		rjson, _ = sjson.SetRaw(string(dpkt), "manufacturerdata.details", parsed)
	} else {
		//rjson = string(dpkt)
		log.Warn("no manufacturer details present")
		rjson, _ = sjson.SetRaw(string(dpkt), "manufacturerdata.details", "{}")
	}

	//spew.Dump(dpkt)
	//fmt.Printf(">>>%s\n", rjson)

	// TODO: need to work out Duplicates. need 'allowdups=600' flag
	// so sha256 of elastic data, if over 'x' seconds old re-send, else drop

	// send to elastic
	log.Debug("sending to Elastic: %s", rjson)

	//fmt.Printf("DEBUG: %+v\n", parsed)

	decodeJSON := false
	if decodeJSON {
		var dat map[string]interface{}
		bytes := []byte(rjson)
		if err := json.Unmarshal(bytes, &dat); err != nil {
			panic(err)
		}
		spew.Dump(dat)
	}

	sendToES := true
	// index parsed event
	// https://pkg.go.dev/github.com/olivere/elastic/v7@v7.0.22?utm_source=gopls#IndexResponse
	if sendToES {
		var indexResponse *elastic.IndexResponse // only good when err == nill
		indexResponse, err = esclient.Index().
			Index(cfg.Elastic.Index).
			Type("_doc").
			BodyJson(string(rjson)).
			Do(esctx)

		if err != nil {
			// added for debug of unmatched issues
			log.Error("ERROR ==============================================================================================")
			log.Error("parsed: %s", parsed)
			log.Error("rjson: %s", rjson)
			// now die!
			log.Panic("%s", err)
		}

		// log that to was saved to ES
		log.Debug("Index: %s, Id: %s", indexResponse.Index, indexResponse.Id)
	}
}

//###############################################################################################################################################
//###############################################################################################################################################
//###############################################################################################################################################

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		if cfg.LogBlanks == true {
			log.Info("--- no advertisments seen")
		}
	case context.Canceled:
		log.Warn("Cancelled\n")
		isRunning = false
	default:
		log.Fatal(err.Error())
	}
}

// GetESClient - open a connection to ES
func getESClient() (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL(cfg.Elastic.Servers...),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))

	var exists bool
	// does index exist?
	exists, err = client.IndexExists(cfg.Elastic.Index).Do(context.Background())
	if err != nil {
		log.Panic("%s", err)
	}
	if !exists {
		log.Panic(fmt.Sprintf("index '%s' does not exist, %s\n", cfg.Elastic.Index, err))
	}

	log.Info("ES initialized...")

	return client, err

}

// based on: https://dev.to/koddr/let-s-write-config-for-your-golang-web-app-on-right-way-yaml-5ggp

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
