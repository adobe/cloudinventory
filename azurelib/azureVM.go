package azurelib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-03-01/resources"
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
func ExtractVMInventory() {

	sess, err := newSession()
	if err != nil {
		fmt.Printf("could not create a session. Here is the error: %s \n", err)
		panic(errors.New("could not create a session"))
	}
	listOfSubscriptions, _ := ListSubscriptions(sess)
	//fmt.Println("The credentials supplied have access to the following subscriptions: ", listOfSubscriptions)

	var wg sync.WaitGroup

	for _, subID := range listOfSubscriptions {
		wg.Add(1)
		go func(subscriptionID string) {

			if err != nil {
				fmt.Printf("cannot obtain the session -- %v\n", err)
				os.Exit(1)
			}

			vmClient := compute.NewVirtualMachinesClient(subID)
			vmClient.Authorizer = sess.Authorizer

			fmt.Println("============ VM Data ============== ", " subscription: ", subID, " =======")
			for vmList, vmErr := vmClient.ListAllComplete(context.Background()); vmList.NotDone(); vmErr = vmList.Next() {
				if vmErr != nil {
					fmt.Println("cannot retrieve VM list", vmErr)
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
						fmt.Println("cannot describe resource")
					}
					var eachNWInt NWInt
					bs, _ := nicRes.MarshalJSON()
					json.Unmarshal(bs, &eachNWInt)

					for _, eachIP := range eachNWInt.Properties.IPConfigurations {
						EachVM.PrivateIP = eachIP.Properties.PrivateIPAddress
						if len(eachIP.Properties.PublicIPAddress.ID) > 0 {
							pubipRes, err := describeResource(sess, subID, eachIP.Properties.PublicIPAddress.ID)
							if err != nil {
								fmt.Println("cannot retrieve Public IP")
							}

							publicIP, err := extractIPfromNicRes(pubipRes)
							if err != nil {
								fmt.Println("error while extracting Public IP from its generic resource ", err)
							}
							if len(publicIP) > 0 {
								EachVM.PublicIP = publicIP
							}

						}

					}

				}
				fmt.Printf("%v \n", EachVM)
			}
			wg.Done()
		}(subID)
		wg.Wait()

	}

}
