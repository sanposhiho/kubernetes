#!/usr/bin/env bash

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/../../..

source "${KUBE_ROOT}/hack/lib/init.sh"
export PATH="${KUBE_ROOT}/third_party/etcd:${PATH}"
export KUBE_SCHEDULER_SIMULATOR_ETCD_URL=${KUBE_INTEGRATION_ETCD_URL}

checkEtcdOnPath() {
  kube::log::status "Checking etcd is on PATH"
  which etcd && return
  kube::log::status "Cannot find etcd on PATH."
  kube::log::status "Please see https://git.k8s.io/community/contributors/devel/sig-testing/integration-tests.md#install-etcd-dependency for instructions."
  kube::log::usage "You can use 'hack/install-etcd.sh' to install a copy in third_party/."
  return 1
}

CLEANUP_REQUIRED=
start_etcd() {
  kube::log::status "Starting etcd instance"
  CLEANUP_REQUIRED=1
  kube::etcd::start
  KUBE_SCHEDULER_SIMULATOR_ETCD_URL=${KUBE_INTEGRATION_ETCD_URL}
  kube::log::status "etcd started"
}

cleanup_etcd() {
  if [[ -z "${CLEANUP_REQUIRED}" ]]; then
    return
  fi
  kube::log::status "Cleaning up etcd"
  kube::etcd::cleanup
  CLEANUP_REQUIRED=
  kube::log::status "Clean up finished"
}

trap cleanup_etcd EXIT

start_etcd

PORT=1212 ENV=development KUBE_TIMEOUT="-timeout=3600s" ./bin/simulator
