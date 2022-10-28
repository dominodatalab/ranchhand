#!/usr/bin/env bash
set -ex

INSTANCE_NAME="${INSTANCE_NAME:-"ranchhand-local-$USER"}"
INSTANCE_AMI_ID="${INSTANCE_AMI_ID:-ubuntu_xenial}"
INSTANCE_TAGS="${INSTANCE_TAGS:-"{\"Environment\": \"Test\"}"}"
SSH_KEY_FILE="${SSH_KEY_FILE:-${HOME}/.ssh/id_rsa_5a19aee7ef984f2f68c3be7262f91d35}"
SSH_USER="${SSH_USER:-ubuntu}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

ssh-keygen -y -f "$SSH_KEY_FILE" > "$SSH_KEY_FILE.pub"

function setup_instance() {
  cd "$SCRIPT_DIR/terraform" || exit 1
  terraform init
  terraform apply -auto-approve -var name="$INSTANCE_NAME" \
    -var public_key="$SSH_KEY_FILE.pub" \
    -var tags="$INSTANCE_TAGS" \
    -var ami_name="$INSTANCE_AMI_ID"

  local max_retries=20

  ipaddr=$(terraform output -raw public_ip)
  echo "$ipaddr" > "$SCRIPT_DIR"/../instance-ip

  private_ipaddr=$(terraform output -raw private_ip)
  echo "$private_ipaddr" > "$SCRIPT_DIR"/../private-instance-ip

  for retries in $(seq 0 ${max_retries}); do
    sleep 10

    if ssh -o StrictHostKeyChecking=no -i "${SSH_KEY_FILE}" "${SSH_USER}@${ipaddr}" exit; then
      break
    fi

    echo "${ipaddr} ssh connection timeout. Sleeping for 10 seconds..."
  done

  if [ "$retries" -eq "${max_retries}" ]; then echo "$INSTANCE_NAME not SSH ready!"; exit 5; fi
}

function teardown_instance() {
  cd "$SCRIPT_DIR/terraform" || exit 1
  terraform destroy -auto-approve \
    -var name="$INSTANCE_NAME" \
    -var public_key="$SSH_KEY_FILE.pub" \
    -var tags="$INSTANCE_TAGS" \
    -var ami_name="$INSTANCE_AMI_ID"
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
