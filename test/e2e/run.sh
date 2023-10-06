#!/usr/bin/env bash
set -ex

INSTANCE_NAME="${INSTANCE_NAME:-"ranchhand-local-$USER"}"
INSTANCE_BLUEPRINT_ID="${INSTANCE_BLUEPRINT_ID:-ubuntu_16_04_2}"
SSH_KEY_FILE="${SSH_KEY_FILE:-${HOME}/.ssh/id_rsa_a4d238e594137d6a2ec652c68f7f0e6b}"
SSH_USER="${SSH_USER:-ubuntu}"
export AWS_REGION="us-east-1"

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

  local max_retries=20

  for retries in $(seq 0 ${max_retries}); do
    sleep 10

    local state=$(aws lightsail get-instance-state --instance-name $INSTANCE_NAME --query 'state.name' --output text)
    if [[ $state == "running" ]]; then
      break
    fi

    echo "$INSTANCE_NAME is not ready (state: $state), trying again in 10 secs"
  done

  if [ "$retries" -eq "${max_retries}" ]; then echo "$INSTANCE_NAME is not ready!"; exit 5; fi

  local ipaddr=$(aws lightsail get-instance --instance-name $INSTANCE_NAME --query 'instance.publicIpAddress' --output text)
  echo $ipaddr > instance-ip

  local private_ipaddr=$(aws lightsail get-instance --instance-name $INSTANCE_NAME --query 'instance.privateIpAddress' --output text)
  echo $private_ipaddr > private-instance-ip

  for retries in $(seq 0 ${max_retries}); do
    sleep 10

    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ${SSH_KEY_FILE} ${SSH_USER}@${ipaddr} exit && break || true
    echo "${ipaddr} ssh connection timeout. Sleeping for 10 seconds..."
  done

  if [ "$retries" -eq "${max_retries}" ]; then echo "$INSTANCE_NAME not SSH ready!"; exit 5; fi

  aws lightsail open-instance-public-ports \
    --port-info fromPort=443,toPort=443,protocol=tcp \
    --instance-name $INSTANCE_NAME
  aws lightsail open-instance-public-ports \
    --port-info fromPort=6443,toPort=6443,protocol=tcp \
    --instance-name $INSTANCE_NAME
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
