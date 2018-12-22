# Cloud Inventory Gatherer

This package attempts to collect inventory of used services and resources across multiple clouds. It primarily wraps around the excellend SDKs already available but adds helper methods to aid in creation of inventories.
Current supported:

- AWS
  - EC2
  - RDS

(PRs welcome for more!)

## CLI

A command line tool is available and can be used as follows:

```bash
cloudinventory -h
Cloud Inventory is a wrapper around cloud provider SDKs to build a complete inventory for multiple services

Usage:
  cloudinventory [command]

Available Commands:
  dump        Dumps the inventory for the given options
  help        Help about any command

Flags:
  -h, --help   help for cloudinventory

Use "cloudinventory [command] --help" for more information about a command.
```

The subcommands all have additional options which be found by doing a `-h`
Example for AWS:

```bash
cloudinventory dump aws -h
Dump AWS inventory. Currently supports EC2/RDS

Usage:
  cloudinventory dump aws [flags]

Flags:
  -a, --ansible              Create a an ansible inventory as well (only for EC2)
      --ansible_inv string   File to create the EC2 ansible inventory in (default "ansible.inv")
      --ansible_private      Create Ansible Inventory with private DNS instead of public
  -h, --help                 help for aws
      --partition string     Which partition of AWS to run for default/china (default "default")

Global Flags:
  -f, --filter string   limit dump to a particular cloud service, e.g ec2/rds
  -p, --path string     file path to dump the inventory in (default "cloudinventory.json")
```

Available under `releases` and as a docker image at `tchaudhry/cloudinventory`

The tool reads credentials from your environment.

For AWS see: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html>

## Library Use

The packages with helping wrappers can be imported individually.
GoDocs available at:

[collector] (https://godoc.org/github.com/tchaudhry91/cloudinventory/collector)

[awslib] (https://godoc.org/github.com/tchaudhry91/cloudinventory/awslib)

## Contributing

Contributions are very welcome. Please see [Contributing Guide] (CONTRIBUTING.md) for more information

## Licensing

The project is licensed under the Apache V2 License. See [License] (LICENSE) for more information


Maintained By:
*tanmay.chaudhry@gmail.com*