# (W)ork(b)ench (I)nstaller 

## Getting started

wbi is a CLI tool aimed at streamlining the installation and configuration of Posit Workbench. Please read through the assumptions below to ensure that your target architecture matches the current capabilities of wbi. Also, note wbi requires to be run as root.

## Installation

### Apt Install

To install wbi on Ubuntu 22.04 or 20.04 using apt install:
```
echo "deb [trusted=yes] https://apt.fury.io/wbi/ /" | sudo tee -a /etc/apt/sources.list.d/fury.list && sudo apt update && sudo apt install wbi

sudo wbi setup
```

### Yum Install

To install wbi on RHEL 7/CentOS 7, RHEL 8/CentOS 8 or RHEL 9/CentOS 9 using yum install:
```
sudo tee -a /etc/yum.repos.d/fury.repo > /dev/null <<EOT
[fury]
name=wbi
baseurl=https://yum.fury.io/wbi/
enabled=1
gpgcheck=0
EOT

sudo yum install -y wbi

sudo wbi setup
```

### Manual Install

Visit the [release page](https://github.com/sol-eng/wbi/releases) to find install instructions for the latest release.

## Assumptions
- Single server
- SQLite database
- Internet access (online installation)

## Supported Operating Systems
- RHEL 9/CentOS 9
- RHEL 8/CentOS 8
- RHEL 7/CentOS 7
- Ubuntu 22.04
- Ubuntu 20.04

## Usage

### Interactive Prompts

To get started, run the setup command as root and follow the prompts:
```
sudo wbi setup
```

You can also pass the `--step` flag to begin at a certain spot in the interactive flow. For example, to start at the Workbench installation step:
```
sudo wbi setup --step workbench
```

The following steps are valid options: start, prereqs, firewall, security, languages, r, python, workbench, license, quarto, jupyter, prodrivers, ssl, packagemanager, connect, restart, status, verify.

### Individual Commands

wbi has individual commands to simplify different parts of the installation and configuration process. The complete list is outlined below. To get more information and examples, please use the `--help` flag (for example, for more information about the `install` command use `wbi install --help`).

#### activate

`wbi activate license`

#### config

`wbi config ssl`  
`wbi config repo`  
`wbi config connect-url`  

#### install

`wbi install r`  
`wbi install python`  
`wbi install quarto`  
`wbi install workbench`  
`wbi install prodrivers`  
`wbi install jupyter`  

#### scan

`wbi scan r`  
`wbi scan python`

#### verify

`wbi verify packagemanager`  
`wbi verify connect-url`  
`wbi verify workbench`  
`wbi verify ssl`  
`wbi verify license`  

### Command Log

A timestamped bash script will be generated in the same directory as `wbi` containing a record of each command executed. This is especially helpful if you wish to repeat the same setup process on another machine by running this script. Please note that this script is only to be used on an identical machine as `wbi` was run on (same OS, users, etc.)

## Support

**IMPORTANT:**

wbi is provided as a convenience to Posit customers. If you have
questions about this tool, you can ask them in the
[issues](https://github.com/sol-eng/wbi/issues/new) in the repository
or to your support representative, who will route them appropriately.

Bugs or feature requests should be opened in an [issue](https://github.com/sol-eng/wbi/issues/new).

### Logs

wbi will output detailed log information in a timestamped file in the same directory as `wbi`. If you are encountering issues with using `wbi` please refer to the logs and if reaching out for help, include the logs (after removing any sensitive data).

## License

[MIT License](./LICENSE)