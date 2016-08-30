package benchmark

import (
	"time"
	"sync"
	"github.com/030io/whalefs/master/api"
	"io/ioutil"
	"os"
	"fmt"
	"strconv"
	"crypto/rand"
	"io"
	"bytes"
	"crypto/md5"
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

func Benchmark(masterHost string, masterPort int, concurrent int, num int, size int) {
	uploadResult := &result{
		concurrent: concurrent,
		num: num,
		startTime: time.Now(),
	}
	loop := make(chan int)
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}

	randBytes := make([]byte, size)
	rand.Read(randBytes)

	dataMd5 := md5.Sum(randBytes)

	testFile, _ := ioutil.TempFile(os.TempDir(), "")
	testFile.Truncate(int64(size))
	io.Copy(testFile, bytes.NewReader(randBytes))
	testFile.Close()
	defer os.Remove(testFile.Name())

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			for b := range loop {
				err := api.Upload(masterHost, masterPort, testFile.Name() + strconv.Itoa(b), testFile.Name())
				mutex.Lock()
				if err == nil {
					uploadResult.completed += 1
				} else {
					uploadResult.failed += 1
					fmt.Println("write failed:", err.Error())
				}
				mutex.Unlock()
			}
			wg.Done()
		}()
	}

	for i := 0; i < num; i++ {
		loop <- i
	}
	close(loop)

	wg.Wait()
	uploadResult.endTime = time.Now()
	timeTaken := float64(uploadResult.endTime.UnixNano() - uploadResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("upload %d %dbyte file:\n\n", uploadResult.num, size)
	fmt.Printf("concurrent:             %d\n", uploadResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", uploadResult.completed)
	fmt.Printf("failed:                 %d\n", uploadResult.failed)
	fmt.Printf("transferred:            %d byte\n", uploadResult.completed * size)
	fmt.Printf("request per second:     %.2f\n", float64(uploadResult.num) / timeTaken)
	fmt.Printf("transferred per second: %.2f byte\n", float64(uploadResult.completed) * float64(size) / timeTaken)

	readResult := &result{
		concurrent: concurrent,
		num: num,
		startTime: time.Now(),
	}
	loop = make(chan int)

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			for b := range loop {
				data, err := api.Get(masterHost, masterPort, testFile.Name() + strconv.Itoa(b))
				mutex.Lock()
				if err == nil &&md5.Sum(data) == dataMd5 {
					readResult.completed += 1
				} else {
					readResult.failed += 1
					fmt.Println("read failed:", err.Error())
				}
				mutex.Unlock()
			}
			wg.Done()
		}()
	}

	for i := 0; i < num; i++ {
		loop <- i
	}
	close(loop)
	wg.Wait()

	readResult.endTime = time.Now()
	timeTaken = float64(readResult.endTime.UnixNano() - readResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("\n\nread %d %dbyte file:\n\n", readResult.num, size)
	fmt.Printf("concurrent:             %d\n", readResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", readResult.completed)
	fmt.Printf("failed:                 %d\n", readResult.failed)
	fmt.Printf("transferred:            %d byte\n", readResult.completed * size)
	fmt.Printf("request per second:     %.2f\n", float64(readResult.num) / timeTaken)
	fmt.Printf("transferred per second: %.2f byte\n", float64(readResult.completed) * float64(size) / timeTaken)

	deleteResult := &result{
		concurrent: concurrent,
		num: num,
		startTime: time.Now(),
	}
	loop = make(chan int)

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			for b := range loop {
				err := api.Delete(masterHost, masterPort, testFile.Name() + strconv.Itoa(b))
				mutex.Lock()
				if err == nil {
					deleteResult.completed += 1
				} else {
					deleteResult.failed += 1
					fmt.Println("delete failed:", err.Error())
				}
				mutex.Unlock()
			}
			wg.Done()
		}()
	}

	for i := 0; i < num; i++ {
		loop <- i
	}
	close(loop)
	wg.Wait()

	deleteResult.endTime = time.Now()
	timeTaken = float64(deleteResult.endTime.UnixNano() - deleteResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("\n\ndelete %d %dbyte file:\n\n", deleteResult.num, size)
	fmt.Printf("concurrent:             %d\n", deleteResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", deleteResult.completed)
	fmt.Printf("failed:                 %d\n", deleteResult.failed)
	fmt.Printf("transferred:            %d byte\n", deleteResult.completed * size)
	fmt.Printf("request per second:     %.2f\n", float64(deleteResult.num) / timeTaken)
	fmt.Printf("transferred per second: %.2f byte\n", float64(deleteResult.completed) * float64(size) / timeTaken)
}
