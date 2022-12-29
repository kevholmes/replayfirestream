package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/Jeffail/gabs/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// determine a project_id from a given GCP Service Account JSON key file
func projectIdInServiceAcctFile(p string) (id string, err error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Printf("FATAL: Unable to read service account file: %s", p)
		return id, err
	}
	j, err := gabs.ParseJSON(data)
	if err != nil {
		log.Printf("FATAL: Unable to parse service account file JSON: %s", err)
		return id, err
	}
	id, ok := j.Path("project_id").Data().(string)
	if !ok {
		log.Printf("FATAL: Unable to find project_id key in service account JSON file: %s", p)
		return id, err
	} else {
		return id, nil
	}

}

type fsClientConfig struct {
	c string // firestore connection, could be host:port or service account
	e string // emulator project id, not required for service account file
	l bool   // is this an emulator conn?
}

// create a new Firestore client using a projectId
func createFirestoreClient(ctx context.Context, conf fsClientConfig) *firestore.Client {
	if !conf.l && conf.e == "" { // traditional service account file
		// projId comes from within our credentials file...
		projId, err := projectIdInServiceAcctFile(conf.c)
		if err != nil { // printed out warnings to user already, just bail out
			os.Exit(1)
		}
		o := option.WithCredentialsFile(conf.c)
		client, err := firestore.NewClient(ctx, projId, o)
		if err != nil {
			log.Printf("%s %s", red("ERROR"), err)
		}
		return client
	} else if conf.l && conf.e != "" && conf.c != "" { // firebase projectId provided, ze emulator
		grpcConn, err := grpc.Dial(conf.c, grpc.WithInsecure(), grpc.WithPerRPCCredentials(emulatorCreds{}))
		if err != nil {
			fmt.Printf("ERROR: dialing emulator firestore address")
			os.Exit(1)
		}
		tc, err := firestore.NewClient(ctx, conf.e, option.WithGRPCConn(grpcConn))
		return tc
	} else {
		fmt.Printf("%s invalid configuration passed to createFirestoreClient()\n", red("FATAL"))
		os.Exit(1)
	}
	return &firestore.Client{} // meh, this should never happen
}

// emulatorCreds is an instance of grpc.PerRPCCredentials that will configure a
// client to act as an admin for the Firestore emulator. It always hardcodes
// the "authorization" metadata field to contain "Bearer owner", which the
// Firestore emulator accepts as valid admin credentials.
type emulatorCreds struct{}

func (ec emulatorCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer owner"}, nil
}
func (ec emulatorCreds) RequireTransportSecurity() bool {
	return false
}
