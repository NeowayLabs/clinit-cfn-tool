[![Build Status](https://travis-ci.org/NeowayLabs/clinit-cfn-tool.svg?branch=master)](https://travis-ci.org/NeowayLabs/clinit-cfn-tool) [![Build Status](https://drone.io/github.com/NeowayLabs/clinit-cfn-tool/status.png)](https://drone.io/github.com/NeowayLabs/clinit-cfn-tool/latest)

clinit-cfn-tool
===============

Cloudinit inject/extract into/from AWS CloudFormation.

The motivation for create this tool was the very annoying/hard way of work with
[AWS CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/gettingstarted.templatebasics.html) integrated with CloudInit [Cloud-Config](http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data) user-data files.

AWS CloudFormation uses the JSON format while cloud-config uses YAML format. Cloudinit provides a standard for server initialization in the cloud, but there's no standard followed by the Cloud Providers (AWS, GCE, etc) for the stack bootstrap. This tool only try to solve your problems when using AWS CloudFormation!

The example below is a very simple cloud-config user-data file:

Eg.: examples/cloud-config.yml
```YAML
# Add groups to the system
# The following example adds the ubuntu group with members foo and bar and
# the group cloud-users.
groups:
  - ubuntu: [foo,bar]
  - cloud-users

# Add users to the system. Users are added after groups are added.
users:
  - default
  - name: foobar
    gecos: Foo B. Bar
    primary-group: foobar
    groups: users
    selinux-user: staff_u
    expiredate: 2012-09-01
    ssh-import-id: foobar
    lock-passwd: false
    passwd: $6$j212wezy$7H/1LT4f9/N3wpgNunhsIqtMj62OKiS3nyNwuizouQc3u7MbYCarYeAHWYPYb2FT.lbioDm2RrkJPb9BZMN1O/
  - name: barfoo
    gecos: Bar B. Foo
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin
    ssh-import-id: None
    lock-passwd: true
    ssh-authorized-keys:
      - <ssh pub key 1>
      - <ssh pub key 2>
  - name: cloudy
    gecos: Magic Cloud App Daemon User
    inactive: true
    system: true
```

For this example work with AWS CloudFormation, we need write something like the following JSON file:
eg.: examples/cfn1.json
```JSON
{
    "Resources": {
	"MasterInstance": {
	    "Type": "AWS::EC2::Instance",
	    "Properties": {
		"UserData": { "Fn::Base64": {"Fn::Join" : ["", [
		    "# Add groups to the system\n",
		    "# The following example adds the ubuntu group with members foo and bar and\n",
		    "# the group cloud-users.\n",
		    "groups:\n",
		    "  - ubuntu: [foo,bar]\n",
		    "  - cloud-users\n",
		    "\n",
		    "# Add users to the system. Users are added after groups are added.\n",
		    "users:\n",
		    "  - default\n",
		    "  - name: foobar\n",
		    "    gecos: Foo B. Bar\n",
		    "    primary-group: foobar\n",
		    "    groups: users\n",
		    "    selinux-user: staff_u\n",
		    "    expiredate: 2012-09-01\n",
		    "    ssh-import-id: foobar\n",
		    "    lock-passwd: false\n",
		    "    passwd: $6$j212wezy$7H/1LT4f9/N3wpgNunhsIqtMj62OKiS3nyNwuizouQc3u7MbYCarYeAHWYPYb2FT.lbioDm2RrkJPb9BZMN1O/\n",
		    "  - name: barfoo\n",
		    "    gecos: Bar B. Foo\n",
		    "    sudo: ALL=(ALL) NOPASSWD:ALL\n",
		    "    groups: users, admin\n",
		    "    ssh-import-id: None\n",
		    "    lock-passwd: true\n",
		    "    ssh-authorized-keys:\n",
		    "      - <ssh pub key 1>\n",
		    "      - <ssh pub key 2>\n",
		    "  - name: cloudy\n",
		    "    gecos: Magic Cloud App Daemon User\n",
		    "    inactive: true\n",
		    "    system: true\n"
		]]}
			    }
	    }
	}
    }
}

```

I don't know an easy/correct way to maintain this two files in sync. One choice is convert the AWS file to YAML and embed the cloud-config inside, start versioning this unique file and convert them to JSON when needed. But have the drawback that you is forced to copy-and-paste the UserData section of the config between your formation files.

Ok. But... How this project can be useful? See the sections below.

Install
===========

```
$ go get github.com/NeowayLabs/clinit-cfn-tool
$ GOPATH/bin/clinit-cfn-tool
```

Usage
=============

The first usage of clinit-cfn-tool is to extract the cloud-config from CloudFormation's file. For that we can use the cfn1.json in the example above and extract the UserData content. 

Run:
```
clinit-cfn-tool extract -i ./examples/cfn1.json -o ./cloud-config
Generating file './cloud-config1.yaml'
```
The command above will generate exactly the file ./examples/cloudconfig1.yml that we used to create the file examples/cfn1.json.

For now you can easily extract the UserData section of AWS CloudFormation of the internet.

The second usage is for the reverse, inject cloud-config inside a CloudFormation template file. For that, we will rename the file examples/cfn1.json to examples/cfn1.tpl.json and edit like this:
```JSON
{
    "Resources": {
	    "MasterInstance": {
	        "Type": "AWS::EC2::Instance",
	        "Properties": {
		        "UserData": {{ .CloudInitData }}
	        }
	    }
    }
}
```
Run:
```
clinit-cfn-tool inject -c ./cloudconfig1.yml -f ./examples/cfn1.tpl.json > ./cfn-output.json
```

More than one instance ? Ok, we have limitations here. The "extract" sub-command extract user-data of every Resource of type "AWS::EC2::Instance" and write to output files with format name ./{{name}}XX.yaml, where 'X' is a incremental number. But for inject we can't generate an output AWS CloudFormation file from multiples cloud-config files. For now, the clinit-cfn-tool is limited to generate a cfn with the maximum of two cloud-config, one called "master" and the other "node", in that case, the usage is like this:

CFN-Template: ./aws-master-node.tpl.json
```JSON
{
  "AWSTemplateFormatVersion": "{{ .CurrentDate }}",
  "Description": "{{ .Description }}",
  "Mappings": {
      "RegionMap": {
          "eu-central-1": {"AMI": "ami-54ccfa49"},
          "ap-northeast-1": {"AMI": "ami-f7b08ff6"},
          "sa-east-1": {"AMI": "ami-1304b30e"},
          "ap-southeast-2": {"AMI": "ami-0f117e35"},
          "ap-southeast-1": {"AMI": "ami-c04f6c92"},
          "us-east-1": {"AMI": "ami-7ae66812"},
          "us-west-2": {"AMI": "ami-e18dc5d1"},
          "us-west-1": {"AMI": "ami-45fbec00"},
          "eu-west-1": {"AMI": "ami-a27fd5d5"}
      }
  },
  "Parameters": {
    "InstanceType": {
      "Description": "EC2 HVM instance type (m3.medium, etc).",
      "Type": "String",
      "Default": "m3.medium",
      "AllowedValues": [
        "m3.medium",
        "m3.large",
        "m3.xlarge",
        "m3.2xlarge",
        "c3.large",
        "c3.xlarge",
        "c3.2xlarge",
        "c3.4xlarge",
        "c3.8xlarge",
        "cc2.8xlarge",
        "cr1.8xlarge",
        "hi1.4xlarge",
        "hs1.8xlarge",
        "i2.xlarge",
        "i2.2xlarge",
        "i2.4xlarge",
        "i2.8xlarge",
        "r3.large",
        "r3.xlarge",
        "r3.2xlarge",
        "r3.4xlarge",
        "r3.8xlarge",
        "t2.micro",
        "t2.small",
        "t2.medium"
      ],
      "ConstraintDescription": "Must be a valid EC2 HVM instance type."
    },
    "ClusterSize": {
      "Description": "Number of nodes in cluster (3-12).",
      "Default": "3",
      "MinValue": "3",
      "MaxValue": "12",
      "Type": "Number"
    },
    "AllowSSHFrom": {
      "Description": "The net block (CIDR) that SSH is available to.",
      "Default": "0.0.0.0/0",
      "Type": "String"
    },
    "KeyPair" : {
      "Description": "The name of an EC2 Key Pair to allow SSH access to the instance.",
      "Type": "String"
    }
  },
  "Resources": {
    "KubernetesSecurityGroup": {
      "Type": "AWS::EC2::SecurityGroup",
      "Properties": {
        "GroupDescription": "Kubernetes SecurityGroup",
        "SecurityGroupIngress": [
          {
            "IpProtocol": "tcp",
            "FromPort": "22",
            "ToPort": "22",
            "CidrIp": {"Ref": "AllowSSHFrom"}
          }
        ]
      }
    },
    "KubernetesIngress": {
      "Type": "AWS::EC2::SecurityGroupIngress",
      "Properties": {
        "GroupName": {"Ref": "KubernetesSecurityGroup"},
        "IpProtocol": "tcp",
        "FromPort": "1",
        "ToPort": "65535",
        "SourceSecurityGroupId": {
          "Fn::GetAtt" : [ "KubernetesSecurityGroup", "GroupId" ]
        }
      }
    },
    "KubernetesIngressUDP": {
      "Type": "AWS::EC2::SecurityGroupIngress",
      "Properties": {
        "GroupName": {"Ref": "KubernetesSecurityGroup"},
        "IpProtocol": "udp",
        "FromPort": "1",
        "ToPort": "65535",
        "SourceSecurityGroupId": {
          "Fn::GetAtt" : [ "KubernetesSecurityGroup", "GroupId" ]
        }
      }
    },
    "KubernetesMasterInstance": {
      "Type": "AWS::EC2::Instance",
      "Properties": {
        "ImageId": {"Fn::FindInMap" : ["RegionMap", {"Ref": "AWS::Region" }, "AMI"]},
        "InstanceType": {"Ref": "InstanceType"},
        "KeyName": {"Ref": "KeyPair"},
        "SecurityGroups": [{"Ref": "KubernetesSecurityGroup"}],
        "UserData": {{ .MasterCloudInitData }}
      }
    },
    "KubernetesNodeLaunchConfig": {
      "Type": "AWS::AutoScaling::LaunchConfiguration",
      "Properties": {
        "ImageId": {"Fn::FindInMap" : ["RegionMap", {"Ref": "AWS::Region" }, "AMI" ]},
        "InstanceType": {"Ref": "InstanceType"},
        "KeyName": {"Ref": "KeyPair"},
        "SecurityGroups": [{"Ref": "KubernetesSecurityGroup"}],
	"UserData": {{ .CloudInitData }}
      }
    },
    "KubernetesAutoScalingGroup": {
      "Type": "AWS::AutoScaling::AutoScalingGroup",
      "Properties": {
        "AvailabilityZones": {"Fn::GetAZs": ""},
        "LaunchConfigurationName": {"Ref": "KubernetesNodeLaunchConfig"},
        "MinSize": "3",
        "MaxSize": "12",
        "DesiredCapacity": {"Ref": "ClusterSize"}
      }
    }
  },
  "Outputs": {
    "KubernetesMasterPublicIp": {
    "Description": "Public Ip of the newly created Kubernetes Master instance",
      "Value": {"Fn::GetAtt": ["KubernetesMasterInstance" , "PublicIp"]}
    }
  }
}
```

Note the .MasterCloudInitData and .CloudInitData variables. For generate the formation for this two instances we can run:
```
clinit-cfn-tool inject -m ./master-cloudconfig.yml -c ./node-cloudconfig.yml -f ./aws-master-node.tpl.json > ./aws-master-node.json
```
If you really need generate a CloudFormation using more than two cloud-config files, we accept a PR for that =D

Found a bug? Open an issue [here](https://github.com/NeowayLabs/clinit-cfn-tool/issues)

License: BSD
