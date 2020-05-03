tput setaf 2; echo "Starting benchmarks"; tput sgr0

l=3
for i in {1..$l}; do
go test -run=xxx queues/ -bench=BenchmarkQueuesPopSingle
echo "Run $i/$l"
done
