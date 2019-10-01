#!/usr/bin/env bash
set -e

INSTANCE_NAME="${INSTANCE_NAME:-"ranchhand-local-$USER"}"
INSTANCE_BLUEPRINT_ID="${INSTANCE_BLUEPRINT_ID:-ubuntu_16_04_2}"
SSH_KEY_FILE="${SSH_KEY_FILE:-${HOME}/.ssh/id_rsa_5a19aee7ef984f2f68c3be7262f91d35}"
SSH_USER="${SSH_USER:-ubuntu}"

function setup_instance() {
  if [[ -n $INSTANCE_TAGS ]]; then
    local tags=($INSTANCE_TAGS)
  else
    local tags=("key=Environment,value=Test")
  fi

  aws lightsail create-instances \
    --instance-names $INSTANCE_NAME \
    --availability-zone us-east-1a \
    --blueprint-id $INSTANCE_BLUEPRINT_ID \
    --bundle-id medium_2_0 \
    --tags ${tags[@]}

  local counter=0
  while [[ $counter -lt 12 ]]; do
    sleep 10

    local state=$(aws lightsail get-instance-state --instance-name $INSTANCE_NAME | jq -r '.state.name')
    if [[ $state == "running" ]]; then
      break
    fi
    let counter+=1

    echo "$INSTANCE_NAME is not ready (state: $state), trying again in 10 secs"
  done

  aws lightsail open-instance-public-ports \
    --port-info fromPort=443,toPort=443,protocol=tcp \
    --instance-name $INSTANCE_NAME
  aws lightsail open-instance-public-ports \
    --port-info fromPort=6443,toPort=6443,protocol=tcp \
    --instance-name $INSTANCE_NAME

  local ipaddr=$(aws lightsail get-instance --instance-name $INSTANCE_NAME \
    | jq --raw-output '.instance | .publicIpAddress')
  echo $ipaddr > instance-ip

  local private_ipaddr=$(aws lightsail get-instance --instance-name $INSTANCE_NAME \
    | jq --raw-output '.instance | .privateIpAddress')
  echo $private_ipaddr > private-instance-ip

  for i in {1..10}; do
    ssh -o StrictHostKeyChecking=no -i ${SSH_KEY_FILE} ${SSH_USER}@${ipaddr} exit && break || echo "${ipaddr} ssh connection timeout. Sleeping for 10 seconds..."
    sleep 10
  done
}

function teardown_instance() {
  aws lightsail delete-instance --instance-name $INSTANCE_NAME
}

if [[ $0 == $BASH_SOURCE ]]; then
  case "$1" in
    "setup")
      setup_instance
      ;;
    "teardown")
      teardown_instance
      ;;
    *)
      echo "Usage: $0 setup|teardown"
      exit 1
      ;;
  esac
fi
