#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT="$( cd $DIR && cd .. && pwd)"
SCRIPTS=$ROOT/scripts

mkdir -p /tmp/vhive-logs
./scripts/cloudlab/setup_node.sh > >(tee -a /tmp/vhive-logs/setup_node.stdout) 2> >(tee -a /tmp/vhive-logs/setup_node.stderr >&2)
sleep 5
./scripts/setup_runtime.sh
sleep 3

MODE=$1
if [ -z "$MODE" ]; then
    MODE="image"
fi

$SCRIPTS/cluster/setup_worker_kubelet.sh > >(tee -a /tmp/vhive-logs/setup_worker_kubelet.stdout) 2> >(tee -a /tmp/vhive-logs/setup_worker_kubelet.stderr >&2)
sudo screen -dmS containerd bash -c "containerd > >(tee -a /tmp/vhive-logs/containerd.stdout) 2> >(tee -a /tmp/vhive-logs/containerd.stderr >&2)"
sudo PATH=$PATH screen -dmS firecracker bash -c "/usr/local/bin/firecracker-containerd --config /etc/firecracker-containerd/config.toml > >(tee -a /tmp/vhive-logs/firecracker.stdout) 2> >(tee -a /tmp/vhive-logs/firecracker.stderr >&2)"
source /etc/profile && go build
if [ "$MODE" == "image" ]; then
    sudo screen -dmS vhive bash -c "./vhive > >(tee -a /tmp/vhive-logs/vhive.stdout) 2> >(tee -a /tmp/vhive-logs/vhive.stderr >&2)"
fi
if [ "$MODE" == "snapshot" ]; then
    sudo screen -dmS vhive bash -c "./vhive -snapshots> >(tee -a /tmp/vhive-logs/vhive.stdout) 2> >(tee -a /tmp/vhive-logs/vhive.stderr >&2)"
fi
if [ "$MODE" == "upf" ]; then
    sudo screen -dmS vhive bash -c "./vhive -snapshots -upf > >(tee -a /tmp/vhive-logs/vhive.stdout) 2> >(tee -a /tmp/vhive-logs/vhive.stderr >&2)"
fi
sleep 3
mkdir -p ~/.kube