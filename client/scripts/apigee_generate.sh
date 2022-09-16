#!/bin/bash

export OUTDIR=/codegen/output
# set thes environment vars
export GO_POST_PROCESS_FILE="`command -v gofmt` -w"
export GO111MODULE=on

openapi-generator-cli version-manager set 4.3.1

node ./generate.js

# just in case, update all go imports
goimports -w=true /codegen/output

for f in "/codegen/output/model_*"; do
  sed -i -e 's/ int32 / int /g' $f
done

# timestamp is an int64 not a string
sed -i -r 's/Timestamp.string/Timestamp int64/g' /codegen/output/model_metrics_values.go

# replace the model_metrics_metrics.go file with the template for the custom unmarshal
cp ./scripts/model_metrics_metrics.tmpl /codegen/output/model_metrics_metrics.go