#!/bin/bash

# doStuff() {
#     ls
#     sleep 3
# }

printBlue() {
    tput setaf 4; echo "$1"; tput sgr0;
}

startNetcatListener() {
    SCION_DAEMON_ADDRESS='127.0.0.19:30255'
    export SCION_DAEMON_ADDRESS
    tail -f /dev/null | ./../scion-apps/netcat/netcat -l $1 > ../scion-apps/netcat/data/server1.output
}


# exec 3>&1 4>&2
# time=$(TIMEFORMAT="%R"; { time doStuff 1>&3 2>&4; } 2>&1)
# exec 3>&- 4>&-

# echo "-----------------------"
# echo $time

tput setaf 4; echo "Starting the demo"; tput sgr0;

# # Generate topology and copy configuration files
# echo "Generate topology and copy configuration files"

# ./scion.sh topology -c topology/DemoTiny.topo

# cp go/border/qos/testdata/DemoConfig.yaml gen/ISD1/ASff00_0_110/br1-ff00_0_110-1/qosConfig.yaml
# cp go/border/qos/testdata/DemoConfigEmpty.yaml gen/ISD1/ASff00_0_111/br1-ff00_0_111-1/qosConfig.yaml
# cp go/border/qos/testdata/DemoConfigEmpty.yaml gen/ISD1/ASff00_0_112/br1-ff00_0_112-1/qosConfig.yaml

# # # Start SCION
# printBlue "Start SCION"

# ./scion.sh start nobuild
# ./scion.sh status
# sleep 5

# # # Do PING for 5 seconds AS110 to AS111
# printBlue "AS110 to AS111"
# ./bin/scmp echo -local 1-ff00:0:110,[127.0.0.1] -remote 1-ff00:0:111,[0.0.0.0] -sciond 127.0.0.11:30255 -c 5
# # # Do PING for 5 seconds AS110 to AS112
# printBlue "AS110 to AS112"
# ./bin/scmp echo -local 1-ff00:0:110,[127.0.0.1] -remote 1-ff00:0:112,[0.0.0.0] -sciond 127.0.0.11:30255 -c 5
# # # Do PING for 5 seconds AS111 to AS112
# printBlue "AS111 to AS112"
# ./bin/scmp echo -local 1-ff00:0:111,[127.0.0.1] -remote 1-ff00:0:112,[0.0.0.0] -sciond 127.0.0.19:30255 -c 5

# printBlue "Press enter to continue" 
# read -p ""

# # Start netcat server 1 in AS111
SCION_DAEMON_ADDRESS='127.0.0.19:30255'
export SCION_DAEMON_ADDRESS
tail -f /dev/null | ./../scion-apps/netcat/netcat -l 34234 > ../scion-apps/netcat/data/server1.output &
pid1=$!
# # Start netcat server 2 in AS111
SCION_DAEMON_ADDRESS='127.0.0.19:30255'
export SCION_DAEMON_ADDRESS
tail -f /dev/null | ./../scion-apps/netcat/netcat -l 35234 > ../scion-apps/netcat/data/server1.output &
pid2=$!
# # Start netcat server 3 in AS111
SCION_DAEMON_ADDRESS='127.0.0.19:30255'
export SCION_DAEMON_ADDRESS
tail -f /dev/null | ./../scion-apps/netcat/netcat -l 36234 > ../scion-apps/netcat/data/server1.output &
pid3=$!

printBlue "Started netcat servers"

printBlue "Start transfer file"
# # Transfer File from AS110 to AS111 to show that 10 Mbit/s can be reached
SCION_DAEMON_ADDRESS='127.0.0.11:30255'
export SCION_DAEMON_ADDRESS
# pv ../scion-apps/netcat/data/test100Mb.db | ./../scion-apps/netcat/netcat -vv 1-ff00:0:111,[127.0.0.1]:34234

# printBlue "Press enter to continue" 
# read -p ""

# # Transfer File from AS110 to AS111 to show that the transfer is two times faster
SCION_DAEMON_ADDRESS='127.0.0.11:30255'
export SCION_DAEMON_ADDRESS
time ./../scion-apps/netcat/netcat 1-ff00:0:111,[127.0.0.1]:35234 < ../scion-apps/netcat/data/test100Mb.db
&
# # Transfer File from AS112 to AS111 to show that it does not starve
SCION_DAEMON_ADDRESS='127.0.0.11:30255'
export SCION_DAEMON_ADDRESS
time ./../scion-apps/netcat/netcat 1-ff00:0:111,[127.0.0.1]:36234 < ../scion-apps/netcat/data/test100Mb.db

# # Measure time of execution and divide by file size to get the transfer speed
# # Show the transfer speeds, the ratio of the transfer speeds and the ratio attempted

# # Kill all started processes

kill -9 "$pid1"
kill -9 "$pid2"
kill -9 "$pid3"

# ./scion.sh stop
