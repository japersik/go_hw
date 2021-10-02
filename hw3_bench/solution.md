1. Поиск не оптимизированных мест:

```
 go test -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1  

goos: linux
goarch: amd64
pkg: hw_coursera/hw3_bench
cpu: AMD Ryzen 7 4800H with Radeon Graphics         
BenchmarkSlow-16               2         716976461 ns/op        19626600 B/op     189830 allocs/op
BenchmarkFast-16               2         658613765 ns/op        19611980 B/op     189820 allocs/op
PASS
ok      hw_coursera/hw3_bench   5.455s

```
```
go tool pprof main_test.go cpu.out
list SlowSearch 

go tool pprof main_test.go mem.out
list SlowSearch 
```

Наиболее затратные операции: 

1. CPU: 
```
         .      1.69s     38:           err := json.Unmarshal([]byte(line), &user)
         .      1.85s     62:                   if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
         .      1.54s     84:                   if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
```
2. MEM
```
         .    24.50MB     22:   fileContents, err := ioutil.ReadAll(file)
    5.47MB     5.59MB     32:   lines := strings.Split(string(fileContents), "\n")
    4.53MB    16.83MB     38:           err := json.Unmarshal([]byte(line), &user)
         .    62.41MB     62:                   if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
         .    41.24MB     84:                   if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {

```