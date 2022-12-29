package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/jessevdk/go-flags"
	progressbar "github.com/schollz/progressbar/v3"
)

// Replay data from a Test document will look for report_data, video_data, eld_data - or whatever data collections are set as by Firestream
// Replays /Test/docId/report_data onto /account/accId/device/deviceId/report_data
// We will sit and wait while doing this. Report progress (we know total result count / where we are) (we also know total test time and how far along we are)
// Any errors just return false, upstream func with do further handling

func replay(ctx context.Context, p *flags.Parser) error {
	// collect args provided by user
	var opts optsReplay
	// populate our opts
	ok := opts.set(p)
	if !ok {
		fmt.Printf("%s cmd line args cannot be parsed!\n", red("ERROR"))
	}

	// For source data we're using cloud firestore + a service account file (most likely it's test-latinum)...
	conf := fsClientConfig{c: opts.Source}
	sc := createFirestoreClient(ctx, conf)

	// For target data we're using a local firestore emulator, attempt to connect via option.WithGRPCCONN
	conf = fsClientConfig{c: opts.Target}
	if opts.TargetEmulator { // pointing at a firestore emulator, good b/c that's all we support right now
		conf.e = opts.EmulatorProjectId
		conf.l = true
	} else {
		// streaming to a non-local firestore db is currently unsupported
		errMsg := fmt.Sprintf("%s replay sub-command doesn't support anything beyond a local firestore emulator", red("FATAL"))
		return errors.New(errMsg)
	}
	tc := createFirestoreClient(ctx, conf) // firestore source connection

	// get our supported data collections so we can combine them all into a big structure
	// then we sort ALL events by fsCreateTimestamp to get our "playlist" of data
	// all of our data is now in a queue to be written keyed by the time it was 1st written
	// next we will hit "play" and start writing, this will now require us to
	// modify each record's reportTimestamp ideally to an offset equal to fsCreateTimestamp - reportTimestamp
	// of our original report, so if we have a 7 second diff between reportTimestamp -> fsCreateTimestamp
	// we will carry that over to our new reportTimestamp to ensure that the same effect of "late" or delayed
	// data is visible within the replay environment...

	// find our "Tests" document in Firestore: Tests/{testDocId} to locate our test data collections

	// We only support FirestoreTransponderReportV1 type "report_data" structures in Firestore.
	// At some point we'll need to go get other types of reports and play those back as well.
	// We are using Firestore to sort all of our entries back to us by fsCreateTimestamp
	reportCollection := SupportedTransponderReports[0]
	// Tests/{testDocId}/{reportCollection}/{reportDataDocuments}
	sCollection := sc.Collection("Tests/" + opts.Name + "/" + reportCollection)
	//fmt.Printf("DEBUG: source collection query: %v\n", sCollection)
	sIter1, err := sCollection.OrderBy("fsCreateTimestamp", firestore.Asc).Documents(ctx).GetAll() // no query params here, get it all
	if err != nil {
		fmt.Printf("%s querying source collection: %s\n", red("ERROR"), blue(reportCollection))
		fmt.Println(err)
		return err
	}
	// if there were no documents available for this report type, warn and move on
	docsTotal := len(sIter1)
	if docsTotal == 0 {
		// warn
		fmt.Printf("%s no reports found for type %s in Test Firestore document...\n", yellow("WARNING"), blue(reportCollection))
	}

	// destination setup
	tCollectionRef := fmt.Sprintf("account/" + strconv.Itoa(opts.Account) + "/vehicle/" + strconv.Itoa(opts.Transponder) + "/" + reportCollection)
	//fmt.Printf("DEBUG: assembled target ref string:: %s\n", tCollectionRef)
	tCollection := tc.Collection(tCollectionRef)

	// basic copy operation metrics
	var docsAdded int
	bar := progressbar.Default(int64(docsTotal))

	// timer items to pace our writes back into Firestore so they appear real
	var sleepyTime time.Duration
	var lastFsCreationTime time.Time

	for i, doc := range sIter1 { // pull source collection's documents and push them out to target collection
		// unpack report data into struct
		p := FirestoreTransponderReportV1{}
		doc.DataTo(&p)

		// set new sleep time upon second iteration through our range
		if i != 0 {
			sleepyTime = p.FirestoreCreation.Sub(lastFsCreationTime)
		}
		// wait until we are ready
		time.Sleep(sleepyTime)

		// intentionally zero out eventStart, because we aren't synth'ing it
		p.EventStart = time.Time{}

		// set serial number to user-requested
		p.Serial = float64(opts.Transponder)

		// keep track of our re-play timeline
		lastFsCreationTime = p.FirestoreCreation

		// differential between this report's fsCreateTimestamp and reportTimestamp
		// this gives us our "delay" between transponder making a report, it hitting cl api and then firestore
		diff := p.FirestoreCreation.Sub(p.ReportTimestamp)
		now := now()
		p.ReportTimestamp = now.Add(-diff)

		// write it out
		tCollection.NewDoc().Set(ctx, p)
		bar.Add(1) // progress tracking
		docsAdded++
	}
	if docsTotal != 0 {
		fmt.Printf("\n%s\n", green("success"))
	}

	return nil
}

func (o *optsReplay) set(p *flags.Parser) (ok bool) {
	o.Account, ok = p.Active.FindOptionByLongName("accountId").Value().(int)
	if !ok {
		return false
	}
	o.Transponder, ok = p.Active.FindOptionByLongName("transponderId").Value().(int)
	if !ok {
		return false
	}
	o.Name, ok = p.Active.FindOptionByLongName("name").Value().(string)
	if !ok {
		return false
	}
	o.Source, ok = p.Active.FindOptionByLongName("source").Value().(string)
	if !ok {
		return false
	}
	o.EmulatorProjectId, ok = p.Active.FindOptionByLongName("projectId").Value().(string)
	if !ok {
		return false
	}
	o.Target, ok = p.Active.FindOptionByLongName("target").Value().(string)
	if !ok {
		return false
	} else {
		// see if this matches a localhost:port string (ze emulator)
		if strings.Contains(o.Target, "localhost:") || strings.Contains(o.Target, "127.0.0.1:") {
			o.TargetEmulator = true
		} else {
			o.TargetEmulator = false
			// non-emulator replay targets not supported atm
			fmt.Printf("%s replaying into non-local db is unsupported", red("ERROR"))
			return false
		}
	}
	return true
}
