#!/bin/bash

GOOS=("linux" "darwin")
GOARCH=("amd64" "amd64")

for((i=0;i<${#GOOS[@]};++i))
do
  _goos=${GOOS[$i]}
  _goarch=${GOARCH[$i]}
  _out=git-encrypt-agent
  _zip=${_out}_${_goos}_${_goarch}.zip
  GOOS=${_goos} GOARCH=${_goarch} go build -o ${_out}
  zip ${_zip} ${_out}
  rm -f ${_out}
done
