#!/bin/bash

echo "kill all exist ts_node process !!!"
pkill ts_node
sleep 5s

echo "begin build ..."
go build ./cmd/node/main.go
mv main ./_data/ts_node

echo "ts_node build done. "

mkdir -p _data/cluster/meta1
mkdir -p _data/cluster/meta2
mkdir -p _data/cluster/meta3

mkdir -p _data/cluster/data1
mkdir -p _data/cluster/data2
mkdir -p _data/cluster/data3


echo "begin start meta server ... "
./_data/ts_node -config tools/conf/cluster/meta1/conf/meta.toml 2>>_data/cluster/meta1/node.log &
./_data/ts_node -config tools/conf/cluster/meta2/conf/meta.toml 2>>_data/cluster/meta2/node.log &
./_data/ts_node -config tools/conf/cluster/meta3/conf/meta.toml 2>>_data/cluster/meta3/node.log &

sleep 10s

echo "begin start data server ... "
./_data/ts_node -config tools/conf/cluster/data1/conf/data.toml 2>>_data/cluster/data1/node.log &
./_data/ts_node -config tools/conf/cluster/data2/conf/data.toml 2>>_data/cluster/data2/node.log &
./_data/ts_node -config tools/conf/cluster/data3/conf/data.toml 2>>_data/cluster/data3/node.log &


echo "All start done !"