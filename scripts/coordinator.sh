#/bin/bash
MODE=$1
if [ -z "$MODE" ]; then
    echo "Please specify the mode of vHive"
    exit 1
fi

MASTER=$2
shift 2

echo Cloning vHive repository on master...
ssh Hidea@$MASTER git clone --branch test https://github.com/huasiy/vHive.git
echo Deploying master node...
res=$(ssh Hidea@$MASTER 'export DEBIAN_FRONTEND=noninteractive && cd vHive && ./scripts/deploy_master.sh')
s='sudo '
join=$s$(echo $res | grep -o 'kubeadm join.*')
echo The join command is $join
for host in $@
do
    ssh Hidea@$host git clone --branch test https://github.com/huasiy/vHive.git
    ssh Hidea@$host 'export DEBIAN_FRONTEND=noninteractive && cd vHive && ./scripts/deploy_worker.sh '$MODE
done

scp Hidea@$MASTER:~/.kube/config ./config
for host in $@
do
    scp ./config Hidea@$host:~/.kube
done

for host in $@
do
    ssh Hidea@$host ${join//\\}
done

ssh Hidea@$MASTER 'export DEBIAN_FRONTEND=noninteractive && cd vHive && ./scripts/cluster/setup_master_node.sh'
sleep 10
ssh Hidea@$MASTER kubectl get pods -A