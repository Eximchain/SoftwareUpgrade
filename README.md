Introduction
==
This is a software written to help in upgrading the software running on the Eximchain Blockchain nodes.

In order to run this, you'll need to build it.

To build this, you'll need to refer instructions given in the BUILD.md.

Command line parameters
==

-debug Specifies debug mode. When this is specified, more debug information go into the debug log.
-debug-log logfilename - specifies the name of the debug log to write to.
-json jsonfilename - specifies the name of the JSON configuration file to read from.
--help - brings up information about the parameters.

Example
    -json=/Users/chuacw/Documents/GitHub/SoftwareUpgrade/LaunchUpgrade.json -debug=true -debug-log=/Users/chuacw/EximchainUpgrade.log

This launches the upgrader telling it to read the upgrade inforamtion from the file LaunchUpgrade.json, and to enable debug log output to the EximchainUpgrade.log file.

JSON configuration file format
==

The JSON configuration file is a JSON object, which consists of x objects. The software object defines multiple objects that specifies start, stop and files to copy to the target node. 

The common object defines the location of the ssh certificate, the username to use.

The groupnodes object lists nodes belonging to groups listed in the software object child nodes.


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
        "ssh_cert": "/Users/chuacw/.ssh/quorum",
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

Troubleshooting
==
By default, this software produces a debug log called Upgrade-debug.log at ~/, unless it is disabled.

Any errors should appear in the debug log.
