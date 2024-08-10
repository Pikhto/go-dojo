package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
func ExecutePipeline(works ...job) {
	out := make(chan interface{})
	wg := sync.WaitGroup{}
	wg.Add(len(works))

	for _, work := range works {
		out = Worker(&wg, out, work)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	var lock sync.Mutex

	for data := range in {
		wg.Add(1)
		go func(data int) {
			defer wg.Done()

			strNum := strconv.Itoa(data)
			crc32Value := make(chan string, 1)
			go func() {
				crc32Value <- DataSignerCrc32(strNum)
			}()

			lock.Lock()
			md5Value := DataSignerMd5(strNum)
			lock.Unlock()

			crc32OfMd5 := DataSignerCrc32(md5Value)
			out <- <-crc32Value + "~" + crc32OfMd5
		}(data.(int))
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for data := range in {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()

			dataStr := data.(string)
			result := make([]string, 6)
			var innerWg sync.WaitGroup
			for i := 0; i < 6; i++ {
				innerWg.Add(1)
				go func(i int) {
					defer innerWg.Done()
					result[i] = DataSignerCrc32(strconv.Itoa(i) + dataStr)
				}(i)
			}
			innerWg.Wait()
			out <- strings.Join(result, "")
		}(data)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	res := make([]string, 0)
	for s := range in {
		res = append(res, s.(string))
	}
	sort.Strings(res)
	out <- strings.Join(res, "_")
}

func Worker(wg *sync.WaitGroup, in chan interface{}, fn job) chan interface{} {
	out := make(chan interface{})
	go func() {
		defer wg.Done()
		defer close(out)
		fn(in, out)
	}()
	return out
}
