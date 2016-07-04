package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	masterServer "github.com/030io/whalefs/master"
	"github.com/030io/whalefs/volume/manager"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
)

const version = "2.2 beta"

var (
	app = kingpin.New("whalefs", "A simple filesystem for small file.  Version: " + version)

	verbose = app.Flag("verbose", "verbose level").Short('v').Default("0").Int()

	master = app.Command("master", "master server")
	masterPort = master.Flag("port", "master port").Short('p').Default("8888").Int()
	masterReplication = master.Flag("replication", "replication setting").Short('r').Default("000").String()
	masterRedisServer = master.Flag("redisIP", "ip of redis server").Default("localhost").String()
	masterRedisPort = master.Flag("redisPort", "ip of redis server").Default("6379").Int()
	masterRedisPW = master.Flag("redisPW", "password of redis server").String()
	masterRedisN = master.Flag("redisN", "database of redis server").Default("0").Int()

	volumeManager = app.Command("volume", "volume manager server")
	vmDir = volumeManager.Flag("dir", "directory to store data").Required().String()
	vmAdminHost = volumeManager.Flag("adminHost", "for manage files (default: auto detect by master)").String()
	vmAdminPort = volumeManager.Flag("adminPort", "for manage files (default: 7800-7899)").Int()
	vmPublicHost = volumeManager.Flag("publicHost", "for access files (default: auto detect by master)").String()
	vmPublicPort = volumeManager.Flag("publicPort", "for access files (default: 7900-7999)").Int()
	vmMasterHost = volumeManager.Flag("masterHost", "host of master server").Default("localhost").String()
	vmMasterPort = volumeManager.Flag("masterPort", "port of master server").Default("8888").Int()
	vmMachine = volumeManager.Flag("machine", "machine tag of volume manager server (defalut: auto detect by master)").String()
	vmDataCenter = volumeManager.Flag("dataCenter", "datacenter tag of volume manager server (defalut: \"\")").String()

	benchmark = app.Command("benchmark", "benchmark")
	bmMasterHost = benchmark.Flag("masterHost", "host of master server").Default("localhost").String()
	bmMasterPort = benchmark.Flag("masterPort", "post of master server").Default("8888").Int()
	bmConcurrent = benchmark.Flag("concurrent", "concurrent").Default("16").Int()
	bmNum = benchmark.Flag("num", "number of file write/read").Default("1000").Int()
	bmSize = benchmark.Flag("size", "size of file write/read").Default("1024").Int()
)

func main() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1024

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	switch command {
	case master.FullCommand():
		m, err := masterServer.NewMaster()
		if err != nil {
			panic(m)
		}
		m.Metadata, err = masterServer.NewMetadataRedis(*masterRedisServer, *masterRedisPort, *masterRedisPW, *masterRedisN)
		if err != nil {
			panic(err)
		}
		m.Port = *masterPort
		for i, c := range *masterReplication {
			m.Replication[i] = int(c) - int('0')
		}

		go func() {
			sig := <-signals
			m.Stop()
			switch sig {
			case syscall.SIGINT:
				os.Exit(int(syscall.SIGINT))
			case syscall.SIGTERM:
				os.Exit(int(syscall.SIGTERM))
			}
		}()

		m.Start()
	case volumeManager.FullCommand():
		vm, err := manager.NewVolumeManager(*vmDir)
		if err != nil {
			panic(err)
		}

		vm.AdminHost = *vmAdminHost
		if *vmAdminPort == 0 {
			*vmAdminPort, err = getFreeTcpPort(7800, 7900)
			if err != nil {
				panic(err)
			}
		}
		vm.AdminPort = *vmAdminPort

		vm.PublicHost = *vmPublicHost
		if *vmPublicPort == 0 {
			*vmPublicPort, err = getFreeTcpPort(7900, 8000)
			if err != nil {
				panic(err)
			}
		}
		vm.PublicPort = *vmPublicPort

		vm.MasterHost = *vmMasterHost
		vm.MasterPort = *vmMasterPort
		vm.Machine = *vmMachine
		vm.DataCenter = *vmDataCenter

		go func() {
			sig := <-signals
			vm.Stop()
			switch sig {
			case syscall.SIGINT:
				os.Exit(int(syscall.SIGINT))
			case syscall.SIGTERM:
				os.Exit(int(syscall.SIGTERM))
			}
		}()

		vm.Start()
	case benchmark.FullCommand():
		go func() {
			sig := <-signals
			switch sig {
			case syscall.SIGINT:
				os.Exit(int(syscall.SIGINT))
			case syscall.SIGTERM:
				os.Exit(int(syscall.SIGTERM))
			}
		}()

		benchmark_()
	}
}

func getFreeTcpPort(start, end int) (int, error) {
	for i := start; i < end; i++ {
		if ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", i)); err == nil {
			ln.Close()
			return ln.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return 0, fmt.Errorf("can't get a free port between [%d, %d)", start, end)
}
