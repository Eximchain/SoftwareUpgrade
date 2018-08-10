Table of Contents
=================
   * [Introduction](#introduction)
   * [CreateConfig](#createconfig)
        * [Command line parameters](#createconfig-command-line-parameters)
   * [Upgrade](#upgrade)
        * [Command line parameters](#upgrade-command-line-parameters)
        * [JSON configuration file format](#json-configuration-file-format)
        * [Troubleshooting](#troubleshooting)
  

Introduction
==
This is a suite of two software (CreateConfig and Upgrade) written to help in upgrading, as well as adding software onto Eximchain nodes.

In order to run this, you'll need to build it.

To build this, you'll need to refer instructions given in the BUILD.md.

CreateConfig
==

CreateConfig is a tool created to transform Terraform's json output into a configuration suitable to be used for the upgrade tool.
To get Terraform's output in JSON format, run the following command:

```
terraform output -json > terraformoutput.txt
```
The above command causes Terraform to output its data is JSON format into the file terraformoutput.txt. You can use any filename.

An example of the CreateConfig invocation is as below:
```
CreateConfig template.json terraformoutput.txt upgrade.json
```


CreateConfig command line parameters
==

* First parameter is the filename of the template.
* Second parameter is the filename of the Terraform output in JSON format.
* Third parameter is the filename to write the output to.



Upgrade
==

Upgrade is a tool created to upgrade and add software to the target Eximchain nodes. It can be used

Upgrade command line parameters
==

* -debug Specifies debug mode. When this is specified, more debug information go into the debug log.
* -debug-log logfilename - specifies the name of the debug log to write to.
* -disable-file-verification - disables source file existence verification.
* -disable-target-dir-verification - disables target directory existence verification.
* -dry-run - Enables testing mode, doesn't perform actual action, but starts and stops the software running on remote nodes
* -failed-nodes - Specifies the filename to load/save nodes that failed to upgrade.
* -json jsonfilename - specifies the name of the JSON configuration file to read from. This must always be present.
* -mode - Specifies the operating mode - add, delete-rollback, resume-upgrade, rollback, upgrade (default: upgrade)
* -rollback-filename - Specifies the rollback filename for this session.
  * Mode: add, adds the specified software in the configuration to the target nodes.
  * Mode: delete-rollback, removes the rollback files on the target nodes (only for software upgraded, not for software added)
  * Mode: resume-upgrade, continues the previous upgrade.
  * Mode: rollback, the files specified in this session will be used to remove the upgraded software on the target nodes.
  * Mode: upgrade, upgrade the software on the target nodes.
* -help - brings up information about the parameters.

Example
    -json=~/Documents/GitHub/SoftwareUpgrade/LaunchUpgrade.json -debug=true -debug-log=~/EximchainUpgrade.log

This launches the upgrader telling it to read the upgrade information from the file LaunchUpgrade.json, and to enable debug log output to the EximchainUpgrade.log file in the user home directory.

The rollback-filename parameter allows target nodes to rollback to the state they were before being upgraded.

JSON configuration file format
==

The JSON configuration file is a JSON object, which consists of a number of objects (Three objects at the moment). 
1. The software object defines multiple objects that specifies start, stop and files to copy to the target node, as well as commands to execute on the target node. 

2. The common object defines the location of the ssh certificate, the username to use.

3. The groupnodes object lists nodes belonging to groups listed in the software object's child nodes.

The software object defines the software objects that are to be upgraded/added on the target nodes.
Each child software object can be arbitrarily named (**_____the same names must be used in the groupnodes children nodes___**). It has a start and a stop string, and a copy object. 

The start and stop string specifies commands to execute, in order to start and stop the software being upgraded.
The stop command is executed first.
Each Copy object has numbered objects starting from 0, or 1. Each numbered object has a Local_Filename, Remote_Filename, and a Permissions string.
The Local_Filename string specifies the filename of the file to copy from. The Remote_Filename specifies the destination on the target node to copy the file to. The Permissions string specifies the ownership of the copied file, and is applied after the file has been copied over to the target node.
After all numbered objects are copied, the start command is then executed.

Table of child software object properties.
| Property  	| Type  	| Description  	|
|---	|---	|---	|
| start  	| string  	| The command to execute, in order to start the software after being added/upgraded.  	|
| stop  	| string  	| The command to execute, in order to stop the software before being upgraded. May be empty if the software is to be added. 	|
| Copy  	| object  	| The file(s) to copy, in order to add/upgrade the software to/on the target node.  	|

Table of Copy object properties.
| Property  	| Type  	| Description  	|
|---	|---	|---	|
| Local_Filename  	| string  	| Full path to the file to copy.  	|
| Remote_Filename  	| string  	| Full path on the target node for the file to be copied to.  	|
| Permissions  	| string  	| A 4-digit permissions string.  	|
| preupgrade  	| array of strings  	| Command(s) to execute before the upgrade starts. If empty, no commands are executed. 	|
| postupgrade  	| arroy of strings  	| Command(s) to execute after the upgrade is completed. If empty, no commands are executed. 	|

Table of common object properties.
| Property  	| Type  	| Description  	|
|---	|---	|---	|
| ssh_cert  	| string  	| Filename of the SSH certificate used to SSH to target nodes.  	|
| ssh_username  	| string  	| Username used to SSH to target nodes.  	|
| group_pause_after_upgrade  	| string  	| Specifies the amount of time to delay after upgrading a software group. 1h5m3s would mean 1 hour 5 minute and 3 seconds. The amount of time to delay is specified using this nomenclature. 	|
| software_group  	| array of strings  	| Specifies the list of software that comprised this group. The software names used must be the same as those listed under the top level software object.  	|

Table of groupnode properties.
| Property  	| Type  	| Description  	|
|---	|---	|---	|
| Same name as used under the top-level software object.  	| array of string 	| Specifies the hostname of the target nodes.	| 

An example of the JSON configuration file format follows.

```
{
    "software": {
        "blockmetrics": {
            "start": "sudo supervisorctl start blockmetrics",
            "stop": "sudo supervisorctl stop blockmetrics",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/block-metrics.py",
                    "Remote_Filename": "/opt/quorum/bin/block-metrics.py",
                    "Permissions": "0644"
                }
            }
        },
        "bootnode": {
            "start": "sudo supervisorctl start bootnode",
            "stop": "sudo supervisorctl stop bootnode",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/bootnode",
                    "Remote_Filename": "/usr/local/bin/bootnode",
                    "Permissions": "0755"
                }
            }
        },
        "cloudwatchmetrics": {
            "start": "sudo supervisorctl start cloudwatchmetrics",
            "stop": "sudo supervisorctl stop cloudwatchmetrics",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/cloudwatch-metrics.sh",
                    "Remote_Filename": "/opt/quorum/bin/cloudwatch-metrics.sh",
                    "Permissions": "0644"
                }
            }
        },
        "constellation": {
            "start": "sudo supervisorctl start constellation",
            "stop": "sudo supervisorctl stop constellation",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/constellation-node",
                    "Remote_Filename": "/usr/local/bin/constellation-node",
                    "Permissions": "0755"
                }
            }
        },
        "consul": {
            "start": "sudo supervisorctl start consul",
            "stop": "sudo supervisorctl stop consul",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/consul",
                    "Remote_Filename": "/opt/consul/bin/consul",
                    "Permissions": "0755"
                }
            }
        },
        "crashconstellation": {
            "start": "sudo supervisorctl start crashconstellation",
            "stop": "sudo supervisorctl stop crashconstellation",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/crashcloudwatch.py",
                    "Remote_Filename": "/opt/quorum/bin/crashcloudwatch.py",
                    "Permissions": "0744"
                }
            }
        },
        "crashquorum": {
            "start": "sudo supervisorctl start crashquorum",
            "stop": "sudo supervisorctl stop crashquorum"
        },
        "quorum": {
            "start": "sudo supervisorctl start quorum",
            "stop": "sudo supervisorctl stop quorum",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/geth",
                    "Remote_Filename": "/usr/local/bin/geth",
                    "Permissions": "0755"
                }
            }
        },
        "vault": {
            "preupgrade": [""],
            "postupgrade": [""],
            "start": "sudo supervisorctl start vault",
            "stop": "sudo supervisorctl stop vault",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/vault",
                    "Remote_Filename": "/opt/vault/bin/vault",
                    "Permissions": "0755"
                }
            }
        }
    },
    "common": {
        "ssh_cert": "~/.ssh/quorum",
        "ssh_timeout": "1m",
        "ssh_username": "ubuntu",
        "group_pause_after_upgrade": "6m15s",
        "software_group": {
            "Quorum-Makers": [
                "blockmetrics",
                "consul",
                "constellation",
                "quorum"
            ],
            "Quorum-Observers": [
                "blockmetrics",
                "cloudwatchmetrics",
                "constellation",
                "consul",
                "crashconstellation",
                "crashquorum",
                "quorum"
            ],
            "Quorum-Validators": [
                "blockmetrics",
                "cloudwatchmetrics",
                "constellation",
                "consul",
                "crashconstellation",
                "crashquorum",
                "quorum"
            ],
            "Bootnodes": [
                "constellation",
                "consul",
                "bootnode"
            ],
            "VaultServers": [
                "consul",
                "vault"
            ]
        }
    },
    "groupnodes": {
        "Quorum-Makers": [
            "ec2-54-164-95-40.compute-1.amazonaws.com",
            "name3",
            "name4"
        ],
        "Quorum-Observers": [
            "ec2-52-201-244-132.compute-1.amazonaws.com",
            "name5",
            "name6"
        ],
        "Quorum-Validators": [
            "ec2-52-72-195-7.compute-1.amazonaws.com",
            "name7",
            "name8",
            "name9"
        ],
        "Bootnodes": [
            "ec2-54-166-128-218.compute-1.amazonaws.com",
            "name10..."
        ],
        "VaultServers": [
            "34.228.16.117",
            "moreIP, or DNS"
        ]
    }
}
```

The "Quorum-Makers" in "software_group" specifies that it consists of the "blockmetrics", "consul", "constellation" and "quorum" software.
The "Quorum-Makers" in "groupnodes" specifies that the hostnames are: "ec2-54-164-95-40.compute-1.amazonaws.com", "name3", "name4", and that the software in the "Quorum-Makers" in "software_group" will be deployed to these hostnames. _The software group name used in "software_group" and "groupnodes" must be the same, so that the application knows that the software specified in the the "software_group" is to be deployed to the nodes specified in the "groupnodes" under the same name._

Troubleshooting
==
By default, this software produces a debug log called Upgrade-debug.log at ~/, unless it is disabled.

Any errors should appear in the debug log.
