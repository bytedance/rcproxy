package main

import "github.com/garyburd/redigo/redis"
import "log"
import "fmt"
import "math/rand"
import "runtime"

var maxKeyNum = 10000000
var numWriter = 100
var numReader = 500
var server = "127.0.0.1:8088"

func write(server string, exitChan chan struct{}) {
	var conn redis.Conn
	var err error
	defer func() {
		if conn != nil {
			conn.Close()
		}
		if err != nil {
			log.Print(err)
		}
		exitChan <- struct{}{}
	}()

	conn, err = redis.Dial("tcp", server)
	if err != nil {
		return
	}
	for {
		i := rand.Int() % maxKeyNum
		key := fmt.Sprintf("key:%d", i)
		if _, err := conn.Do("SET", key, key); err != nil {
			log.Printf("set error %v", err)
			break
		}
	}
}

func read(server string, exitChan chan struct{}) {
	var conn redis.Conn
	var err error
	defer func() {
		if conn != nil {
			conn.Close()
		}
		if err != nil {
			log.Print(err)
		}
		exitChan <- struct{}{}
	}()

	conn, err = redis.Dial("tcp", server)
	if err != nil {
		return
	}
	for {
		i := rand.Int() % maxKeyNum
		key := fmt.Sprintf("key:%d", i)
		if reply, err := conn.Do("GET", key); err != nil {
			log.Printf("get error %v %v", key, err)
			break
		} else {
			log.Printf("key->%s", reply)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(8)
	exitChan := make(chan struct{})
	for i := 0; i < numWriter; i++ {
		go write(server, exitChan)
	}
	for i := 0; i < numReader; i++ {
		go read(server, exitChan)
	}
	for i := 0; i < numReader+numWriter; i++ {
		<-exitChan
	}
}
