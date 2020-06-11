package azurelib

import (
        "context"
        "errors"
        "github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/network/mgmt/network"
        "github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
        "github.com/Azure/go-autorest/autorest/azure/auth"
        "strings"
)

//Clients is a struct that contains all the necessary clients
type Clients struct {
        //Network Interface Client
        VMInterface network.InterfacesClient
        //Public IP Addresses Client
        VMPublicIP network.PublicIPAddressesClient
        //Virtual Machine Client
        VMClient compute.VirtualMachinesClient
}
//VirtualMachineinfo is a  struct  that contains information related to a virtual machine
type VirtualMachineinfo struct {
        VM                  *compute.VirtualMachine
        PrivateIpaddress    *string
        PublicIpname        *string
        PublicIpaddress     *string
        VirtualnetandSubnet *string
        Ipconfig            *string
        Dns                 *string
}

//GetNewClients function returns a New Client
//Parameters - subscriptionID : Subscription ID for Azure
func GetNewClients(subscriptionID string) Clients {
        VMInterface := network.NewInterfacesClient(subscriptionID)
        VMPublicIP := network.NewPublicIPAddressesClient(subscriptionID)
        VMClient := compute.NewVirtualMachinesClient(subscriptionID)

        c := Clients{VMInterface, VMPublicIP, VMClient}
        return c
}

// AuthorizeClients function authorizes all the clients
func AuthorizeClients(c Clients) (Clients, error) {
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return c, err
        }
        c.VMClient.Authorizer = authorizer
        c.VMPublicIP.Authorizer = authorizer
        c.VMInterface.Authorizer = authorizer
        return c, nil
}

//GetallVMS function returns list of virtual machines
func GetallVMS(client Clients, ctx context.Context) (Vmlist []*VirtualMachineinfo, err error) {

        vmClient := client.VMClient
        results, err := vmClient.ListAllComplete(ctx)
        if err != nil {
                return
        }
        
        for results.NotDone() {
                var vminfo VirtualMachineinfo
                vm := results.Value()
                vminfo.VM = &vm
                vmresourceGroup, errvm := GetVMResourcegroup(&vm)
                if errvm != nil {
                        Vmlist = append(Vmlist, &vminfo)
                        continue
                }
                vmnetworkinterface, errvm := GetVmnetworkinterface(&vm)
                if errvm != nil {
                        Vmlist = append(Vmlist, &vminfo)
                        continue
                }
                vmprivateIPAddress, vmipconfig, errvm := GetPrivateIP(client, ctx, vmresourceGroup, vmnetworkinterface, "")
                if errvm == nil {
                        vminfo.PrivateIpaddress = &vmprivateIPAddress
                        vminfo.Ipconfig = &vmipconfig
                }

                vmvirtualnetandsubnet, errvm := GetSubnetandvirtualnetwork(client, ctx, vmresourceGroup, vmnetworkinterface, "")
                if errvm == nil {
                        vminfo.VirtualnetandSubnet = &vmvirtualnetandsubnet
                }
                vmdns, errvm := GetDNS(client, ctx, vmresourceGroup, vmnetworkinterface, "")
                if errvm == nil {
                        vminfo.Dns = &vmdns
                }
                vmpublicIpname, errvm := GetPublicIPAddressID(client, ctx, vmresourceGroup, vmnetworkinterface, "")
                if errvm == nil {
                        vminfo.PublicIpname = &vmpublicIpname
                        vmpublicIpaddress, errvm := GetPublicIPAddress(client, ctx, vmresourceGroup, vmpublicIpname, "")
                        if errvm == nil {
                                vminfo.PublicIpaddress = &vmpublicIpaddress

                        }
                }

                Vmlist = append(Vmlist, &vminfo)
                if err = results.Next(); err != nil {
                        return
                }

        }
        return
}

//GetVMResourcegroup function returns resourcegroup to which the virtual machine belongs to
func GetVMResourcegroup(vm *compute.VirtualMachine) (resourceGroup string, err error) {

        if vm.ID != nil {
                s := strings.Split(*vm.ID, "/")
                resourceGroup = s[4]
                err = nil
        } else {
                err = errors.New("No resourceGroup")
        }
        return
}

//GetVMname function returns the virtual machine's name
func GetVMname(vm *compute.VirtualMachine) (Name string, err error) {

        if vm.ID != nil {
                s := strings.Split(*vm.ID, "/")
                Name = s[8]
                err = nil
        } else {
                err = errors.New("No vm name")
        }
        return
}

//GetVMSubscription function returns the subscription ID
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

//GetVMTags function returns the tags related to the virtual machine
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
func GetVMSize(vm *compute.VirtualMachine) (Vmsize compute.VirtualMachineSizeTypes) {

        Vmsize = vm.VirtualMachineProperties.HardwareProfile.VMSize
        return

}

// GetVMOsType function returns the OStype used in the virtual machine
func GetVMOsType(vm *compute.VirtualMachine) (VMOS compute.OperatingSystemTypes) {

        VMOS = vm.VirtualMachineProperties.StorageProfile.OsDisk.OsType
        return
}

//GetVMadminusername function returns Virtual machine's adminusername
func GetVMadminusername(vm *compute.VirtualMachine) (VMadminusername string, err error) {
        if vm.VirtualMachineProperties.OsProfile.AdminUsername != nil {
                VMadminusername = *vm.VirtualMachineProperties.OsProfile.AdminUsername
                err = nil
        } else {
                err = errors.New("Vm has no admin user name")
        }
        return
}

//GetVmnetworkinterface function returns network interface of the Virtual machine
func GetVmnetworkinterface(vm *compute.VirtualMachine) (networkInterface string, err error) {
        if vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces != nil {
                networkinterface := *vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces
                netinterface := *networkinterface[0].ID
                ID := strings.Split(netinterface, "/")
                networkInterface = ID[8]
                err = nil
        } else {
                err = errors.New("Vm has no network interface")
        }
        return

}

//GetPrivateIP function returns Private IP Address of a Virtual Machine
func GetPrivateIP(client Clients, ctx context.Context,
        resourceGroup string, networkInterface string, expand string) (PrivateIPAddress string,
        IPConfiguration string, err error) {
        vmInterface := client.VMInterface
        interfaces, err := vmInterface.Get(ctx, resourceGroup, networkInterface, expand)
        if err != nil {
                return
        }
        interfaceinfo := *interfaces.InterfacePropertiesFormat.IPConfigurations
        interfID := *interfaceinfo[0].InterfaceIPConfigurationPropertiesFormat
        IPConfiguration = *interfaceinfo[0].Name
        if interfID.PrivateIPAddress != nil {
                PrivateIPAddress = *interfID.PrivateIPAddress
        }
        return
}

//GetPublicIPAddressID function returns Public IP Address ID (PublicIPName)
func GetPublicIPAddressID(client Clients,
        ctx context.Context, resourceGroup string, networkInterface string,
        expand string) (PublicIPAddressID string, err error) {
        vmInterface := client.VMInterface
        interfaces, err := vmInterface.Get(ctx, resourceGroup, networkInterface, expand)
        if err != nil {
                return
        }
        interfaceinfo := *interfaces.InterfacePropertiesFormat.IPConfigurations
        interfID := *interfaceinfo[0].InterfaceIPConfigurationPropertiesFormat

        if interfID.PublicIPAddress != nil && interfID.PublicIPAddress.ID != nil {
                ID := strings.Split(*interfID.PublicIPAddress.ID, "/")
                PublicIPAddressID = ID[8]
        } else {
                err = errors.New("Vm has no publicIPname")
        }
        return
}

//GetPublicIPAddress function returns the PublicIPAddress of the virtual machine
func GetPublicIPAddress(client Clients, ctx context.Context,
        resourceGroup string, PublicIPname string, expand string) (PublicIPAddress string, err error) {
        vmPublicIP := client.VMPublicIP
        VMIP, err := vmPublicIP.Get(ctx, resourceGroup, PublicIPname, expand)
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

//GetSubnetandvirtualnetwork function returns the virtual network and subnet
func GetSubnetandvirtualnetwork(client Clients,
        ctx context.Context, resourceGroup string, networkinterface string, expand string) (virtualnetworkandsubnet string, err error) {
        vmInterface := client.VMInterface
        interfaces, err := vmInterface.Get(ctx, resourceGroup, networkinterface, expand)
        if err != nil {
                return
        }
        interfaceinfo := *interfaces.InterfacePropertiesFormat.IPConfigurations
        interfID := *interfaceinfo[0].InterfaceIPConfigurationPropertiesFormat
        if interfID.Subnet != nil {
                ID := strings.Split(*interfID.Subnet.ID, "/")
                virtualnetworkandsubnet = ID[8] + "/" + ID[10]
        } else {
                err = errors.New("Vm has no virtual network and subnet")
        }
        return
}

//GetDNS function returns  DNS's Fqdn
func GetDNS(client Clients, ctx context.Context,
        resourceGroup string, PublicIPname string, expand string) (Fqdn string, err error) {
        vmPublicIP := client.VMPublicIP
        VMIP, err := vmPublicIP.Get(ctx, resourceGroup, PublicIPname, expand)
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