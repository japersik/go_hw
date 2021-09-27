package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job){
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, j := range jobs {
		wg.Add(1)
		out:=make(chan interface{})
		go func(j job, in, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			j(in,out)

		}(j,in,out,wg)
		in = out
	}
	wg.Wait()
}
var DSMd5 = &sync.Mutex{}
func SingleHash(in,out chan interface{})  {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		go SingleHashFunc(i,out,wg)
	}
	wg.Wait()
}
func SingleHashFunc(in interface{},out chan interface{},wg *sync.WaitGroup)  {
	defer wg.Done()
	data := strconv.Itoa(in.(int))
	DSMd5.Lock()
	md5Data := DataSignerMd5(data)
	DSMd5.Unlock()

	ch := make(chan string)
	go func() {ch<-DataSignerCrc32(data)}()
	crc32Md5Data := DataSignerCrc32(md5Data)
	crc32Data := <-ch

	out <- crc32Data + "~" + crc32Md5Data
}

func MultiHash(in,out chan interface{})  {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		go multiHashFunc(i.(string), out, wg)
	}
	wg.Wait()
}
func multiHashFunc(in string,out chan interface{},wg *sync.WaitGroup)  {
	const th int = 6
	defer wg.Done()
	mu := &sync.Mutex{}
	mhWg := &sync.WaitGroup{}
	combinedChunks := make([]string, th)

	for i := 0; i < th; i++ {
		mhWg.Add(1)
		data := strconv.Itoa(i) + in

		go func(acc []string, index int, data string, jobWg *sync.WaitGroup, mu *sync.Mutex) {
			defer jobWg.Done()
			data = DataSignerCrc32(data)

			mu.Lock()
			acc[index] = data
			mu.Unlock()
		}(combinedChunks, i, data, mhWg, mu)
	}

	mhWg.Wait()
	out <- strings.Join(combinedChunks, "")
}

func CombineResults(in, out chan interface{}) {
	var result []string

	for i := range in {
		result = append(result, i.(string))
	}
	sort.Strings(result)
	out <- strings.Join(result, "_")
}

func main() {
	fmt.Println("Run test: 'go test . -v' ")
}

