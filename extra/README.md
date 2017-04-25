```
$ go test -bench Read -count 3

BenchmarkRead/mode=2,m=4-8         	       5	 311218250 ns/op
BenchmarkRead/mode=2,m=4-8         	       5	 317086338 ns/op
BenchmarkRead/mode=2,m=4-8         	       5	 308295082 ns/op
BenchmarkRead/mode=2,m=16-8        	       5	 223310483 ns/op
BenchmarkRead/mode=2,m=16-8        	       5	 231452472 ns/op
BenchmarkRead/mode=2,m=16-8        	       5	 242802455 ns/op
BenchmarkRead/mode=2,m=64-8        	      10	 210723926 ns/op
BenchmarkRead/mode=2,m=64-8        	       5	 777420932 ns/op
BenchmarkRead/mode=2,m=64-8        	       5	 200855156 ns/op
BenchmarkRead/mode=3,m=4-8         	       5	 906213188 ns/op
BenchmarkRead/mode=3,m=4-8         	       5	1105820741 ns/op
BenchmarkRead/mode=3,m=4-8         	       2	 527907341 ns/op
BenchmarkRead/mode=3,m=16-8        	       3	 393121910 ns/op
BenchmarkRead/mode=3,m=16-8        	       3	 378802854 ns/op
BenchmarkRead/mode=3,m=16-8        	       3	 398507217 ns/op
BenchmarkRead/mode=3,m=64-8        	       3	 379417432 ns/op
BenchmarkRead/mode=3,m=64-8        	       3	 373250323 ns/op
BenchmarkRead/mode=3,m=64-8        	       3	 392327897 ns/op

```

Loading into RAM has no visible advantage over mmap. Variance seems high.