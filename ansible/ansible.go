/*
Copyright 2019 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/

package ansible

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type ec2AnsibleEntry struct {
	Name string
	Host string
}

// BuildEC2Inventory creates an ansible inventory for EC2 instances
// Requires a region to []*ec2.Instance Map
// It can generate the inventory for public (default) or private dns names
func BuildEC2Inventory(ec2dump map[string][]*ec2.Instance, private bool) (string, error) {
	var ansibleInv string
	ansibleTemplate := `
	{{- range $key, $value := .}}

[{{ $key }}]
		{{- range $value}}
{{ .Name}} ansible_ssh_host={{.Host}}
		{{- end}}
	{{- end}}
	`
	dump := map[string][]ec2AnsibleEntry{}
	for r, d := range ec2dump {
		var regionData []ec2AnsibleEntry
		for _, i := range d {
			var ansibleHost string
			if private {
				ansibleHost = *i.PrivateDnsName
			} else {
				ansibleHost = *i.PublicDnsName
			}
			name, err := extractNamefromEC2Tags(i)
			if err != nil {
				// Ignore Blank Name instance
				continue
			}
			e := ec2AnsibleEntry{
				Name: name,
				Host: ansibleHost}
			if e.Host == "" {
				continue
			}
			regionData = append(regionData, e)
		}
		dump[r] = regionData
		regionData = []ec2AnsibleEntry{}
	}
	tmpl, err := template.New("ec2").Parse(ansibleTemplate)
	if err != nil {
		return ansibleInv, err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, &dump)
	if err != nil {
		return "", fmt.Errorf("Error executing ansible template")
	}
	return b.String(), nil
}

func extractNamefromEC2Tags(i *ec2.Instance) (string, error) {
	var name string
	for _, t := range i.Tags {
		if *t.Key == "Name" {
			name = *t.Value
			if name == "" {
				continue
			}
			//Make sure name has no spaces
			name = strings.Replace(name, " ", "", -1)
			return name, nil
		}
	}
	return name, fmt.Errorf("Invalid or Empty Name Tag")
}
