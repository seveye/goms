# goms
* jkkkls@aliyun.com

#### 介绍
基于gorpc改写微服务框架

#### 软件架构

* rpc性能：qps: 20W+/秒，详细数据可以运行benchmark中代码测试
* watch: 实现基础的基于内存的kv数据库，支持服务的注册和发现
* 后台，基于beego+jquery实现。这部分代码没有上传，需要的话联系我
* 监控模块，后台实现监控服务，每个节点定时向后台推送自己数据，同后台通过supervisord实现节点启动，停止和重启功能
* api网关


#### rpc

* rpc参考examples/rpc代码

#### 微服务

* 微服务参考examples/microservice代码
* watch 服务注册发现
* db    简单内存数据库服务
* query 查询服务，定时查询和修改数据库

1. 编译运行watch
```shell
[guangbo@guangbo-pc goms]$ cd examples/microservice/watch/
[guangbo@guangbo-pc watch]$ go build
[guangbo@guangbo-pc watch]$ ./watch
2020/02/28 15:45:01 WatchServer start, host: :12345
```

2. 编译运行db
```shell
[guangbo@guangbo-pc goms]$ cd examples/microservice/db/
[guangbo@guangbo-pc db]$ go build
[guangbo@guangbo-pc db]$ ./db
2020-02-28 15:49:22.548158 INFO rpc_node.go:143 初始化服务节点 [id=0 name=db0 host=127.0.0.1 port=10001 region=0]
2020-02-28 15:49:22.548513 INFO rpc_node.go:183 初始化服务 [name=DB]
```

2. 编译运行query
```shell
[guangbo@guangbo-pc goms]$ cd examples/microservice/query/
[guangbo@guangbo-pc query]$ go build
[guangbo@guangbo-pc query]$ ./query
2020-02-28 15:50:05.759167 INFO rpc_node.go:143 初始化服务节点 [id=1 name=query0 host=127.0.0.1 port=10002 region=0]
2020-02-28 15:50:05.763381 INFO rpc_node.go:183 初始化服务 [name=Qeury]
2020-02-28 15:50:05.763897 INFO rpc_node.go:82 连接节点 [name=db0 address=127.0.0.1:10001 region=0 isClose=false]
2020-02-28 15:50:15.761677 INFO service.go:62 Query [value=]
2020-02-28 15:50:25.762323 INFO service.go:62 Query [value=1]
2020-02-28 15:50:35.762732 INFO service.go:62 Query [value=2]


```

#### api网关

* 微服务参考examples/gateway
* watch 服务注册发现
* db    简单内存数据库服务
* agte  网关服务

1. 编译运行watch
```shell
[guangbo@guangbo-pc goms]$ cd examples/gateway/watch/
[guangbo@guangbo-pc watch]$ go build
[guangbo@guangbo-pc watch]$ ./watch
2020/02/28 15:45:01 WatchServer start, host: :12345
```

2. 编译运行db
```shell
[guangbo@guangbo-pc goms]$ cd examples/gateway/db/
[guangbo@guangbo-pc db]$ go build
[guangbo@guangbo-pc db]$ ./db
2020-03-06 15:36:57.072443 INFO rpc_node.go:143 初始化服务节点 [id=0 name=db0 host=127.0.0.1 port=10001 region=0]
2020-03-06 15:36:57.073114 INFO rpc_node.go:183 初始化服务 [name=DB]
2020-03-06 15:37:15.489717 INFO rpc_node.go:82 连接节点 [name=gate0 address=127.0.0.1:10002 region=0 isClose=false]
2020-03-06 15:38:11.775770 INFO service.go:28 Update [req=key:"123" value:"aaa" ]
2020-03-06 15:38:25.881003 INFO service.go:34 Query [req=key:"123" ]
```

3. 编译运行gate
```shell
[guangbo@guangbo-pc goms]$ cd ./examples/gateway/gate/
[guangbo@guangbo-pc gate]$ go build
[guangbo@guangbo-pc gate]$ ./gate
2020-03-06 15:37:15.488691 INFO rpc_node.go:143 初始化服务节点 [id=1 name=gate0 host=127.0.0.1 port=10002 region=0]
Now listening on: http://0.0.0.0:8091
Application started. Press CTRL+C to shut down.
2020-03-06 15:37:15.489065 INFO rpc_node.go:183 初始化服务 [name=Gate]
2020-03-06 15:37:15.489334 INFO rpc_node.go:82 连接节点 [name=db0 address=127.0.0.1:10001 region=0 isClose=false]
2020/03/06 15:38:11 -------------handler----------------- /api/DB/Update
2020/03/06 15:38:25 -------------handler----------------- /api/DB/Query
```

4. 调用api，先设置123=aaa,再查询123
```shell
[guangbo@guangbo-pc gate]$ curl -i -X POST \
>    -d \
> '{"key":"123","value":"aaa"}' \
>  'http://127.0.0.1:8091/api/DB/Update'
HTTP/1.1 200 OK
Vary: Origin
Date: Fri, 06 Mar 2020 07:38:11 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

{}
[guangbo@guangbo-pc gate]$  curl -i -X POST \
>    -d \
> '{"key":"123"}' \
>  'http://127.0.0.1:8091/api/DB/Query'
HTTP/1.1 200 OK
Vary: Origin
Date: Fri, 06 Mar 2020 07:38:25 GMT
Content-Length: 15
Content-Type: text/plain; charset=utf-8

{"value":"aaa"}
```