#!/bin/bash
# Different unique functions
pushd ../examples/deployer
source /etc/profile && go build
popd
pushd ../examples/invoker
source /etc/profile && go build
popd
mkdir ../result

func=helloworld.json
i=3
kn service delete --all
# Deploy functions
j=`expr $i - 1`
sed -i "s@\"count\": ${j}@\"count\": ${i}@g" ../examples/deployer/$func
pushd ../
pushd ./examples/deployer
cp $func functions.json
popd
./examples/deployer/deployer
popd
# Invoke
sleep 30
pushd ../
for j in $(seq 1 20)
do
    ./examples/invoker/invoker --wait 30 --time 1 --allreq $j
    mv rps*.csv result/${func}_${i}_${j}.csv
    # wait pods to be terminated
    sleep 180
    # flush os page cache
    timeout 90 stress-ng --vm-bytes $(awk '/MemAvailable/{printf "%d\n", $2 * 0.95;}' < /proc/meminfo)k --vm-keep -m 1
    sleep 10
done
popd
kn service delete --all

for func in pyaes.json rnn-serving.json
do
    for i in $(seq 1 3)
    do
        kn service delete --all
        # Deploy functions
        j=`expr $i - 1`
        sed -i "s@\"count\": ${j}@\"count\": ${i}@g" ../examples/deployer/$func
        pushd ../
        pushd ./examples/deployer
        cp $func functions.json
        popd
        ./examples/deployer/deployer
        popd
        # Invoke
        sleep 30
        pushd ../
        for j in $(seq 1 20)
        do
            ./examples/invoker/invoker --wait 30 --time 1 --allreq $j
            mv rps*.csv result/${func}_${i}_${j}.csv
            # wait pods to be terminated
            sleep 180
            # flush os page cache
            timeout 90 stress-ng --vm-bytes $(awk '/MemAvailable/{printf "%d\n", $2 * 0.95;}' < /proc/meminfo)k --vm-keep -m 1
            sleep 10
        done
        popd
        kn service delete --all
    done
done
