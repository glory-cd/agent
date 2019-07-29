# Agent

[![Build Status](https://travis-ci.com/auto-cdp/agent.svg?branch=master)](https://travis-ci.com/auto-cdp/agent)
![LICENSE](https://img.shields.io/badge/license-GPLv3-blue.svg)
![LICENSE](https://img.shields.io/badge/license-Anti%20996-blue.svg?style=flat-square)

agent is a part of cdp(a continuous deployment tool), which runs on the machine, manages some services , receives instructions from the server and do some operations with the services, for example: deploy 、update、start、stop、check and so on.

# Documentations

- not available now

# Prerequisite

- Go >= 1.11

# Getting started

## Getting agent

```shell
git clone https://github.com/auto-cdp/agent.git
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
    "host": "your-redis-address:your-redis-port",
    "maxidele": 3,
    "maxactive": 0,
    "timeout": 300
  },
  "rest": {
    "addr": "127.0.0.1:9527"
  },
    "upload":{
          "addr":"192.168.1.75:32749",
          "type":"http",
          "username":"admin",
          "password":"YWZpczIwMTk="
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

## Running agent

```shell
/your/agent/path/agent --etcd 192.168.1.151:2379,192.168.1.152:2379(etcd cluster endpoint)
```

## Accept the service registration

`Some scripts need to be called when your service starts,who push the metadata of the service to agent, the agent register it in etcd.`

- env.sh

```shell
#!/bin/bash

export JAVA_HOME=/usr/java/jdk1.8.0_162

export MAIN_SERVER_OBJECT=com.afis.oper.AlarmWechat
export MAIN_SERVRt_OPTS=" -d64 -server -Xms1024M -Xmx1536M"
export HFP_SERVER_MODULE='AlarmWechat'
export HFP_SERVER_HOME=$HOME/$HFP_SERVER_MODULE
export HFP_SERVER_BIN=$HFP_SERVER_HOME/bin
export HFP_SERVER_CONFIG=$HFP_SERVER_HOME/config
export HFP_SERVER_LIB=$HFP_SERVER_HOME/lib
export HFP_SERVER_LOG=$HFP_SERVER_HOME/logs
export HFP_SERVER_PID=$HFP_SERVER_BIN/$MAIN_SERVER_OBJECT.pid

CLASSPATH=.:$HFP_SERVER_CONFIG
cd $HFP_SERVER_LIB
for  lname in `ls -rt *.jar`
do
        CLASSPATH=$CLASSPATH:$HFP_SERVER_LIB"/"$lname
done
export CLASSPATH=$CLASSPATH

export CODEPATTERN='"lib", "bin/*.sh"'
export HFP_SERVER_STARTCMD="./startup.sh"
export HFP_SERVER_STOPCMD="./shutdown.sh"

cd $HFP_SERVER_HOME/bin 
. ./meta.sh

curl -i \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST --data "$(generate_post_data)" "http://127.0.0.1:9527/register" >> $HFP_SERVER_LOG/register.log
```

- meta.sh

```shell
generate_post_data()
{
  cat <<EOF
{
    "serviceosuser":"$USER",
    "servicedir":"$HFP_SERVER_HOME",
    "servicecodepattern":[$CODEPATTERN],
    "servicemodulename":"$HFP_SERVER_MODULE",
    "servicepidfile":"$HFP_SERVER_PID",
    "servicestartcmd":"$HFP_SERVER_STARTCMD",
    "servicestopcmd":"$HFP_SERVER_STOPCMD"
}
EOF
}
```



# Limits

- code package format

**Compressed package named with module name, `zip`, `tar.gz` formats are supported**

Suppose you have a test module,the compressed package must like this:

```shell
test.tar.gz
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

- RSS operations depend on the bin directory of deployed service, and the corresponding scripts are there

- The check operation relies on pid files

- You cannot deploy two same services under the same user

# License

Agent is under the GPL 3.0 license. See the [LICENSE](LICENSE) file for details.