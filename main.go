package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	masterServer "github.com/030io/whalefs/master"
	"github.com/030io/whalefs/volume/manager"
	"fmt"
	"net"
)

const version = "1.1 beta"

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
)

func main() {
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

		vm.Start()
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
