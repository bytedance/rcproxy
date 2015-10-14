# rcproxy
[![Build Status](https://travis-ci.org/collinmsn/rcproxy.svg)](https://travis-ci.org/collinmsn/rcproxy)
## Introduction

  RCProxy, short for Redis Cluster Proxy, is a redis proxy built on redis 3.0(see [redis cluster spec](http://redis.io/topics/cluster-spec "悬停显示")). It aims to be simple, robust and high performance. It caches the cluster topology and implements the MOVED and ASK redirection. Cluster topology is updated on backend server failure or MOVED redirection received. Client is totally transpanrent about the backend data migration.
  
## Getting started
  * Make sure Redis 3.0.0 RC or above has been installed in your PATH. Run 'redis-server --version' to check it.
  * Run 'python start_cluster.py' in tests directory to start the redis cluster
  * Run 'sh bootstrap.sh' in rcproxy to build rcproxy. Note that [godep](https://github.com/tools/godep) is required to solved dependencies. Otherwise, you can use go get to fetch missing packages.
  * Run 'sh start_proxy.sh' to run rcproxy. 
  * Allmost done! Now you can play with it with redis-cli, or run tests/loadrunner.go and tests/random_migration.py to see how it handles slots migration properly, or run tests/bench/redis-benchmark.sh to benchmark it. 
 
## Architecture
![](https://github.com/collinmsn/rcproxy/blob/master/rcproxy.png)

  Each client connection is wrapped with a session, which spawns two goroutines to read request from and write response to the client. Each session appends it's request to dispatcher's request queue, then dispatcher route request to the right task runner according key hash and slot table. Task runner sends requests to its backend server and read responses from it.
  Upon cluster topology changed, backend server will response MOVED or ASK error. These error is handled by session, by sending request to destination server directly. Session will trigger dispatcher to update slot info on MOVED error. When connection error is returned by task runner, session will trigger dispather to reload topology.

## Performance
  Benchmark it with tests/bench/redis-benchmark.sh
  
## Status
  Product ready
  
## Feedback
  Feedbacks and bug reports are highly appreciated. Contact me by 4130944@qq.com
  

 

