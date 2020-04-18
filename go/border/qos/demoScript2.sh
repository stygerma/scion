#!/bin/bash

# Copyright 2020 ETH Zurich
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
# 
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

currDir=$(pwd)
echo "$currDir"
cd $(dirname $0)

cd .. # Going to border
cd .. # Going to go
cd .. # Going to the SCION folder

interactive='false'
verbose='true'
optRatio='false'

printUseage() {
  echo "Usage:"
  echo "-q for quiet mode to suppress explanations for each of the steps"
  echo "-r will output the ratio to ratio.csv"
  echo "-i for interactive mode. Requires some keypresses to continue."
  exit 0
}

while getopts ':iqrh' flag; do
  case "${flag}" in
    i) interactive='true' ;;
    q) verbose='false' ;;
    r) optRatio='true' ;;
    h) printUseage
  esac
done

output() {
if $verbose; then
    echo "$1"
fi
}

printBlue() {
    tput setaf 4; output "$1"; tput sgr0;
}

waitForEnter() {
    if $interactive; then
        printBlue "Press enter to continue $1" 
        read -p ""
    fi
}

startNetcatListener() {
    SCION_DAEMON_ADDRESS='127.0.0.27:30255'
    export SCION_DAEMON_ADDRESS
    tail -f /dev/null | ./../scion-apps/netcat/netcat -l "$1" > ../scion-apps/netcat/data/server"$2".output
}

transferFileTo() {
    local start
    start=$(date +%s)
    SCION_DAEMON_ADDRESS="127.0.0.$1:30255"
    export SCION_DAEMON_ADDRESS
    ./../scion-apps/netcat/netcat 1-ff00:0:111,[127.0.0.1]:"$2" < ../scion-apps/netcat/data/test100Mb.db
    local end=`date +%s`
    local runtime=$((end-start))
    echo $runtime >  ".tempFile$2"
}

printBlue "Starting the demo"

# Generate topology and copy configuration files

output "Generate topology and copy configuration files"

./scion.sh topology -c topology/DemoTiny2.topo

cp go/border/qos/testdata/DemoConfig.yaml gen/ISD1/ASff00_0_110/br1-ff00_0_110-1/qosConfig.yaml
cp go/border/qos/testdata/DemoConfigEmpty.yaml gen/ISD1/ASff00_0_111/br1-ff00_0_111-1/qosConfig.yaml
cp go/border/qos/testdata/DemoConfigEmpty.yaml gen/ISD1/ASff00_0_112/br1-ff00_0_112-1/qosConfig.yaml
cp go/border/qos/testdata/DemoConfigEmpty.yaml gen/ISD1/ASff00_0_113/br1-ff00_0_113-1/qosConfig.yaml

# # # Start SCION
printBlue "Start SCION"

./scion.sh start
./scion.sh status
sleep 5

# # # # Do PING for 5 seconds AS110 to AS111
printBlue "AS110 to AS111"
./bin/scmp echo -local 1-ff00:0:110,[127.0.0.1] -remote 1-ff00:0:111,[0.0.0.0] -sciond 127.0.0.19:30255 -c 5
# # # # Do PING for 5 seconds AS110 to AS112
printBlue "AS110 to AS112"
./bin/scmp echo -local 1-ff00:0:110,[127.0.0.1] -remote 1-ff00:0:112,[0.0.0.0] -sciond 127.0.0.19:30255 -c 5
# # # # Do PING for 5 seconds AS111 to AS112
printBlue "AS111 to AS112"
./bin/scmp echo -local 1-ff00:0:111,[127.0.0.1] -remote 1-ff00:0:112,[0.0.0.0] -sciond 127.0.0.27:30255 -c 5

waitForEnter

printBlue "BWTester"

SCION_DAEMON_ADDRESS='127.0.0.27:30255'
export SCION_DAEMON_ADDRESS
./../scion-apps/bwtester/bwtestserver/bwtestserver -p 40101 &
pid0=$!

SCION_DAEMON_ADDRESS='127.0.0.19:30255'
export SCION_DAEMON_ADDRESS
./../scion-apps/bwtester/bwtestclient/bwtestclient -s 1-ff00:0:111,[127.0.0.1]:40101 -cs 10,1000,?,100Mbps -sc 1,1000,?,1Mbps

kill -9 $pid0

waitForEnter

# Make sure that no netcat processes are left running
killall netcat

# # Start netcat server 1 in AS111
startNetcatListener 34234 1 &
pid1=$!
# # Start netcat server 2 in AS111
startNetcatListener 35234 2 &
pid2=$!
# # Start netcat server 3 in AS111
startNetcatListener 36234 3 &
pid3=$!

# # Transfer File from AS110 to AS111 to show that 10 Mbit/s can be reached
SCION_DAEMON_ADDRESS='127.0.0.19:30255'
export SCION_DAEMON_ADDRESS
pv ../scion-apps/netcat/data/test100Mb.db | ./../scion-apps/netcat/netcat -vv 1-ff00:0:111,[127.0.0.1]:34234

waitForEnter

# Transfer File from AS110 to AS111 to show that the transfer is two times faster
transferFileTo 43 35234 &
pid4=$!
# Transfer File from AS112 to AS111 to show that it does not starve
start=$(date +%s)
transferFileTo 35 36234 &
pid5=$!

printBlue "Finished netcat sender setup. Waiting for them to finish..."


wait $pid4
printBlue "First transfer is done, kill the second one"
kill -9 $pid5
end2=`date +%s`
result2=$((end2-start))
printBlue "Second transfer is done!"

# Measure time of execution and divide by file size to get the transfer speed
# Show the transfer speeds, the ratio of the transfer speeds and the ratio attempted

result1=$(cat .tempFile35234)

FSSLOWERTRANSFER=$(stat -c%s ../scion-apps/netcat/data/server3.output)
txMB=$(printf %.2f $(echo "$FSSLOWERTRANSFER/1024/1024"| bc -l))
output "Managed to transfer $txMB MB"
txMbit=$(printf %.2f $(echo "$txMB * 8"| bc -l))

output "AS110 1: $result1 s, AS111 2: $result2 s"

result3=$(printf %.2f $(echo "800/$result1"| bc -l))
result4=$(printf %.2f $(echo "$txMbit/$result2"| bc -l))
ratio=$(printf %.2f $(echo "$result3/$result4"| bc -l))

output "Speed AS110 $result3 Mbit/s"
output "Speed AS111 $result4 Mbit/s"
output "Ratio $ratio"

if (( $(echo "$ratio < 2.5" |bc -l) )) && (( $(echo "$ratio > 1.5" |bc -l) )); then
tput setaf 2; output "Passed the test"; tput sgr0;
failed='false'
else
tput setaf 1; output "Failed the test. We wanted a ratio between 1.5 and 2.5 but we had $ratio"; tput sgr0;
failed='true'
fi

# Kill all started processes

# Remove tempfiles
rm -f .tempFile35234
rm -f .tempFile36234

kill -9 $pid1
kill -9 $pid2
kill -9 $pid3

killall netcat

./scion.sh stop

if $optRatio; then
    cd "$currDir"
    echo "$ratio" >> ratioRate.csv
fi

if $failed; then
    exit 1
fi
