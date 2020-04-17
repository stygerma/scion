tc qdisc del dev lo root
tc qdisc add dev lo root handle 1: htb
tc class add dev lo parent 1: classid 1:1 htb rate 500mbit
tc filter add dev lo parent 1: protocol ip prio 1 u32 flowid 1:1 match ip dst $1
tc qdisc add dev lo parent 1:1 handle 10: tbf rate 5mbit burst 10kbit latency 5ms
