#!/usr/bin/env bash
set -ex

INSTANCE_NAME="${CIRCLE_WORKFLOW_JOB_ID:-local-$USER}"

aws lightsail delete-instance --instance-name $INSTANCE_NAME
