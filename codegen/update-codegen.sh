#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

TEMP_DIR=$(mktemp -d)
SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=$TEMP_DIR/cgen

export GO111MODULE=on
if [[ -z "${GOPATH}" ]]; then
    export GOPATH=~/go
fi

git clone -b kubernetes-1.15.6 "https://github.com/kubernetes/code-generator" "$CODEGEN_PKG"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
"${CODEGEN_PKG}"/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/nais/alerterator/pkg/client github.com/nais/alerterator/pkg/apis \
  alerterator:v1 \
  --output-base "${TEMP_DIR}" \
  --go-header-file "${SCRIPT_ROOT}/codegen/boilerplate.go.txt"

rsync -av "${TEMP_DIR}/github.com/nais/alerterator/" "$SCRIPT_ROOT/"

rm -rf "$TEMP_DIR"

# To use your own boilerplate text use:
#   --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt
