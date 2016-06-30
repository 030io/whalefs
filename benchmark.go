package main

import (
	"time"
	"sync"
	"github.com/030io/whalefs/master/api"
	"io/ioutil"
	"os"
	"fmt"
	"strconv"
)

type result struct {
	concurrent  int
	num         int
	startTime   time.Time
	endTime     time.Time
	completed   int
	failed      int
	transferred uint64
}

func benchmark_() {
	uploadResult := &result{
		concurrent: *bmConcurrent,
		num: *bmNum,
		startTime: time.Now(),
	}
	loop := make(chan int)
	wg := sync.WaitGroup{}
	testFile, _ := ioutil.TempFile(os.TempDir(), "")
	testFile.Truncate(int64(*bmSize))
	testFile.Close()
	defer os.Remove(testFile.Name())

	for i := 0; i < *bmConcurrent; i++ {
		wg.Add(1)
		go func() {
			for b := range loop {
				err := api.Upload(*bmMasterHost, *bmMasterPort, testFile.Name() + strconv.Itoa(b), testFile.Name())
				if err == nil {
					uploadResult.completed += 1
				}else {
					uploadResult.failed += 1
					fmt.Println(err.Error())
				}
			}
			wg.Done()
		}()
	}

	for i := 0; i < *bmNum; i++ {
		loop <- i
	}
	close(loop)

	wg.Wait()
	uploadResult.endTime = time.Now()
	timeTaken := float64(uploadResult.endTime.UnixNano() - uploadResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("upload %d %dbyte file:\n\n", uploadResult.num, *bmSize)
	fmt.Printf("concurrent:             %d\n", uploadResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", uploadResult.completed)
	fmt.Printf("failed:                 %d\n", uploadResult.failed)
	fmt.Printf("transferred:            %d byte\n", uploadResult.completed * *bmSize)
	fmt.Printf("request per second:     %.2f\n", float64(uploadResult.num) / timeTaken)
	fmt.Printf("transferred per second: %.2f b/s\n", float64(uploadResult.completed) * float64(*bmSize) / timeTaken)

	readResult := &result{
		concurrent: *bmConcurrent,
		num: *bmNum,
		startTime: time.Now(),
	}
	loop = make(chan int)

	for i := 0; i < *bmConcurrent; i++ {
		wg.Add(1)
		go func() {
			for b := range loop {
				data, err := api.Get(*bmMasterHost, *bmMasterPort, testFile.Name() + strconv.Itoa(b))
				if err == nil &&len(data) == *bmSize {
					readResult.completed += 1
				}else {
					readResult.failed += 1
				}
			}
			wg.Done()
		}()
	}

	for i := 0; i < *bmNum; i++ {
		loop <- i
	}
	close(loop)
	wg.Wait()

	readResult.endTime = time.Now()
	timeTaken = float64(readResult.endTime.UnixNano() - readResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("\n\nread %d %dbyte file:\n\n", readResult.num, *bmSize)
	fmt.Printf("concurrent:             %d\n", readResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", readResult.completed)
	fmt.Printf("failed:                 %d\n", readResult.failed)
	fmt.Printf("transferred:            %d byte\n", readResult.completed * *bmSize)
	fmt.Printf("request per second:     %.2f\n", float64(readResult.num) / timeTaken)
	fmt.Printf("transferred per second: %.2f b/s\n", float64(readResult.completed) * float64(*bmSize) / timeTaken)
}
