[![Build Status](https://travis-ci.org/NeowayLabs/clinit-cfn-tool.svg?branch=master)](https://travis-ci.org/NeowayLabs/clinit-cfn-tool)

clinit-cfn-tool
===============

Cloudinit inject/extract into/from AWS CloudFormation

Install
===========

```
$ go get github.com/NeowayLabs/clinit-cfn-tool
$ GOPATH/bin/clinit-cfn-tool
```

Usage
===========

```
cat > ./test-aws.tpl.json
{
    "Resources": {
        "MasterInstance": {
            "Type": "AWS::EC2::Instance",
            "UserData": {{ .CloudInitData }}
        }
    }
}

cat > ./cloudconfig.yml
#cloud-config

hostname: test.neoway.com.br
coreos:
    etcd:
        discovery_url: http://example.com/4562338

```

Run:

```
clinit-cfn-tool inject -c ./cloudconfig.yml -f ./test-aws.tpl.json
```
License: BSD
