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
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/sausheong/ble"
	"github.com/sausheong/ble/examples/lib/dev"
	"github.com/sirupsen/logrus"

	"github.com/mohclips/BLEAS2/internal/aggregator"
	log "github.com/mohclips/BLEAS2/internal/logging"
	mf "github.com/mohclips/BLEAS2/internal/manufacturers"
	"github.com/mohclips/BLEAS2/internal/manufacturers/apple"
	"github.com/mohclips/BLEAS2/internal/manufacturers/microsoft"
	"github.com/mohclips/BLEAS2/internal/services"
	"github.com/mohclips/BLEAS2/internal/sink"
	"github.com/mohclips/BLEAS2/internal/utils"
)

var (
	configPath  = flag.String("configPath", "./config.yml", "path to config file")
	maxDuration = flag.Duration("max-duration", 0, "exit after this duration (0 = run forever); useful for time-bounded test runs")
)

// Global config struct
var cfg *Config

var esctx context.Context
var esclient *elastic.Client

var (
	agg       *aggregator.Aggregator
	jsonlSink *sink.JSONL
)

func main() {
	if !utils.RunningAsRoot() {
		log.Fatal("This program must be run as root! (sudo)")
	}

	// USB power must be 'on' (not 'auto') or the controller suspends.
	utils.CheckBtUsbPower()

	flag.Parse()

	var err error
	cfg, err = LoadConfig(*configPath)
	if err != nil {
		log.Fatal("config: %s", err)
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(level)
	}

	if len(cfg.Elastic.Servers) > 0 && cfg.Elastic.Index != "" {
		esctx = context.Background()
		esclient, err = getESClient()
		if err != nil {
			log.Fatal("elastic init: %s", err)
		}
	}

	if cfg.Output.JSONL.Enabled {
		jsonlSink, err = sink.NewJSONL(sink.JSONLConfig{
			Dir:           cfg.Output.JSONL.Dir,
			Rotation:      cfg.Output.JSONL.Rotation,
			MaxSize:       int64(cfg.Output.JSONL.MaxSizeMB) * 1024 * 1024,
			CompressAfter: cfg.Output.JSONL.CompressAfter,
		})
		if err != nil {
			log.Fatal("jsonl init: %s", err)
		}
		defer jsonlSink.Close()
		log.Info("JSONL sink: %s (rotation=%s)", cfg.Output.JSONL.Dir, cfg.Output.JSONL.Rotation)
	}

	// One aggregator owns the dedup + per-window stats.
	// window <= 0 → no batching (emit on every observation, count=1).
	var window time.Duration
	if cfg.Dedup.Enabled {
		window = cfg.Dedup.Window
	}
	agg = aggregator.New(window, emitObservation)
	defer agg.Close()
	if window > 0 {
		log.Info("aggregation window: %s", window)
	} else {
		log.Info("aggregation: disabled (per-packet emit)")
	}

	if cfg.Exclude.Enabled {
		log.Info("exclude: %d addresses, %d prefixes, %d local names; %d airdrop, %d icloud, %d homekit, %d findmy fingerprints",
			len(cfg.Exclude.Addresses),
			len(cfg.Exclude.AddressPrefixes),
			len(cfg.Exclude.LocalNames),
			len(cfg.Exclude.AirdropHashes),
			len(cfg.Exclude.ICloudIDs),
			len(cfg.Exclude.HomeKitDeviceIDs),
			len(cfg.Exclude.FindMyPublicKeys))
	}

	log.Info("Running...")

	d, err := dev.NewDevice(cfg.Device)
	if err != nil {
		log.Fatal("can't 'new' device : %s", err)
	}
	ble.SetDefaultDevice(d)

	// One top-level context owns shutdown. Ctrl-C cancels it once and the
	// outer loop drains; individual scans get a child context with a timeout.
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *maxDuration > 0 {
		var cancel context.CancelFunc
		rootCtx, cancel = context.WithTimeout(rootCtx, *maxDuration)
		defer cancel()
		log.Info("will exit after %s", *maxDuration)
	}

	for rootCtx.Err() == nil {
		ctx, cancel := context.WithTimeout(rootCtx, cfg.Duration)
		err := ble.Scan(ctx, cfg.AllowDuplicates, advHandler, nil)
		cancel()
		chkErr(rootCtx, err)
	}
	log.Info("shutdown")
}

// #######################################################################################

func advHandler(a ble.Advertisement) {

	// parsed - data ready for elastic
	var (
		// service data
		sParsed string
		// manufacture data
		mParsed string
	)

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

	// User-configurable exclusion list — drop our own household devices so
	// the capture reflects external/passing broadcasters only.
	if cfg.Excludes(a.Addr().String(), a.LocalName()) {
		return
	}

	// ID152 is a very chatty device; drop unless allowed, and even then only
	// keep packets that carry manufacturer data.
	if a.LocalName() == "ID152" {
		if !cfg.AllowID152 || len(a.ManufacturerData()) == 0 {
			return
		}
	}

	// do not display NHS app
	if !cfg.AllowNHS {
		for _, s := range a.Services() {
			if s.String() == "fd6f" {
				return
			}
		}
	}

	// Cheap dedup short-circuit. Build the dedup key from raw bytes (cheap)
	// BEFORE the expensive parse/marshal below. If an identical packet (same
	// addr + manufacturer + advertising bytes) has already been seen this
	// window, just record the new RSSI sample and bail — the bucket already
	// passed every exclusion check on its first sighting, and re-parsing it
	// only to discard the result is what pegs the CPU at busy locations.
	//
	// The HCI LE Advertising Report has a trailing RSSI byte that DOES change
	// per packet — strip it so identical payloads collapse to one bucket.
	advReport := a.LEAdvertisingReportRaw()
	if len(advReport) > 0 {
		advReport = advReport[:len(advReport)-1]
	}
	key := a.Addr().String() + "|" +
		hex.EncodeToString(a.ManufacturerData()) + "|" +
		hex.EncodeToString(advReport)
	rssi := a.RSSI()

	if agg.ObserveExisting(key, rssi) {
		return
	}

	// Per-packet trace — the JSONL sink is the source of truth, this is only
	// for debugging. Demoted to Debug to keep info-level logs readable.
	log.Debug("Addr:%s rssi:%d", a.Addr(), a.RSSI())

	if len(a.LocalName()) > 0 {
		log.Debug("LocalName: %s", a.LocalName())
	}

	// Trace level debug
	log.Trace("RAW: %s SR:%s TxPowerLevel:%d OverflowService:%s SolicitedService:%s Connectable:%t",
		utils.FormatHexComma(hex.EncodeToString(a.LEAdvertisingReportRaw())),
		utils.FormatHexComma(hex.EncodeToString(a.ScanResponseRaw())),
		a.TxPowerLevel(),
		a.OverflowService(),
		a.SolicitedService(),
		a.Connectable(),
	)

	addressType := utils.BitmaskToNames(int(a.LEAdvertisingReportRaw()[3]), MACaddressTypes)

	////////////////////////////////////////////////////////////////////////////////////////////////
	// Services data — walk every ServiceData entry, dispatch each to the
	// services package, merge into a flat {"key": {...}} dict that mirrors
	// the manufacturerdata.details.apple shape.
	//
	// Also collect the advertised service UUIDs (which can appear separately
	// from service data — e.g. a device declares it supports Battery without
	// putting any data in the advert).
	//
	var sID string
	var sName string
	var sUUIDs []string

	for _, u := range a.Services() {
		sUUIDs = append(sUUIDs, u.String())
	}

	if sds := a.ServiceData(); len(sds) > 0 {
		// First UUID still surfaced as sID/sName for backward-compat with
		// the older reporter queries; the full dict lives under details.
		sID = sds[0].UUID.String()
		if n := services.Name(sID); n != "" {
			sName = n
		}

		combined := map[string]json.RawMessage{}
		for _, sd := range sds {
			uuid := sd.UUID.String()
			name, body := services.Parse(uuid, sd.Data)
			if body == nil {
				continue
			}
			combined[name] = body
		}
		// Special-case Exposure Notification's old name expected by the
		// pre-existing reporter views.
		if _, ok := combined["exposure_notification"]; ok {
			sName = "Google Exposure Notification"
		}
		if buf, err := json.Marshal(combined); err == nil {
			sParsed = string(buf)
		}
	}

	////////////////////////////////////////////////////////////////////////////////////////////////
	// Manufacturer data provided
	//

	// manufacturers ID
	var mID uint16
	// manufacturers Name
	var mName string
	if len(a.ManufacturerData()) > 0 {
		mID = mf.GetID(a.ManufacturerData())
		mName = mf.GetName(mID)

		log.Debug("Manufacturer: 0x%04x : \"\033[1;36m%s\033[0m\"", mID, mName)
		//
		// do each known Manufacturer
		//

		// did we parse the data?
		//parsedOk := false

		// list known manufacturer parsers here
		//if mName == "Apple, Inc." { // 4c
		// if mID == 0x004c {
		// 	mParsed = apple.ParseMF(a.ManufacturerData())
		// 	parsedOk = true
		// 	//} else if mName == "Microsoft" { // 06
		// } else if mID == 0x0006 {
		// 	mParsed = microsoft.ParseMF(a.ManufacturerData())
		// 	parsedOk = true
		// }

		switch mID {
		case 0x0006:
			mParsed = microsoft.ParseMF(a.ManufacturerData())
		case 0x004c:
			mParsed = apple.ParseMF(a.ManufacturerData())
		default:
			log.Debug("ManufacturerID: 0x%04x Manufacturer: %s ManufacturerData: %s",
				mID, mName,
				utils.FormatDecComma(hex.EncodeToString(a.ManufacturerData())),
			)
			unparsed, _ := json.Marshal(map[string][]byte{"unparsed": a.ManufacturerData()})
			mParsed = string(unparsed)
		}
	}
	////////////////////////////////////////////////////////////////////////////////////////////////

	// Identity-fingerprint exclusion: pulled from Apple subtype payloads
	// AFTER parsing, so devices that rotate their BLE MAC can still be
	// filtered consistently.
	if mID == 0x004c && mParsed != "" {
		if cfg.ExcludesByFingerprint(extractFingerprints(mParsed)) {
			return
		}
	}

	// The dedup key and rssi were computed up front (see the short-circuit
	// above). Reaching here means this is the first sighting of this exact
	// payload in the current window, so build the full record.

	// Capture identity values up front — by the time the flush runs the BLE
	// library may have invalidated the Advertisement object.
	addr := a.Addr().String()
	addrType := strings.Join(addressType, ",")
	name := a.LocalName()
	adv := utils.FormatHex(hex.EncodeToString(a.LEAdvertisingReportRaw()))
	sr := utils.FormatHex(hex.EncodeToString(a.ScanResponseRaw()))

	agg.Observe(key, rssi, func() interface{} {
		return &Device{
			Common: CommonData{
				Address:       addr,
				AddressType:   addrType,
				Name:          name,
				Advertisement: adv,
				ScanResponse:  sr,
			},
			ManufacturerData: ParsedManufacturerData{ID: mID, Name: mName, Details: rawOrEmpty(mParsed)},
			ServiceData:      ParsedServiceData{ID: sID, Name: sName, UUIDs: sUUIDs, Details: rawOrEmpty(sParsed)},
		}
	})
}

// emitObservation is the aggregator's flush callback. It receives one bucket
// per (key, window) and writes a single enriched row to the JSONL sink and,
// optionally, Elasticsearch.
func emitObservation(b *aggregator.Bucket) {
	d, ok := b.Record.(*Device)
	if !ok || d == nil {
		log.Error("emit: unexpected record type %T", b.Record)
		return
	}
	first := b.FirstSeen.Format(time.RFC3339)
	d.Timestamp = first
	d.Common.FirstSeen = first
	d.Common.LastSeen = b.LastSeen.Format(time.RFC3339)
	d.Observation = summariseRSSI(b.RSSI)

	dpkt, err := json.Marshal(d)
	if err != nil {
		log.Error("marshal device: %s", err)
		return
	}
	log.Debug("packet: %s", dpkt)

	if jsonlSink != nil {
		if err := jsonlSink.Write(dpkt); err != nil {
			log.Error("jsonl write: %s", err)
		}
	}

	if esclient == nil {
		return
	}
	indexResponse, err := esclient.Index().
		Index(cfg.Elastic.Index).
		Type("_doc").
		BodyJson(string(dpkt)).
		Do(esctx)
	if err != nil {
		log.Error("ES index failed: %s", err)
		log.Error("dpkt: %s", dpkt)
		return
	}
	log.Debug("Index: %s, Id: %s", indexResponse.Index, indexResponse.Id)
}

// extractFingerprints pulls the stable identity fields out of an Apple
// manufacturer JSON blob ({"apple":{"airdrop":{...}, "homekit":{...}, ...}}).
// Used by the exclusion layer to filter MAC-rotating devices by identity.
func extractFingerprints(appleJSON string) Fingerprints {
	var outer map[string]json.RawMessage
	if err := json.Unmarshal([]byte(appleJSON), &outer); err != nil {
		return Fingerprints{}
	}
	apple, ok := outer["apple"]
	if !ok {
		return Fingerprints{}
	}
	var subtypes map[string]json.RawMessage
	if err := json.Unmarshal(apple, &subtypes); err != nil {
		return Fingerprints{}
	}

	var fp Fingerprints

	if ad, ok := subtypes["airdrop"]; ok {
		var v struct {
			AppleIDHash string `json:"apple_id_hash"`
			PhoneHash   string `json:"phone_hash"`
			EmailHash   string `json:"email_hash"`
			Email2Hash  string `json:"email2_hash"`
		}
		if json.Unmarshal(ad, &v) == nil {
			fp.AirdropTuple = v.AppleIDHash + ":" + v.PhoneHash + ":" + v.EmailHash + ":" + v.Email2Hash
		}
	}
	if tt, ok := subtypes["tethering_target"]; ok {
		var v struct {
			ICloudID string `json:"icloud_id"`
		}
		if json.Unmarshal(tt, &v) == nil {
			fp.ICloudID = v.ICloudID
		}
	}
	if hk, ok := subtypes["homekit"]; ok {
		var v struct {
			DeviceID string `json:"device_id"`
		}
		if json.Unmarshal(hk, &v) == nil {
			fp.HomeKitDeviceID = v.DeviceID
		}
	}
	if fm, ok := subtypes["findmy"]; ok {
		var v struct {
			Variant   string `json:"variant"`
			PublicKey string `json:"public_key"`
		}
		if json.Unmarshal(fm, &v) == nil && v.Variant == "separated" {
			fp.FindMyPublicKey = v.PublicKey
		}
	}
	return fp
}

func summariseRSSI(samples []int) Observation {
	if len(samples) == 0 {
		return Observation{}
	}
	min, max, sum := samples[0], samples[0], 0
	for _, v := range samples {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	return Observation{
		Count:       len(samples),
		RSSIMin:     min,
		RSSIMax:     max,
		RSSIMean:    float64(sum) / float64(len(samples)),
		RSSISamples: samples,
	}
}

// rawOrEmpty turns a parser's string output into a valid json.RawMessage,
// substituting "{}" when the parser produced nothing.
func rawOrEmpty(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(s)
}

// Known-noisy scan errors from sausheong/ble that happen on every scan restart
// and are not actionable. Demoted to Debug so info-level logs stay readable.
var transientScanErrs = []string{
	"Command Disallowed",
	"no associated Advertising Data",
}

func chkErr(rootCtx context.Context, err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		if cfg.LogBlanks {
			log.Info("--- no advertisments seen")
		}
	case context.Canceled:
		// Could be either the per-scan timeout's cancel or Ctrl-C. If it's
		// Ctrl-C the rootCtx is already done and the outer loop exits.
	default:
		msg := err.Error()
		noisy := false
		for _, t := range transientScanErrs {
			if strings.Contains(msg, t) {
				noisy = true
				break
			}
		}
		if noisy {
			log.Debug("scan: %s", err)
		} else {
			log.Warn("scan: %s", err)
		}
		// Throttle the retry loop so the HCI controller has time to settle.
		select {
		case <-rootCtx.Done():
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// GetESClient - open a connection to ES
func getESClient() (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL(cfg.Elastic.Servers...),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))
	if err != nil {
		return nil, fmt.Errorf("elastic.NewClient: %w", err)
	}

	exists, err := client.IndexExists(cfg.Elastic.Index).Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("IndexExists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("index %q does not exist", cfg.Elastic.Index)
	}

	log.Info("ES initialized...")

	return client, nil
}
