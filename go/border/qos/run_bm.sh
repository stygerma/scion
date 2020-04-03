echo "1st Argument = $1, comparing against baseline of file $1"
echo "2nd Argument = $2, writing data into file $2"

rm testdata/benchmarks/BenchmarkEnqueueForProfile$2

for i in {1..30}; do go test -run=xxx -bench=BenchmarkEnqueueForProfile -cpuprofile=testdata/profiles/BenchmarkEnqueueForProfile$2-$i.pprof >> testdata/benchmarks/BenchmarkEnqueueForProfile$2; echo "Run $i/30"; done

benchstat testdata/benchmarks/BenchmarkEnqueueForProfile$1 testdata/benchmarks/BenchmarkEnqueueForProfile$2