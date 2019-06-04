package azurelib

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{"Error": err}).Warning("Error while extracting Azure Postgres inventory. Could not create a session")
		panic(errors.New("could not create a session"))
		return nil, err
	}
	listOfSubscriptions, _ := ListSubscriptions(sess)

	var wg sync.WaitGroup
	var allPg []PGDescriber

	for _, subID := range listOfSubscriptions {
		wg.Add(1)
		go func(subscriptionID string) {

			postgresqlClient := postgresql.NewServersClient(subID)
			postgresqlClient.Authorizer = sess.Authorizer

			var pgServer PGServer

			pgList, pgErr := postgresqlClient.List(context.Background())
			if pgErr != nil {
				log.WithFields(log.Fields{"PGError": pgErr}).Error("cannot list Postgres Servers")
				os.Exit(1)
			}
			log.WithFields(log.Fields{"subscriptionID": subID}).Info("POSTGRES DATA")
			for _, eachpgServer := range *pgList.Value {
				var bs []byte
				var err error

				bs, err = eachpgServer.MarshalJSON()
				if err != nil {
					log.WithFields(log.Fields{"PGError": err}).Error("cant make sense of pg server list")
					os.Exit(1)

				}
				json.Unmarshal(bs, &pgServer)
				allPg = append(allPg, PGDescriber{
					FQDN:         pgServer.Properties.FullyQualifiedDomainName,
					Version:      pgServer.Properties.Version,
					StorageSpace: pgServer.Properties.StorageProfile.StorageMB,
				})

				log.WithFields(log.Fields{"FQDN": pgServer.Properties.FullyQualifiedDomainName, "Version": pgServer.Properties.Version, "size": pgServer.Properties.StorageProfile.StorageMB, "subscription": subID}).Info("Data from each postgres instance")
			}
			wg.Done()
		}(subID)
		wg.Wait()

	}
	return allPg, nil
}
