package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"google.golang.org/genproto/googleapis/type/latlng"
)

type optsBase struct {
	Verbose  bool         `short:"v" long:"verbose" description:"enable verbose output"`
	Copy     optsCopy     `command:"copy" description:"copy interesting data into Firestore (preferrably test-latinum)"`
	Replay   optsReplay   `command:"replay" description:"replay interesting data from Firestore (test-latinum) into Firestore emulator with re-written reportTimestamps"`
	List     optsList     `command:"list" description:"list available replays from within test-latinum Firestore db"`
	ListTags optsListTags `command:"tags" description:"list all tags available in test db, start here :)"`
}
type optsCopy struct {
	Transponder int       `short:"x" long:"transponderId" description:"cartwheel's transponder id (aka webId)" required:"true"`
	Account     int       `short:"a" long:"accountId" description:"account id transponder belongs to" required:"true"`
	Stime       int64     `short:"s" long:"startTime" description:"milliseconds unix epoch" required:"true"`
	Etime       int64     `short:"e" long:"endTime" description:"milliseconds unix epoch" required:"true"`
	Description string    `short:"d" long:"description" description:"Short description of test data"`
	Name        string    `short:"n" long:"name" description:"Name this test data chunk" required:"true"`
	Source      string    `short:"b" long:"source" description:"Source Firestore db serivce account file" required:"true"`
	Target      string    `short:"g" long:"target" description:"Target (use test-latinum!!) Firestore db service account file" required:"true"`
	Tag         []string  `short:"t" long:"tag" description:"Add provided tag(s) to test ex: '-t e2e -t smoke_test'" required:"true"`
	StartTime   time.Time // contains Stime type-converted into time.Time
	EndTime     time.Time // contains Etime type-converted into time.Time
}
type optsReplay struct {
	Name              string `short:"n" long:"name" description:"Name of test packet to replay" required:"true"`
	Transponder       int    `short:"x" long:"transponderId" description:"transponder serial number to replay onto" required:"true"`
	Account           int    `short:"a" long:"accountId" description:"account id to replay data onto" required:"true"`
	Target            string `short:"g" long:"target" description:"Target env-latinum Firestore db service account file" required:"true"`
	EmulatorProjectId string `short:"p" long:"projectId" description:"projectId used when starting your local firebase emulator" required:"true"`
	Source            string `short:"b" long:"source" description:"Source Test Firestore db 'host:port' string" required:"true"`
	TargetEmulator    bool   // true if we detect a localhost:port string as target
}
type optsList struct {
	Source  string `short:"b" long:"source" description:"Test data Firestore service account file (test-latinum project most likely...)" required:"true"`
	Tag     string `short:"t" long:"tag" description:"List results matching provided tag" required:"true"`
	Results int    `short:"r" long:"results" description:"Number of results to show, default of 10" default:"10"`
}
type optsListTags struct {
	Source  string `short:"b" long:"source" description:"Test data Firestore service account file (test-latinum project most likely...)" required:"true"`
	Results int    `short:"r" long:"results" description:"Number of results to show, default of 10" default:"10"`
}

// Transponder generated reports (speeding, status, hard_accel, ...)
type FirestoreTransponderReportV1 struct {
	ConfigId           float64        `firestore:"configId,omitempty"`
	Duration           float64        `firestore:"duration,omitempty"`
	EventStart         time.Time      `firestore:"eventStart,omitempty"`
	InProgress         bool           `firestore:"inProgress,omitempty"`
	LocationAccuracy   float64        `firestore:"locationAccuracy,omitempty"`
	Heading            float64        `firestore:"heading,omitempty"`
	Address            string         `firestore:"address,omitempty"`
	DotOrientation     string         `firestore:"dotOrientation,omitempty"`
	GeoTags            []string       `firestore:"geoTags,omitempty"`
	LatLng             *latlng.LatLng `firestore:"latLng,omitempty"`
	BatteryVoltage     float64        `firestore:"batteryVoltage,omitempty"`
	CellSignalStrength float64        `firestore:"cellSignalStrength,omitempty"`
	IsLowBattery       bool           `firestore:"isLowBatteryVoltage,omitempty"`
	Odometer           float64        `firestore:"odometer,omitempty"`
	Speed              float64        `firestore:"speed,omitempty"`
	SpeedLimit         float64        `firestore:"speedLimit,omitempty"`
	ReportTimestamp    time.Time      `firestore:"reportTimestamp,omitempty"`
	Serial             float64        `firestore:"serial,omitempty"`
	Type               string         `firestore:"type"`                              // this is the "dataType" field of a streaming packet
	FirestoreCreation  time.Time      `firestore:"fsCreateTimestamp,serverTimestamp"` // if zero, Firestore sets this on their end
}

var (
	args                        = new(optsBase)
	red                         = color.New(color.FgRed).SprintfFunc()
	green                       = color.New(color.FgGreen).SprintfFunc()
	yellow                      = color.New(color.FgYellow).SprintfFunc()
	blue                        = color.New(color.FgBlue).SprintfFunc()
	SupportedTransponderReports = [...]string{"report_data"} // future additions of eld_data, video_upload ...
)

func init() {
}

func main() {
	ctx := context.Background()
	// parse args
	p := flags.NewParser(args, flags.Default)
	_, err := p.Parse()
	if err != nil {
		log.Printf("%s Unable to parse args.\n", red("FATAL"))
		os.Exit(1)
	}
	switch p.Active.Name {
	case "copy":
		err := copy(ctx, p)
		if err != nil {
			//
		}
	case "replay":
		err := replay(ctx, p)
		if err != nil {
			//
		}
	case "list":
		err := list(ctx, p)
		if err != nil {
			//
		}
	case "tags":
		tags(ctx, p)
	}
}
