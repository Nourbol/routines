package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for input := range in {

		wg.Add(1)

		go func(input interface{}) {
			defer wg.Done()

			data := strconv.Itoa(input.(int))
			crc32chan := make(chan string)
			crc32md5chan := make(chan string)

			go func(crc32chan chan string, data string) {
				crc32chan <- DataSignerCrc32(data)
			}(crc32chan, data)

			mu.Lock()
			md5 := DataSignerMd5(data)
			mu.Unlock()

			go func(crc32md5chan chan string, md5 string) {
				crc32md5chan <- DataSignerCrc32(md5)
			}(crc32md5chan, md5)

			crc32 := <-crc32chan
			crc32md5 := <-crc32md5chan

			fmt.Printf("%s SingleHash data %s\n", data, data)
			fmt.Printf("%s SingleHash md5(data) %s\n", data, md5)
			fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32md5)
			fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32)
			fmt.Printf("%s SingleHash result %s\n", data, crc32+"~"+crc32md5)

			result := crc32 + "~" + crc32md5
			out <- result

		}(input)

	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var array [6]string

	for input := range in {
		wg.Add(1)

		go func(input interface{}) {
			defer wg.Done()
			var wgInner sync.WaitGroup
			data := input.(string)
			for th := 0; th < 6; th++ {
				wgInner.Add(1)
				go func(th int) {
					defer wgInner.Done()
					tmp := DataSignerCrc32(strconv.Itoa(th) + data)
					mu.Lock()
					array[th] = tmp
					mu.Unlock()
				}(th)
			}
			wgInner.Wait()
			result := strings.Join(array[:], "")
			out <- result
		}(input)
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var tmp []string
	wg := &sync.WaitGroup{}
	for input := range in {
		wg.Add(1)
		go func(input interface{}) {
			defer wg.Done()
			tmp = append(tmp, input.(string))
		}(input)
	}
	wg.Wait()
	sort.Strings(tmp)
	result := strings.Join(tmp, "_")
	fmt.Printf("CombineResults \n%s\n", result)
	out <- result
}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, job := range jobs {
		wg.Add(1)

		out := make(chan interface{})
		go worker(in, out, job, wg)
		in = out
	}
	wg.Wait()
}

func worker(in, out chan interface{}, job job, wg *sync.WaitGroup) {
	defer close(out)
	defer wg.Done()
	job(in, out)
}
