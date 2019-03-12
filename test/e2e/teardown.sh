#!/usr/bin/env bash
set -ex

INSTANCE_NAME="${CIRCLE_JOB:-manual}-${CIRCLE_BUILD_NUM:-$USER}"

aws lightsail delete-instance --instance-name $INSTANCE_NAME
