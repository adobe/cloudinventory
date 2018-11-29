package ansible

import (
	"bytes"
	"text/template"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type ec2AnsibleEntry struct {
	Name string
	Host string
}

// BuildEC2Inventory creates an ansible inventory for EC2 instances
func BuildEC2Inventory(ec2dump map[string][]*ec2.Instance) (string, error) {
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
			e := ec2AnsibleEntry{
				Name: extractNamefromEC2Tags(i),
				Host: *i.PublicDnsName}
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
	return b.String(), nil
}

func extractNamefromEC2Tags(i *ec2.Instance) string {
	var name string
	for _, t := range i.Tags {
		if *t.Key == "Name" {
			name = *t.Value
			return name
		}
	}
	return name
}
