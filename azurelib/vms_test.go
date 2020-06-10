package azurelib

import (
        "context"
        "fmt"
        "testing"
        "time"
)

const subscriptionID = "282160c0-3c83-43f1-bff1-9356b1678ffb"

func GetAuthorizedclients(subscriptionID string) (client Clients, err error) {
        clients := GetNewClients(subscriptionID)
        client, err = AuthorizeClients(clients)
        return
}

//TestGetallVMS tests function GetallVMS
func TestGetallVMS(t *testing.T) {

        client, err := GetAuthorizedclients(subscriptionID)
        if err != nil {

                t.Errorf("Failed to authorize: %v", err)
        }
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        Vmlist, err := GetallVMS(client, ctx)
        if err != nil {
                t.Errorf("Failed to  get all VMs: %v", err)
        } else {
                t.Logf("GetallVMS successful")
                for i := 0; i < len(Vmlist); i++ {
                        fmt.Println("Virtual machine Name: ", *Vmlist[i].Name)
                }
        }
}

