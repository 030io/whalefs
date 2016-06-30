# whalefs
目标: PB级分布式文件存储,优化海量小文件存储.

##特性

-   高并发,低延迟
-   存储节点没有单点错误(冗余)
-   维护简单,代码量少

##安装

```sh
$ go get github.com/030io/whalefs
$ go install github.com/030io/whalefs
```

##帮助

```sh
$ whalefs --help-long
```

##运行

####master

```sh
# master需要使用redis存储元数据
$ whalefs master --redisIP localhost --redisPort 6379 --redisPW password --redisN databaseNum
# 冗余: --replication abc
# a: 相同machine的备份数
# b: 相同datacenter但不同machine的备份数
# b: 不同datacenter且不同machine的备份数
```

####volume manager(存储节点)

```sh
$ whalefs volume --dir volume_dir
# 以下两个选项跟冗余有关
# --machine 默认为volume manager跟master通信的ip
# --dataCenter 默认为空
```

####benchmark

```sh
$ whalefs benchmark
```

##API

-	上传:  `curl -F file=@xxx.jpg http://localhost:8888/dir/dst.jpg`
-	删除: `curl -X DELETE http://localhost:8888/dir/dst.jpg`
-	获取: `wget http://localhost:8888/dir/dst.jpg`

##性能测试:

```
# mac i7 ssd 单个volume manager节点
upload 10000 1024byte file:

concurrent:             16
time taken:             6.81 seconds
completed:              10000
failed:                 0
transferred:            10240000 byte
request per second:     1467.62
transferred per second: 1502843.99 b/s


read 10000 1024byte file:

concurrent:             16
time taken:             0.62 seconds
completed:              10000
failed:                 0
transferred:            10240000 byte
request per second:     16185.35
transferred per second: 16573796.70 b/s

# vps 单核 机械硬盘 单个volume manager节点
upload 800000 1024byte file:

concurrent:             64
time taken:             2415.91 seconds
completed:              800000
failed:                 0
transferred:            819200000 byte
request per second:     331.14
transferred per second: 339085.81 b/s


read 800000 1024byte file:

concurrent:             64
time taken:             242.18 seconds
completed:              800000
failed:                 0
transferred:            819200000 byte
request per second:     3303.31
transferred per second: 3382590.28 b/s
```

##注意:

-   `--replication` 冗余需要在搭建时就确定好,在运行过程中更改master的冗余选项,volume有可能变成只读,需要手动平衡(复制/删除)volume
-   volume manager节点的 `machine` `dataCenter` 同上
-   数据无价,且行且珍惜
-   如果觉得feature不够,不合你胃口,直接fork,就是这么任性
