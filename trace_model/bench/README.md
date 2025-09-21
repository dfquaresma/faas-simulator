# faas-simulator Benchmark 

Sample execution commend:

```bash
go test -bench=BenchmarkSimulator -benchmem -memprofile memprofile.out -cpuprofile cpuprofile.out
```

To evaluate use the following
```bash
go tool pprof profile.out
```

Inside it you can run top, list or web.


To run all tests, do:
```bash
go test -bench=. -benchmem -memprofile memprofile_all.out -cpuprofile cpuprofile_all.out
```
