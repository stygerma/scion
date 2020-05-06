tput setaf 2; echo "Starting benchmarks"; tput sgr0

tput setaf 2; echo "Benchmarking Classifier Part 5"; tput sgr0
for i in {1..30}; do
go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification5 >> testdata/BenchmarkClassification5.txt
echo "Run $i/30"
done

tput setaf 2; echo "Benchmarking Classifier Part 6"; tput sgr0
for i in {1..30}; do
go test -run=xxx github.com/scionproto/scion/go/border/qos/queues -bench=BenchmarkClassification6 >> testdata/BenchmarkClassification6.txt
echo "Run $i/30"
done
