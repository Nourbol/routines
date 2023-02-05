package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	data := (<-in).(string)

	firstPartChannel := make(chan string)
	wg.Add(1)
	go func() {
		defer wg.Done()
		firstPartChannel <- DataSignerCrc32(data)
	}()

	secondPartChannel := make(chan string)
	wg.Add(1)
	go func() {
		defer wg.Done()
		secondPartChannel <- DataSignerCrc32(DataSignerMd5(data))
	}()

	wg.Wait()
	out <- <-firstPartChannel + "~" + <-secondPartChannel
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	data := (<-in).(string)

	hashedValues := make([]string, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			hashedValues[i] = DataSignerCrc32(strconv.Itoa(i) + data)
		}(i)
	}
	wg.Wait()
	out <- strings.Join(hashedValues, "")
}

func CombineResults(in, out chan interface{}) {
	select {
	case <-in:
		return
	default:
		results := make([]string, 0)

		for result := range in {
			results = append(results, result.(string))
		}

		sort.Strings(results)
		out <- strings.Join(results, "_")
	}
}

func ExecutePipeline(jobs ...job) {
	channels := make([]chan interface{}, len(jobs))
	for i := range channels {
		channels[i] = make(chan interface{})
		defer close(channels[i])
	}

	for i, pipe := range jobs {
		go func(i int, pipe job) {
			if i == len(channels)-1 {
				pipe(channels[i], make(chan interface{}))
				return
			}
			pipe(channels[i], channels[i+1])
		}(i, pipe)
	}

	channels[0] <- 1
}
