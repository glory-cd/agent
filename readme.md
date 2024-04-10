# Agent

[![Build Status](https://travis-ci.com/glory-cd/agent.svg?branch=master)](https://travis-ci.com/glory-cd/agent)
![LICENSE](https://img.shields.io/badge/license-GPLv3-blue.svg)

agent is a part of glory-cd(a continuous deployment tool), which runs on the machine, manages some services , receives instructions from the server and do some operations with the services, for example: deploy 、update、start、stop、check and so on.

# Documentations

- not available now

# Prerequisite

- Go >= 1.11

# Getting started

## Getting agent

```shell
git clone https://github.com/glory-cd/agent.git
cd agent
go build
```

## etcd && redis

You must install etcd and redis first.

Then put  a key named `/agentConfig/template`in etcd:

```json
{
  "debug": true,
  "redis": {
    "host": "192.168.1.41:6133",
    "maxidele": 3,
    "maxactive": 0,
    "timeout": 300
  },
  "rest": {
    "addr": "127.0.0.1:9527"
  },
  "storeserver":{
        "addr":"192.168.1.75:32749",
        "type":"http",
        "username":"admin",
        "password":"YWZpczIwMTk=",
        "s3region":"",
        "s3bucket":""
    },
  "log": {
    "loglevel": "debug",
    "filename": "/var/log/agent.log",
    "maxsize": 128,
    "maxbackups": 30,
    "maxage": 7,
    "compress": true
  }
}

```
> note: The password field in storeserver is base64 encrypted.

## Running agent

```shell
/your/agent/path --etcd 192.168.1.151:2379,192.168.1.152:2379(etcd cluster endpoint)
```

## Accept the service registration

`Some scripts need to be called when your service starts,who push the metadata of the service to agent, the agent register it in etcd.`

`so, you must put the meta.sh to your *bin path*, and modify the first five variables.`


- meta.sh

```shell
#!/bin/bash

SERVER_PID_NAME=com.afis.oper.AlarmWechat
SERVER_MODULE='AlarmWechat'
SERVER_CODEPATTERN='"lib", "bin/*.sh"'
SERVER_STARTCMD="./startup.sh"
SERVER_STOPCMD="./shutdown.sh"


# get server path
SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BASENAME=$(basename "$SCRIPTPATH")

if [ "$BASENAME" == "bin" ] || [ "$BASENAME" == "BIN" ];then
  SERVERPATH=$(dirname $SCRIPTPATH)
else
  SERVERPATH=$SCRIPTPATH
fi

SERVER_HOME=$SERVERPATH
SERVER_LOG=$SERVER_HOME/logs
SERVER_PID_FILE=$SERVER_HOME/bin/$SERVER_PID_NAME.pid

# build json
generate_post_data()
{
  cat <<EOF
{
    "serviceosuser":"$USER",
    "servicedir":"$SERVER_HOME",
    "servicecodepattern":[$SERVER_CODEPATTERN],
    "servicemodulename":"$SERVER_MODULE",
    "servicepidfile":"$SERVER_PID_FILE",
    "servicestartcmd":"$SERVER_STARTCMD",
    "servicestopcmd":"$SERVER_STOPCMD"
}
EOF
}

# post metadata
curl -i \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(generate_post_data)" "http://127.0.0.1:9527/register" >> $SERVER_LOG/register.log
```

> note: readlink and curl are dependent

# Limits

- code package format

**Compressed package named with module name, only `zip` format is supported**

Suppose you have a test module,the compressed package must like this:

```shell
test.zip
    test/
    ├── bin
    ├── config
    ├── data
    ├── lib
    └── logs
```

- the `servicecodepattern`json column of communication protocol is relative path, and only the end of the path is supported(*shell pattern*), be careful not to include each other.

**The correct sample:**

```json
{
    "servicecodepattern":[
        "lib",
        "config/static",
        "config/template",
        "bin/*.sh",
        "logs/age?nt*.log"
    ]
}
```

**The wrong sample:**

```json
{
    "servicecodepattern":[
        "lib",
        "config/static",
        "config/static/index.js"
    ]
}
```

- Support for adding folders and files, the `servicecodepattern` field of service in the database must be updated before deployment

- The RSS operation assumes that the script is in bin or base directory.

- The check operation relies on pid files

# License

Agent is under the GPL 3.0 license. See the [LICENSE](LICENSE) file for details.
