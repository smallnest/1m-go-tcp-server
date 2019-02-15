#!/bin/bash
## 移除所有的测试客户端， 名称以 1mclient_ 开头的容器都会被删除
## 小心使用，别删错了

docker rm -vf  $(docker ps --format '{{.ID}} {{.Names}}'|grep '1mclient_' |awk '{print $1}')