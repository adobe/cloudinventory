package azurelib

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-03-01/resources"
	log "github.com/sirupsen/logrus"
)

// NWInt - captures the structure of a Network Interface
type NWInt struct {
	ID         string
	Location   string
	Name       string
	Properties NWIntProperties
}

//NWIntProperties - captures the "Properties" in a Network Interface
type NWIntProperties struct {
	IPConfigurations []NWIntConf
}

// NWIntConf - struct that stores the Network Interface response
type NWIntConf struct {
	Etag       string
	ID         string
	Name       string
	Properties NWIntProp
}

// NWIntProp - struct that stores the network interface "properties" field
type NWIntProp struct {
	PrivateIPAddress          string
	PrivateIPAllocationMethod string
	PrivateIPAddressVersion   string
	PublicIPAddress           NWIntPubIP
}

// NWIntPubIP - struct that stores the PublicIP's ID
type NWIntPubIP struct {
	ID string
}

// VMDescriber - describes each VM
type VMDescriber struct {
	VMName    string
	PrivateIP string
	PublicIP  string
	Location  string
}

//VirtualMachineData - Struct that describes the VM
type VirtualMachineData struct {
	Location   string
	ID         string
	Properties VMProperties
}

//VMProperties - struct that describes the "properties"  of VM
type VMProperties struct {
	HardwareProfile VMHardwareProfile
	NetworkProfile  VMNetworkProfile
	OSprofile       VMOSProfile
}

//OSProfile from VMObject has the VM Name
type VMOSProfile struct {
	ComputerName string
}

// VMHardwareProfile - describes Hardware profile of the VM
type VMHardwareProfile struct {
	VMSize string
}

//VMNetworkProfile - struct that describes the Network profile of the VM
type VMNetworkProfile struct {
	NetworkInterfaces []NWInt
}

// takes the resourceID as an input and returns a generic resource object
func describeResource(sess *AzureSession, subscriptionID string, rID string) (resources.GenericResource, error) {
	resourcesClient := resources.NewClient(subscriptionID)
	resourcesClient.Authorizer = sess.Authorizer

	resData, resErr := resourcesClient.GetByID(context.Background(), rID)

	return resData, resErr
}

// takes a NIC resource object and extracts the IP address from it
func extractIPfromNicRes(genRes resources.GenericResource) (string, error) {
	var bs []byte
	var err error
	type GenProps struct {
		IPAddress string
	}
	type GenResType struct {
		ID         string
		Name       string
		Properties GenProps
	}

	var GenRes1 GenResType

	bs, err = genRes.MarshalJSON()
	json.Unmarshal(bs, &GenRes1)
	return GenRes1.Properties.IPAddress, err

}

//ExtractVMInventory - function that gets the clientID authenticated and runs a go-routine to extract VM inventory for each subscription the client_id has access to
func ExtractVMInventory() ([]VMDescriber, error) {

	sess, err := newSession()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("Error while extracting Azure VM inventory. Could not create a session")
		panic(errors.New("could not create a session"))
		return nil, err
	}
	listOfSubscriptions, _ := ListSubscriptions(sess)

	var wg sync.WaitGroup
	var allVMs []VMDescriber

	for _, subID := range listOfSubscriptions {
		wg.Add(1)
		go func(subscriptionID string) {

			vmClient := compute.NewVirtualMachinesClient(subID)
			vmClient.Authorizer = sess.Authorizer

			for vmList, vmErr := vmClient.ListAllComplete(context.Background()); vmList.NotDone(); vmErr = vmList.Next() {
				if vmErr != nil {
					log.WithFields(log.Fields{"Error": vmErr}).Warning("Error while obtaining a list of AzureVMs")
					os.Exit(1)
				}
				var bs []byte
				var EachVM VMDescriber
				EachVM.PublicIP = "Public-IP-Not-Created"
				var VMData VirtualMachineData
				bs, _ = vmList.Value().MarshalJSON()
				json.Unmarshal(bs, &VMData)
				EachVM.VMName = VMData.Properties.OSprofile.ComputerName
				EachVM.Location = VMData.Location
				for _, eachNic := range VMData.Properties.NetworkProfile.NetworkInterfaces {
					nicRes, err := describeResource(sess, subID, eachNic.ID)
					if err != nil {
						log.WithFields(log.Fields{"Error": err}).Warning("Cannot describe resource")
						os.Exit(1)
					}
					var eachNWInt NWInt
					bs, _ := nicRes.MarshalJSON()
					json.Unmarshal(bs, &eachNWInt)

					for _, eachIP := range eachNWInt.Properties.IPConfigurations {
						EachVM.PrivateIP = eachIP.Properties.PrivateIPAddress
						if len(eachIP.Properties.PublicIPAddress.ID) > 0 {
							pubipRes, err := describeResource(sess, subID, eachIP.Properties.PublicIPAddress.ID)
							if err != nil {
								log.WithFields(log.Fields{"Error": err}).Warning("Cannot retrieve Public IP")
							}

							publicIP, err := extractIPfromNicRes(pubipRes)
							if err != nil {
								log.WithFields(log.Fields{"Error": err}).Warning("error while extracting Public IP from its generic resource")
							}
							if len(publicIP) > 0 {
								EachVM.PublicIP = publicIP
							}

						}

					}

				}
				log.WithFields(log.Fields{"Data": EachVM, "subscription": subID}).Info("Printing data for each VM")
				allVMs = append(allVMs, EachVM)
			}
			wg.Done()
		}(subID)
		wg.Wait()

	}
	return allVMs, nil
}
