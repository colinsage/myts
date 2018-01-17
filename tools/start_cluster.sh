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
./_data/tsdb_node -config tools/conf/cluster/meta2/conf/meta.toml 2>>_data/cluster/meta2/node.log &
./_data/tsdb_node -config tools/conf/cluster/meta3/conf/meta.toml 2>>_data/cluster/meta3/node.log &

sleep 10s

echo "begin start data server ... "
./_data/tsdb_node -config tools/conf/cluster/data1/conf/data.toml 2>>_data/cluster/data1/node.log &
./_data/tsdb_node -config tools/conf/cluster/data2/conf/data.toml 2>>_data/cluster/data2/node.log &
./_data/tsdb_node -config tools/conf/cluster/data3/conf/data.toml 2>>_data/cluster/data3/node.log &


echo "All start done !"

sleep 5s
curl -i -XPOST http://localhost:7086/query --data-urlencode "q=CREATE DATABASE mydb"

curl -i -XPOST http://localhost:7086/write?db=mydb --data-binary "cpu_load_long,host=server-01,region=us-west value=11 1514055567000000000"
curl -i -XPOST http://localhost:7086/write?db=mydb --data-binary "cpu_load_long,host=server-02,region=us-west value=12 1514055567000000000"

curl -G "http://localhost:9086/query?pretty=true" --data-urlencode "db=mydb" --data-urlencode "q=SELECT * FROM \"cpu_load_long\""

curl -G "http://localhost:9086/query?pretty=true" --data-urlencode "db=mydb" --data-urlencode "q=EXPLAIN SELECT * FROM \"cpu_load_long\""