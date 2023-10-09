#!/bin/bash
# Different unique functions
pushd ../examples/deployer
source /etc/profile && go build
popd
pushd ../examples/invoker
source /etc/profile && go build
popd
mkdir ../result

# Deploy functions
kn service delete --all
func=helloworld.json
sed -i 's/\"count\".*/\"count\": 1/' ../examples/deployer/$func
pushd ../
pushd ./examples/deployer
cp $func functions.json
popd
./examples/deployer/deployer
popd

# Invoke
i=1
pushd ../
# Message number
for j in 1 2 4 8 16 32 64
do
    # Repeat 
    for k in $(seq 1 5)
    do
        # wait pods to be terminated
        while [ $(kubectl get pods 2>/dev/null | wc -l) -ne 0 ];
        do
            sleep 5;
        done
        sleep 5
        # limit the max size of function instance
        kn service update helloworld-0 --scale-max $j
        # wait pods to be terminated
        while [ $(kubectl get pods 2>/dev/null | wc -l) -ne 0 ];
        do
            sleep 5;
        done
        sleep 5
        # invoke
        ./examples/invoker/invoker --wait 20 --time 1 --allreq $j
        mv rps*.csv result/${func}_${i}_${j}_$[k].csv
    done
    # flush os page cache
    sleep 2
    sudo bash /tmp/sss
    sleep 2
done
popd

