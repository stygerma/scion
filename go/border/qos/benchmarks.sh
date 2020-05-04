tput setaf 2; echo "Starting benchmarks"; tput sgr0

# tput setaf 2; echo "Benchmarking Policer Available"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkAvailable >> testdata/BenchmarkAvailable.txt
# echo "Run $i/10"
# done

# tput setaf 2; echo "Benchmarking Policer Take"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkTake >> testdata/BenchmarkTake.txt
# echo "Run $i/10"
# done

# tput setaf 2; echo "Benchmarking Policer Refill"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkRefill >> testdata/BenchmarkRefill.txt
# echo "Run $i/10"
# done

tput setaf 2; echo "Benchmarking qos PoliceQueue"; tput sgr0
for i in {1..10}; do
go test -run=xxx github.com/scionproto/scion/go/border/qos -bench=BenchmarkPoliceQueue >> testdata/BenchmarkPoliceQueue.txt
echo "Run $i/10"
done

tput setaf 2; echo "Benchmarking qos CheckAction"; tput sgr0
for i in {1..10}; do
go test -run=xxx github.com/scionproto/scion/go/border/qos -bench=BenchmarkCheckAction >> testdata/BenchmarkCheckAction.txt
echo "Run $i/10"
done

tput setaf 2; echo "Benchmarking qos Queue Single Packet Blocking"; tput sgr0
for i in {1..10}; do
go test -run=xxx github.com/scionproto/scion/go/border/qos -bench=BenchmarkQueueSinglePacketBlocking >> testdata/BenchmarkQueueSinglePacketBlocking.txt
echo "Run $i/10"
done

tput setaf 2; echo "Benchmarking qos Queue Single Packet Non-Blocking"; tput sgr0
for i in {1..10}; do
go test -run=xxx github.com/scionproto/scion/go/border/qos -bench=BenchmarkQueueSinglePacket >> testdata/BenchmarkQueueSinglePacket.txt
echo "Run $i/10"
done

# tput setaf 2; echo "Benchmarking Classifier Part 1"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification1 >> testdata/BenchmarkClassification1.txt
# echo "Run $i/10"
# done

# tput setaf 2; echo "Benchmarking Classifier Part 2"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification2 >> testdata/BenchmarkClassification2.txt
# echo "Run $i/10"
# done

# tput setaf 2; echo "Benchmarking Classifier Part 3"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification3 >> testdata/BenchmarkClassification3.txt
# echo "Run $i/10"
# done

# tput setaf 2; echo "Benchmarking Classifier Part 4"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification4 >> testdata/BenchmarkClassification4.txt
# echo "Run $i/10"
# done

# tput setaf 2; echo "Benchmarking Classifier Part 5"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification5 >> testdata/BenchmarkClassification5.txt
# echo "Run $i/10"
# done

# tput setaf 2; echo "Benchmarking Classifier Part 6"; tput sgr0
# for i in {1..10}; do
# go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification6 >> testdata/BenchmarkClassification6.txt
# echo "Run $i/10"
# done