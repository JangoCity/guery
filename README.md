# Guery v0.1
Guery is a pure-go implementation of distributed SQL query engine for big data (like Presto). It is composed of one master and many executors and supports to query vast amounts of data using distributed queries.

## Status
This project is still a work in progress.

## Supported Datasource
### Hive(on hdfs)
### HDFS files

## Supported File type
### CSV
### Parquet


## Building Guery
```sh
make build
```
## Deploy Guery

### Run Master
```sh
#run master on 192.168.0.1
./guery master --address 192.168.0.1:1111 --config ./config.json >> m.log 
```

### Run Executor
```sh
#run 3 executors on 192.168.0.2
./guery executor --master 192.168.0.1:1111 --address 192.168.0.2:0 --config ./config.json >> e1.log
./guery executor --master 192.168.0.1:1111 --address 192.168.0.2:0 --config ./config.json >> e2.log
./guery executor --master 192.168.0.1:1111 --address 192.168.0.2:0 --config ./config.json >> e3.log
#run 3 executors on 192.168.0.3
./guery executor --master 192.168.0.1:1111 --address 192.168.0.3:0 --config ./config.json >> e1.log
./guery executor --master 192.168.0.1:1111 --address 192.168.0.3:0 --config ./config.json >> e1.log
./guery executor --master 192.168.0.1:1111 --address 192.168.0.3:0 --config ./config.json >> e1.log
```

## Query
```sh
curl -XPOST -d"sql=select * from hive.default.table1" 192.168.0.2:1111/query
```

## UI
