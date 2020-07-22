package azurelib

import (
        "context"
        "errors"
        "github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
        "github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/network/mgmt/network"
        "github.com/Azure/go-autorest/autorest/azure/auth"
        "strings"
        "sync"
        "time"
)

// Clients is a struct that contains all the necessary clients
type Clients struct {
        // Subscription ID
        SubscriptionID string
        // Network Interface Client
        VMInterface network.InterfacesClient
        // Public IP Addresses Client
        VMPublicIP network.PublicIPAddressesClient
        // Virtual Machine Client
        VMClient compute.VirtualMachinesClient
}

// AdditionalInfo is a struct that contains extra information regarding the VM (obtained from other clients)
type AdditionalInfo struct {
        SubscriptionID      string
        ResourceGroup       string
        PrivateIPAddress    string
        PublicIPName        string
        PublicIPAddress     string
        VirtualnetAndSubnet string
        VNet                string
        SubNet              string
        IPConfig            string
        DNS                 string
}

// VirtualMachineInfo is a  struct  that contains information related to a virtual machine
type VirtualMachineInfo struct {
        Basic_info      *compute.VirtualMachine
        Additional_info AdditionalInfo
}

// GetNewClients function returns a New Client
// Parameters - subscriptionID : Subscription ID for Azure
func GetNewClients(subscriptionID string) Clients {
        VMInterface := network.NewInterfacesClient(subscriptionID)
        VMPublicIP := network.NewPublicIPAddressesClient(subscriptionID)
        VMClient := compute.NewVirtualMachinesClient(subscriptionID)

        c := Clients{subscriptionID, VMInterface, VMPublicIP, VMClient}
        return c
}

// AuthorizeClients function authorizes all the clients
func (c *Clients) AuthorizeClients() error {
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return err
        }
        c.VMClient.Authorizer = authorizer
        c.VMPublicIP.Authorizer = authorizer
        c.VMInterface.Authorizer = authorizer
        return nil
}

// getVMDetails funtion returns struct VirtualMachineInfo for a given virtual machine
func getVMDetails(ctx context.Context, client Clients, vm compute.VirtualMachine) *VirtualMachineInfo {
        var vmInfo VirtualMachineInfo
        vmInfo.Basic_info = &vm
        vmResourceGroup, errVM := GetVMResourceGroup(&vm)
        if errVM != nil {
                return &vmInfo
        }
        vmInfo.Additional_info.SubscriptionID = client.SubscriptionID
        vmInfo.Additional_info.ResourceGroup = vmResourceGroup
        vmNetworkInterface, errVM := GetVMNetworkInterface(&vm)
        if errVM != nil {
                return &vmInfo
        }
        vmPrivateIPAddress, vmIPConfig, errVM := GetPrivateIP(ctx, client.VMInterface, vmResourceGroup, vmNetworkInterface, "")
        if errVM == nil {
                vmInfo.Additional_info.PrivateIPAddress = vmPrivateIPAddress
                vmInfo.Additional_info.IPConfig = vmIPConfig
        }
        vmVirtualnetandSubnet, vNet, subNet, errVM := GetSubnetAndVirtualNetwork(ctx, client.VMInterface, vmResourceGroup, vmNetworkInterface, "")
        if errVM == nil {
                vmInfo.Additional_info.VirtualnetAndSubnet = vmVirtualnetandSubnet
                vmInfo.Additional_info.VNet = vNet
                vmInfo.Additional_info.SubNet = subNet
        }
        vmDNS, errVM := GetDNS(ctx, client.VMPublicIP, vmResourceGroup, vmNetworkInterface, "")
        if errVM == nil {
                vmInfo.Additional_info.DNS = vmDNS
        }
        vmPublicIPName, errVM := GetPublicIPAddressID(ctx, client.VMInterface, vmResourceGroup, vmNetworkInterface, "")
        if errVM == nil {
                vmInfo.Additional_info.PublicIPName = vmPublicIPName
                vmPublicIPAddress, errVM := GetPublicIPAddress(ctx, client.VMPublicIP, vmResourceGroup, vmPublicIPName, "")
                if errVM == nil {
                        vmInfo.Additional_info.PublicIPAddress = vmPublicIPAddress
                }
        }
        return &vmInfo
}

// GetAllVMS function returns list of virtual machines for a subscriptionID
func GetAllVMS(client Clients) (VMList []*VirtualMachineInfo, err error) {
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        vmClient := client.VMClient
        results, err := vmClient.ListAllComplete(ctx)
        if err != nil {
                return nil, err
        }
        chanCapacity := 1000
        isNotDone := true
        for isNotDone {

                instancesChan := make(chan *VirtualMachineInfo, chanCapacity)
                var wg sync.WaitGroup
                for vmCount := 0; results.NotDone() && vmCount < chanCapacity; vmCount++ {
                        wg.Add(1)
                        vm := results.Value()
                        go func(vm compute.VirtualMachine, client Clients, ctx context.Context, instancesChan chan *VirtualMachineInfo) {
                                defer wg.Done()
                                instancesChan <- getVMDetails(ctx, client, vm)
                        }(vm, client, ctx, instancesChan)

                        if err = results.Next(); err != nil {
                                return
                        }

                }
                isNotDone = results.NotDone()
                wg.Wait()
                close(instancesChan)
                for vmInfo := range instancesChan {
                        VMList = append(VMList, vmInfo)
                }

        }
        return
}

// GetVMCount function returns stats of virtual machines for a subscriptionID
func GetVMCount(subscriptionID string)(int, error) {
        var vmCount int
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return vmCount, err
        }
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        vmClient := compute.NewVirtualMachinesClient(subscriptionID)
        vmClient.Authorizer = authorizer
        result, err := vmClient.ListAllComplete(ctx)
        if err != nil {
                return vmCount, err
        }
        for result.NotDone(){
                vmCount++
                if err = result.Next(); err != nil {
                        return vmCount, err
                }
        }
        return vmCount, nil
}

// GetVMResourceGroup function returns resourcegroup to which the virtual machine belongs to
func GetVMResourceGroup(vm *compute.VirtualMachine) (resourceGroup string, err error) {

        if vm.ID != nil {
                s := strings.Split(*vm.ID, "/")
                resourceGroup = s[4]
                err = nil
        } else {
                err = errors.New("No resourceGroup")
        }
        return
}

// GetVMName function returns the virtual machine's name
func GetVMName(vm *compute.VirtualMachine) (Name string, err error) {

        if vm.ID != nil {
                s := strings.Split(*vm.ID, "/")
                Name = s[8]
                err = nil
        } else {
                err = errors.New("No vm name")
        }
        return
}

// GetVMSubscription function returns the subscription ID
func GetVMSubscription(vm *compute.VirtualMachine) (subscriptionID string, err error) {

        if vm.ID != nil {
                s := strings.Split(*vm.ID, "/")
                subscriptionID = s[2]
                err = nil
        } else {
                err = errors.New("No subscription")
        }
        return
}

// GetVMTags function returns the tags related to the virtual machine
func GetVMTags(vm *compute.VirtualMachine) (tags map[string]*string, err error) {
        if vm.Tags != nil {

                tags = vm.Tags
                err = nil
        } else {
                err = errors.New("no tags present for the vm")
        }
        return
}

// GetVMLocation function  returns the location
func GetVMLocation(vm *compute.VirtualMachine) (location string, err error) {

        if vm.Location != nil {
                location = *vm.Location
                err = nil
        } else {
                err = errors.New("no location assigned to the vm")
        }
        return
}

// GetVMSize function returns size of the virtual machine
func GetVMSize(vm *compute.VirtualMachine) (VMSize compute.VirtualMachineSizeTypes) {

        VMSize = vm.VirtualMachineProperties.HardwareProfile.VMSize
        return

}

// GetVMOsType function returns the OStype used in the virtual machine
func GetVMOsType(vm *compute.VirtualMachine) (VMOS compute.OperatingSystemTypes) {

        VMOS = vm.VirtualMachineProperties.StorageProfile.OsDisk.OsType
        return
}

// GetVMAdminUsername function returns Virtual machine's adminusername
func GetVMAdminUsername(vm *compute.VirtualMachine) (VMAdminUsername string, err error) {
        if vm.VirtualMachineProperties.OsProfile.AdminUsername != nil {
                VMAdminUsername = *vm.VirtualMachineProperties.OsProfile.AdminUsername
                err = nil
        } else {
                err = errors.New("Vm has no admin user name")
        }
        return
}

// GetVMNetworkInterface function returns network interface of the Virtual machine
func GetVMNetworkInterface(vm *compute.VirtualMachine) (networkInterface string, err error) {
        if vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces != nil {
                networkInterfaceID := *vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces
                netInterface := *networkInterfaceID[0].ID
                ID := strings.Split(netInterface, "/")
                networkInterface = ID[8]
                err = nil
        } else {
                err = errors.New("Vm has no network interface")
        }
        return

}

// GetPrivateIP function returns Private IP Address of a Virtual Machine
func GetPrivateIP(ctx context.Context, vmInterface network.InterfacesClient,
        resourceGroup string, networkInterface string, expand string) (PrivateIPAddress string,
        IPConfiguration string, err error) {
        interfaces, err := vmInterface.Get(ctx, resourceGroup, networkInterface, expand)
        if err != nil {
                return
        }
        interfaceInfo := *interfaces.InterfacePropertiesFormat.IPConfigurations
        interfID := *interfaceInfo[0].InterfaceIPConfigurationPropertiesFormat
        IPConfiguration = *interfaceInfo[0].Name
        if interfID.PrivateIPAddress != nil {
                PrivateIPAddress = *interfID.PrivateIPAddress
        }
        return
}

// GetPublicIPAddressID function returns Public IP Address ID (PublicIPName)
func GetPublicIPAddressID(ctx context.Context, vmInterface network.InterfacesClient,
        resourceGroup string, networkInterface string, expand string) (PublicIPAddressID string,
        err error) {
        interfaces, err := vmInterface.Get(ctx, resourceGroup, networkInterface, expand)
        if err != nil {
                return
        }
        interfaceInfo := *interfaces.InterfacePropertiesFormat.IPConfigurations
        interfID := *interfaceInfo[0].InterfaceIPConfigurationPropertiesFormat

        if interfID.PublicIPAddress != nil && interfID.PublicIPAddress.ID != nil {
                ID := strings.Split(*interfID.PublicIPAddress.ID, "/")
                PublicIPAddressID = ID[8]
        } else {
                err = errors.New("Vm has no publicIPname")
        }
        return
}

// GetPublicIPAddress function returns the PublicIPAddress of the virtual machine
func GetPublicIPAddress(ctx context.Context, vmPublicIP network.PublicIPAddressesClient,
        resourceGroup string, PublicIPName string, expand string) (PublicIPAddress string, err error) {
        VMIP, err := vmPublicIP.Get(ctx, resourceGroup, PublicIPName, expand)
        if err != nil {
                return
        }
        if VMIP.PublicIPAddressPropertiesFormat != nil && VMIP.PublicIPAddressPropertiesFormat.IPAddress != nil {
                PublicIPAddress = *VMIP.PublicIPAddressPropertiesFormat.IPAddress

        } else {
                err = errors.New("Vm has no publicIPAddress")
        }
        return

}

// GetSubnetAndVirtualNetwork function returns the virtual network and subnet
func GetSubnetAndVirtualNetwork(ctx context.Context, vmInterface network.InterfacesClient,
        resourceGroup string, networkInterface string, expand string) (virtualNetworkAndSubnet string,
        vNet string, subNet string, err error) {
        interfaces, err := vmInterface.Get(ctx, resourceGroup, networkInterface, expand)
        if err != nil {
                return
        }
        interfaceInfo := *interfaces.InterfacePropertiesFormat.IPConfigurations
        interfID := *interfaceInfo[0].InterfaceIPConfigurationPropertiesFormat
        if interfID.Subnet != nil {
                ID := strings.Split(*interfID.Subnet.ID, "/")
                virtualNetworkAndSubnet = ID[8] + "/" + ID[10]
                vNet = ID[8]
                subNet = ID[10]
        } else {
                err = errors.New("Vm has no virtual network and subnet")
        }
        return
}

// GetDNS function returns  DNS's Fqdn
func GetDNS(ctx context.Context, vmPublicIP network.PublicIPAddressesClient,
        resourceGroup string, PublicIPName string, expand string) (Fqdn string, err error) {
        VMIP, err := vmPublicIP.Get(ctx, resourceGroup, PublicIPName, expand)
        if err != nil {
                return
        }
        if VMIP.PublicIPAddressPropertiesFormat != nil && VMIP.PublicIPAddressPropertiesFormat.DNSSettings != nil {
                Fqdn = *VMIP.PublicIPAddressPropertiesFormat.DNSSettings.Fqdn
        } else {
                err = errors.New("DNS is not configured")
        }
        return
}