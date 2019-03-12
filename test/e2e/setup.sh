#!/usr/bin/env bash
set -ex

INSTANCE_NAME="ranchhand-${CIRCLE_WORKFLOW_JOB_ID:-local-$USER}"

aws lightsail create-instances \
  --instance-names $INSTANCE_NAME \
  --availability-zone us-east-1a \
  --blueprint-id ubuntu_16_04_2 \
  --bundle-id medium_2_0 \
  --tags "key=Environment,value=Test key=BuildUrl,value=$CIRCLE_BUILD_URL"

COUNTER=0
while [[ $COUNTER -lt 10 ]]; do
  state=$(aws lightsail get-instance-state --instance-name $INSTANCE_NAME | jq -r '.state.name')
  if [[ $state == "running" ]]; then
    break
  fi
  let COUNTER+=1

  echo "$INSTANCE_NAME is not ready (state: $state), trying again in 5 secs"
  sleep 5
done

aws lightsail open-instance-public-ports \
  --port-info fromPort=443,toPort=443,protocol=tcp \
  --instance-name $INSTANCE_NAME
aws lightsail open-instance-public-ports \
  --port-info fromPort=6443,toPort=6443,protocol=tcp \
  --instance-name $INSTANCE_NAME
aws lightsail open-instance-public-ports \
  --port-info fromPort=2379,toPort=2379,protocol=tcp \
  --instance-name $INSTANCE_NAME

ipaddr=$(aws lightsail get-instance --instance-name $INSTANCE_NAME | jq '.instance.publicIpAddress')
echo $ipaddr > instance-ip
