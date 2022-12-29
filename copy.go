package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
	progressbar "github.com/schollz/progressbar/v3"
)

// copy documents from a source firestore db to target firestore "cloud tests" db
func copy(ctx context.Context, p *flags.Parser) error {
	// collect args provided by user
	var opts optsCopy
	// populate our opts
	ok := opts.set(p)
	if !ok {
		fmt.Printf("%s cmd line args cannot be parsed!\n", red("ERROR"))
	}
	//fmt.Printf("DEBUG: Opts struct:: %v\n", opts)
	// create our source firestore client
	conf := fsClientConfig{c: opts.Source}
	sc := createFirestoreClient(ctx, conf)
	// create our target firestore client
	conf = fsClientConfig{c: opts.Target}
	tc := createFirestoreClient(ctx, conf)

	// create our base test document in target db using user-provided params
	testDoc := tc.Collection("Tests")
	ref := testDoc.Doc(opts.Name) // ref is used to store our report data in further on, use test "name" as document id
	testDocRef := ref             // keep track of original test document ref if we need to delete it later
	_, err := ref.Set(ctx, opts)
	if err != nil {
		fmt.Printf("%s setting our test document up\n", red("ERROR"))
		fmt.Println(err)
		return err
	}

	// build source query //
	// iterate through each supported Report collection for a transponder
	for _, reportCollection := range SupportedTransponderReports {
		// build query
		// ex: account/18/vehicle/83/report_data
		sq := fmt.Sprintf("account/" + strconv.Itoa(opts.Account) + "/vehicle/" + strconv.Itoa(opts.Transponder) + "/" + reportCollection)
		//fmt.Printf("DEBUG: assembled query string:: %s\n", sq)
		stests := sc.Collection(sq).Where("reportTimestamp", ">", opts.StartTime).Where("reportTimestamp", "<", opts.EndTime)
		//fmt.Printf("DEBUG: firestore.Query:: %v\n", stests)

		// get all docs that match
		iter1, err := stests.Documents(ctx).GetAll()
		if err != nil {
			fmt.Printf("%s querying source collection: %s\n", red("ERROR"), blue(reportCollection))
			fmt.Println(err)
			return err
		}
		// basic copy operation metrics
		var docsAdded int
		docsTotal := len(iter1)
		bar := progressbar.Default(int64(docsTotal))
		// see if we have any results to copy w/ given parameters
		if docsTotal == 0 {
			// warn user
			fmt.Printf("%s no reports were found for given parameters. Cleaning up parent reference document...\n", yellow("WARN"))
			// delete testDocRef
			_, err = testDocRef.Delete(ctx)
			if err != nil {
				fmt.Printf("%s deleting test reference doc...\n", red("ERROR"))
			}
		}

		dtest := ref.Collection(reportCollection) // use our test doc ref + reportCollection type
		// run through returned documents and copy them to target
		for _, doc := range iter1 { // iterate through all docs received and set in destination firestore
			_, err := dtest.NewDoc().Set(ctx, doc.Data())
			//fmt.Printf("DEBUG: %s : %v\n", green("copied"), doc.Data())
			if err != nil {
				fmt.Printf("%s setting new documents in target collection: %s\n", red("ERROR"), blue(reportCollection))
				fmt.Println(err)
				return err
			}
			// metrics
			docsAdded++
			bar.Add(1)
			// if DEBUG disable bar and use colored count?
			//fmt.Printf("%s/%s  ", green(strconv.Itoa(docsAdded)), blue(strconv.Itoa(docsTotal)))
		}
		if docsTotal != 0 {
			fmt.Printf("\n%s\n", green("success"))
		}
	}
	return nil
}

// Methods //

// convert and set user-provided values into opts struct
func (o *optsCopy) set(p *flags.Parser) (ok bool) {
	o.Transponder, ok = p.Active.FindOptionByLongName("transponderId").Value().(int)
	if !ok {
		return false
	}
	o.Account, ok = p.Active.FindOptionByLongName("accountId").Value().(int)
	if !ok {
		return false
	}
	o.Tag, ok = p.Active.FindOptionByLongName("tag").Value().([]string)
	if !ok {
		return false
	}
	o.Stime, ok = p.Active.FindOptionByLongName("startTime").Value().(int64)
	if !ok {
		return false
	} else { // convert epoch millis to time.Time
		o.StartTime = time.Unix(0, o.Stime*int64(time.Millisecond)).UTC()
	}
	o.Etime, ok = p.Active.FindOptionByLongName("endTime").Value().(int64)
	if !ok {
		return false
	} else { // convert epoch millis to time.Time
		o.EndTime = time.Unix(0, o.Etime*int64(time.Millisecond)).UTC()
	}
	o.Description, ok = p.Active.FindOptionByLongName("description").Value().(string)
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
	o.Target, ok = p.Active.FindOptionByLongName("target").Value().(string)
	if !ok {
		return false
	}
	return true
}
