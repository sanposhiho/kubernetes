#!/usr/bin/env bash

OPENAPIFILE="k8sapiserver/openapi/zz_generated.openapi.go"

if [ -e $OPENAPIFILE ]; then
  echo "The OpenAPI file have already generated."
  exit 0
fi

cd kubernetes
make kube-apiserver
cp pkg/generated/openapi/zz_generated.openapi.go "../${OPENAPIFILE}"
