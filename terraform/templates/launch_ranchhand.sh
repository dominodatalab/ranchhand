#!/usr/bin/env bash
# Downloads and launches ranchhand locally.
#
# If ssh_proxy_host is set, the behavior will change as follows:
#   1. rsync ranchhand assets onto the bastion
#   2. launch ranchhand on the bastion
#   3. rsync ranchhand output back onto the local machine
#
set -e

ranchhand="./ranchhand"
workdir="rancher-provision"
password="$RANCHER_PASSWORD"

ensure_workdir() {
  if [[ ! -d $workdir ]]; then
    mkdir -p $workdir
    echo "created dir: $workdir"
  fi
  cd $workdir

  echo "changed cwd: $(pwd)"
}

install_jq() {
  if ! command -v jq &> /dev/null; then
    local urlfrag="linux"

    if [[ $(uname -s) == "Darwin" ]]; then
      urlfrag="osx-amd"
    fi
    curl -sLo /usr/local/bin/jq "https://github.com/stedolan/jq/releases/download/jq-1.6/jq-$${urlfrag}64"
    chmod +x /usr/local/bin/jq

    echo "installed tool: jq"
  fi
}

install_ranchhand() {
  install_jq

  if [[ ! -x $ranchhand ]]; then
    local release_url

    if [[ "${release}" == "latest" ]]; then
      release_url="https://api.github.com/repos/dominodatalab/ranchhand/releases/latest"
    else
      release_url="https://api.github.com/repos/dominodatalab/ranchhand/releases/tags/${release}"
    fi

    local artifact_url="$(curl -s $release_url | jq -r '.assets[] | select(.browser_download_url | contains("${distro}")) | .browser_download_url')"
    curl -sL $artifact_url | tar xz

    echo "installed tool: ranchhand"
  fi
}

# TODO: remove this ssh_proxy_host logic once ranchhand has SSH proxy support
launch_ranchhand() {
  # ssh proxy host
  local ssh_proxy_host=${ssh_proxy_host}
  # ssh proxy user/host address
  local ssh_host_str="${ssh_proxy_user}@$ssh_proxy_host"
  # expand path with ~
  local ssh_key_path=${ssh_key_path}
  # ssh proxy args
  local ssh_args="-o LogLevel=ERROR -o StrictHostKeyChecking=no -i $ssh_key_path"
  # ssh proxy location of node ssh key
  local remote_key_path="$workdir/ssh_key"

  # pre-run file sync to proxy
  if [[ -n $ssh_proxy_host ]]; then
    ssh $ssh_args $ssh_host_str mkdir -p $workdir
    scp $ssh_args $ssh_key_path $ssh_host_str:$remote_key_path
    rsync -ai --rsh="ssh $ssh_args" ./ $ssh_host_str:$workdir
    echo "prepared bastion for run: $ssh_proxy_host"

    # override vars to dispatch call onto ssh proxy
    ranchhand="ssh $ssh_args $ssh_host_str cd $workdir && $ranchhand"
    ssh_key_path='$(pwd)/ssh_key'
    password="'$password'"
  fi

  echo "launching ranchhand"
  $ranchhand run \
    --node-ips "${node_ips}" \
    --cert-ips "${cert_ips}" \
    --cert-dns-names "${cert_dns_names}" \
    --ssh-user "${ssh_user}" \
    --ssh-key-path $ssh_key_path \
    --admin-password $password

  # post-run cleanup and file sync from proxy
  if [[ -n $ssh_proxy_host ]]; then
    ssh $ssh_args $ssh_host_str rm $remote_key_path
    rsync -ai --rsh="ssh $ssh_args" $ssh_host_str:$workdir/ ./
    echo "retrieved output from bastion: $ssh_proxy_host"
  fi

  echo "completed provisioning"
}

ensure_workdir
install_ranchhand
launch_ranchhand
