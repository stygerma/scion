rm logs/*

./scion.sh start

date
echo "Sleep 1"
sleep 1

SCION_DAEMON_ADDRESS='127.0.0.19:30255' && export SCION_DAEMON_ADDRESS && ./../scion-apps/bwtester/bwtestserver/bwtestserver -p 40001 &

SCION_DAEMON_ADDRESS='127.0.0.11:30255' && export SCION_DAEMON_ADDRESS && ./../scion-apps/bwtester/bwtestclient/bwtestclient -s 1-ff00:0:111,[127.0.0.1]:40001 -cs 5,$1,?,$2Mbps -sc 1,1000,?,1Mbps

killall bwtestserver

./scion.sh stop

echo "We are done!"