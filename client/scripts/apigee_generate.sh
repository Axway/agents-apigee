#!/bin/bash

# set thes environment vars
export GO_POST_PROCESS_FILE="`command -v gofmt` -w"
export GO111MODULE=on

node ./scripts/generate.js

## just in case, update all go imports
goimports -w=true ./pkg/apigee/models

for f in "./pkg/apigee/models/model_*"; do
  sed -i -e 's/ int32 / int /g' $f
done
