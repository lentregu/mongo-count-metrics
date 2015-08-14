#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters"
    echo "use: compile.sh darwin-amd64|linux-amd64"
    exit -1
fi

echo "Compiling...."

if ! [ -d "./bin" ]; then
  mkdir ./bin
fi

if [ "$1" == "darwin-amd64" ]; then
    if ! [ -d "./bin/darwin-amd64" ]; then
        mkdir ./bin
    fi
    go build -o ./bin/darwin-amd64/mongo-count-coll
elif [ "$1" == "linux-amd64" ]; then
    if ! [ -d "./bin/linux-amd64" ]; then
        mkdir ./bin
    fi
    GOOS=linux GOARCH=amd64 go build -o ./bin/linux-amd64/mongo-count-coll
else
    echo "use: compile.sh darwin-amd64|linux-amd64"
    exit -1
fi

echo "End compiling"
