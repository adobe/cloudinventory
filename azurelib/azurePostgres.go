package azurelib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
)

// PGServer - struct to store data for each postgres server
type PGServer struct {
	ID         string
	Properties PGServerProperties
	Location   string
	Name       string
	Type       string
}

// PGServerProperties - struct to display "properties" of each PGServer
type PGServerProperties struct {
	AdministratorLogin       string
	StorageProfile           PGStorageProfile
	FullyQualifiedDomainName string
	Version                  string
}

// PGStorageProfile - captures storage profile of the PG server
type PGStorageProfile struct {
	StorageMB int64
}

// PGDescriber - data that we wish to capture on each PG container
type PGDescriber struct {
	FQDN         string
	Version      string
	StorageSpace int64
}

//ExtractPostgresInventory - function that gets the clientID authenticated and runs a go-routine to get Postgres inventory for each subscription the client_id has access to.
func ExtractPostgresInventory() ([]PGDescriber, error) {

	sess, err := newSession()
	if err != nil {
		fmt.Printf("could not create a session. Here is the error: %s \n", err)
		panic(errors.New("could not create a session"))
	}
	listOfSubscriptions, _ := ListSubscriptions(sess)
	//fmt.Println("The credentials supplied have access to the following subscriptions: ", listOfSubscriptions)

	var wg sync.WaitGroup
	var allPg []PGDescriber

	for _, subID := range listOfSubscriptions {
		wg.Add(1)
		go func(subscriptionID string) {

			if err != nil {
				fmt.Printf("cannot obtain the session -- %v\n", err)
				os.Exit(1)
			}
			postgresqlClient := postgresql.NewServersClient(subID)
			postgresqlClient.Authorizer = sess.Authorizer

			var pgServer PGServer

			pgList, pgErr := postgresqlClient.List(context.Background())
			if pgErr != nil {
				fmt.Println(pgErr)
			}

			fmt.Println(" ========== POSTGRES DATA ========= ", " subscription: ", subID, " =======")
			for _, eachpgServer := range *pgList.Value {
				var bs []byte
				var err error

				bs, err = eachpgServer.MarshalJSON()
				if err != nil {
					fmt.Println("cant make sense of pg server list ", err)
				}
				json.Unmarshal(bs, &pgServer)
				allPg = append(allPg, PGDescriber{
					FQDN:         pgServer.Properties.FullyQualifiedDomainName,
					Version:      pgServer.Properties.Version,
					StorageSpace: pgServer.Properties.StorageProfile.StorageMB,
				})

				fmt.Printf("{%v   %v   %v MB} \n\n", pgServer.Properties.FullyQualifiedDomainName, pgServer.Properties.Version, pgServer.Properties.StorageProfile.StorageMB)

			}
			wg.Done()
		}(subID)
		wg.Wait()

	}
	return allPg, nil
}
