#!/bin/bash

# MIT License
#
# Copyright (c) 2020 Dmitrii Ustiugov, Plamen Petrov and EASE lab
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT="$( cd $DIR && cd .. && pwd)"
BINS=$ROOT/bin
CTRDCONFIGS=$ROOT/configs/gvisor-containerd
CNICONFIGS=$ROOT/configs/cni

sudo mkdir -p /etc/gvisor-containerd

cd $ROOT
git lfs pull

DST=/usr/local/bin

for BINARY in containerd-shim-runsc-v1 gvisor-containerd
do
  sudo cp $BINS/$BINARY $DST
done

sudo cp $CTRDCONFIGS/config.toml /etc/gvisor-containerd/

sudo mkdir -p /etc/cni/net.d

DST=/etc/cni/net.d

for CONFIG in 10-bridge.conf 99-loopback.conf
do
  sudo cp $CNICONFIGS/$CONFIG $DST
done
