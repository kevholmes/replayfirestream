package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jessevdk/go-flags"
)

func tags(ctx context.Context, p *flags.Parser) error {
	// collect args provided by user
	var opts optsListTags
	// populate our opts
	ok := opts.set(p)
	if !ok {
		fmt.Printf("%s cmd line args cannot be parsed!\n", red("ERROR"))
	}
	// create our client based on usr provided opts
	conf := fsClientConfig{c: opts.Source}
	c := createFirestoreClient(ctx, conf)

	// map to store result tags and frequency in
	tagMap := make(map[string]int)

	// query //
	q := c.Collection("Tests").Select("Tag").Limit(opts.Results)
	iter1, err := q.Documents(ctx).GetAll()
	if err != nil {
		fmt.Printf("%s querying collection", red("ERROR"))
	}
	for _, doc := range iter1 {
		data := doc.Data()
		if data == nil {
			fmt.Printf("No results returned from Firestore")
			return nil
		}
		// unpack result and potential array of interfaces (strings here)
		tags, ok := data["Tag"].([]interface{})
		if ok {
			for _, tag := range tags {
				tv := tag.(string)
				tagMap[tv]++
			}
		} else {
			//fmt.Printf("document is missing tag field %v\n", data)
		}
	}
	// print out the tags we found and how many times they're used in test db
	for tag, count := range tagMap {
		fmt.Printf("%s : %s\n", blue(tag), green(strconv.Itoa(count)))
	}
	return nil
}

// Methods //

// convert and set user-provided values into opts struct
func (o *optsListTags) set(p *flags.Parser) (ok bool) {
	o.Source, ok = p.Active.FindOptionByLongName("source").Value().(string)
	if !ok {
		return false
	}
	o.Results, ok = p.Active.FindOptionByLongName("results").Value().(int)
	if !ok {
		return false
	}
	return true
}
