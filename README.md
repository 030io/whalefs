# whalefs
目标: PB级分布式文件存储,优化海量小文件存储.

##特性

-   高并发,低延迟
-   存储节点没有单点错误(冗余)
-   维护简单,代码量少
-   没有其他安装依赖

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

####volume manager(存储节点)

```sh
$ whalefs volume --dir volume_dir
```

####master

```sh
$ whalefs master
# --replication abc
# a: 相同machine的备份数
# b: 相同datacenter但不同machine的备份数
# b: 不同datacenter且不同machine的备份数
```

####benchmark

```sh
$ whalefs benchmark
```

##API

-	上传:  `curl -F file=@xxx.jpg http://localhost:8888/dir/dst.jpg`
-	删除: `curl -X DELETE http://localhost:8888/dir/dst.jpg`
-	获取: `wget http://localhost:8888/dir/dst.jpg`