package main

import (
	"context"
	"fmt"

	"github.com/jessevdk/go-flags"
)

func list(ctx context.Context, p *flags.Parser) error {
	// collects args provided by user
	var opts optsList
	// populate into struct
	ok := opts.set(p)
	if !ok {
		fmt.Printf("%s cmd line args cannot be parsed!\n", red("ERROR"))
	}
	// create firestore client using source file's project_id and credentials
	conf := fsClientConfig{c: opts.Source}
	c := createFirestoreClient(ctx, conf)

	// build query //
	tests := c.Collection("Tests").Where("Tag", "array-contains", opts.Tag).Limit(opts.Results)

	// run query //
	testList, err := tests.Documents(ctx).GetAll()
	if err != nil {
		fmt.Printf("%s querying collection\n", red("ERROR"))
	}
	fmt.Printf("Search results for tag: %s\n", yellow(opts.Tag))
	// pretty print results //
	for _, doc := range testList {
		data := doc.Data()
		name, ok := data["Name"].(string)
		if !ok {
			fmt.Printf("%s no name key found %v\n", red("ERROR"), data)
			continue
		}
		fmt.Printf("%s : %v\n", blue("name"), green(name))
	}
	return nil
}

// Methods //

// convert and set user-provided values into opts struct
func (o *optsList) set(p *flags.Parser) (ok bool) {
	o.Source, ok = p.Active.FindOptionByLongName("source").Value().(string)
	if !ok {
		return false
	}
	o.Results, ok = p.Active.FindOptionByLongName("results").Value().(int)
	if !ok {
		return false
	}
	o.Tag, ok = p.Active.FindOptionByLongName("tag").Value().(string)
	if !ok {
		return false
	}
	return true
}
