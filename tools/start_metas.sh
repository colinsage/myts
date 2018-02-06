#!/bin/bash

echo "kill all exist tsdb_node process !!!"
pkill tsdb_node
sleep 5s

rm -rf _data/*

echo "begin build ..."
go build -o _data/tsdb_node ./cmd/node/main.go


echo "tsdb_node build done. "

mkdir -p _data/cluster/meta1
mkdir -p _data/cluster/meta2
mkdir -p _data/cluster/meta3

mkdir -p _data/cluster/data1
mkdir -p _data/cluster/data2
mkdir -p _data/cluster/data3


echo "begin start meta server ... "
./_data/tsdb_node -config tools/conf/cluster/meta1/conf/meta.toml 2>>_data/cluster/meta1/node.log &
sleep 1
./_data/tsdb_node -config tools/conf/cluster/meta2/conf/meta.toml 2>>_data/cluster/meta2/node.log &
sleep 1
./_data/tsdb_node -config tools/conf/cluster/meta3/conf/meta.toml 2>>_data/cluster/meta3/node.log &

sleep 2s
echo "All start done !"

sleep 5s
#curl -i -XPOST http://localhost:7086/query --data-urlencode "q=CREATE DATABASE mydb"


curl "http://localhost:7091/dump"

curl "http://localhost:7091/leader"
