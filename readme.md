# Agent

[![Build Status](https://travis-ci.com/auto-cdp/agent.svg?branch=master)](https://travis-ci.com/auto-cdp/agent)
![LICENSE](https://img.shields.io/badge/license-MIT-orange.svg)
![LICENSE](https://img.shields.io/badge/license-Anti%20996-blue.svg?style=flat-square)



## 前提条件
### 1. 代码包内容为`以模块名称命名的目录的压缩包`，可支持`zip` 、`tar.gz`格式，例如：
```shell
hfp-member.tar.gz
    hfp-member/
    ├── bin
    ├── config
    ├── data
    ├── lib
    └── logs
```
### 2. `code pattern`为相对路径，仅支持路径末尾处的`shell pattern`，且确保不要出现互相包含现象
- 正确示例
```json
{
    "codepattern":[
        "lib",
        "config/static",
        "config/template",
        "bin/*.sh",
        "logs/age?nt*.log"
    ]
}
```

- 错误示例
```json
{
    "codepattern":[
        "lib",
        "config/static",
        "config/static/index.js"
    ]
}
```
### 3. 支持新增文件夹和文件，前提是部署前更新service的code pattern
### 4. service的启动、关闭、重启要求service目录下有bin目录，且执行脚本位于其中
### 5. check功能依赖pid文件
## 数据通信格式

1. 后台发送的任务json格式：

*通道名称：*`cmd.agentid`

```json
{
      "taskid": 123,
      "executionid": 1,
      "serviceid": "",
      "serviceop": 0,
      "servicename": "",
      "serviceosuser": "",
      "servicemodulename":"",
      "servicedir": "",
      "serviceremotecode": "",
      "servicecodepattern": ["lib","config/static"],
      "servicecustompattern": ["lib/custom.jar","config/template"],
      "servicepidfile": "",
      "servicestartcmd": "",
      "servicestopcmd": ""
      
    }

```
`serviceop：`

- *0* deploy
- *1* upgrade
- *2* start
- *3* stop
- *4* restart
- *5* check

2. 优雅指令格式：

*通道名称：*`grace.agentid`

```json
   {
       "taskid":"123456",
       "agentid":"7422abbe-ada0-46f4-9b60-65c5c2e27a2d",
       "gracecmd": "SIGHUP"
   }
```

`gracecmd`包括:***SIGHUP*** 、***SIGTERM*** 、***SIGINT***

3. 所有操作返回格式

*通道名称：*`result.taskid`

```json
{
      "taskid": 123,
      "executionid": 1,
      "rcode": 0,
      "rmsg": "",
      "rsteps": [{"stepnum": 1,"stepmane": "check","stepstate": 0,"stepmsg": "","steptime": ""},
                      {"stepnum": 2,"stepmane": "backup","stepstate": 0,"stepmsg": "","steptime": ""}]  
    }

```

## shell脚本增加内容
1. env.sh
```shell
#!/bin/bash

export JAVA_HOME=/usr/java/jdk1.8.0_162

export MAIN_SERVER_OBJECT=com.afis.oper.AlarmWechat
export MAIN_SERVRt_OPTS=" -d64 -server -Xms1024M -Xmx1536M"
export HFP_SERVER_NAME='AlarmWechat'
export HFP_SERVER_MODULE='AlarmWechat'
export HFP_SERVER_HOME=$HOME/$HFP_SERVER_NAME
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

2. 新增meta.sh
```shell
generate_post_data()
{
  cat <<EOF
{
    "servicename":"$HFP_SERVER_NAME",
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
