package proxy

import (
	"github.com/collinmsn/resp"
)

const (
	CMD_FLAG_READONLY = 1 << iota
	CMD_FLAG_BLACK
)

/**
"b" stands for blacklist command
"r" stands for readonly command
*/
var cmdTable = [][2]string{
	{"ASKING", "b"},
	{"AUTH", "b"},
	{"BGREWRITEAOF", "b"},
	{"BGSAVE", "b"},
	{"BITCOUNT", "r"},
	{"BITOP", "b"},
	{"BITPOS", "r"},
	{"BLPOP", "b"},
	{"BRPOP", "b"},
	{"BRPOPLPUSH", "b"},
	{"CLIENT", "b"},
	{"CLUSTER", "b"},
	{"COMMAND", "r"},
	{"CONFIG", "b"},
	{"DBSIZE", "b"},
	{"DEBUG", "b"},
	{"DISCARD", "b"},
	{"DUMP", "r"},
	{"ECHO", "b"},
	{"EXEC", "b"},
	{"EXISTS", "r"},
	{"FLUSHALL", "b"},
	{"FLUSHDB", "b"},
	{"GET", "r"},
	{"GETBIT", "r"},
	{"GETRANGE", "r"},
	{"HEXISTS", "r"},
	{"HGET", "r"},
	{"HGETALL", "r"},
	{"HKEYS", "r"},
	{"HLEN", "r"},
	{"HMGET", "r"},
	{"HSCAN", "r"},
	{"HVALS", "r"},
	{"INFO", "b"},
	{"KEYS", "b"},
	{"LASTSAVE", "b"},
	{"LATENCY", "r"},
	{"LINDEX", "r"},
	{"LLEN", "r"},
	{"LRANGE", "r"},
	{"MGET", "r"},
	{"MIGRATE", "b"},
	{"MONITOR", "b"},
	{"MOVE", "b"},
	{"MSETNX", "b"},
	{"MULTI", "b"},
	{"OBJECT", "b"},
	{"PFCOUNT", "r"},
	{"PFSELFTEST", "r"},
	{"PING", "b"},
	{"PSUBSCRIBE", "b"},
	{"PSYNC", "r"},
	{"PTTL", "r"},
	{"PUBLISH", "b"},
	{"PUBSUB", "r"},
	{"PUNSUBSCRIBE", "b"},
	{"RANDOMKEY", "b"},
	{"READONLY", "r"},
	{"READWRITE", "r"},
	{"RENAME", "b"},
	{"RENAMENX", "b"},
	{"REPLCONF", "r"},
	{"SAVE", "b"},
	{"SCAN", "b"},
	{"SCARD", "r"},
	{"SCRIPT", "b"},
	{"SDIFF", "r"},
	{"SELECT", "b"},
	{"SHUTDOWN", "b"},
	{"SINTER", "r"},
	{"SISMEMBER", "r"},
	{"SLAVEOF", "b"},
	{"SLOWLOG", "b"},
	{"SMEMBERS", "r"},
	{"SRANDMEMBER", "r"},
	{"SSCAN", "r"},
	{"STRLEN", "r"},
	{"SUBSCRIBE", "b"},
	{"SUBSTR", "r"},
	{"SUNION", "r"},
	{"SYNC", "b"},
	{"TIME", "b"},
	{"TTL", "r"},
	{"TYPE", "r"},
	{"UNSUBSCRIBE", "b"},
	{"UNWATCH", "b"},
	{"WAIT", "r"},
	{"WATCH", "b"},
	{"ZCARD", "r"},
	{"ZCOUNT", "r"},
	{"ZLEXCOUNT", "r"},
	{"ZRANGE", "r"},
	{"ZRANGEBYLEX", "r"},
	{"ZRANGEBYSCORE", "r"},
	{"ZRANK", "r"},
	{"ZREVRANGE", "r"},
	{"ZREVRANGEBYLEX", "r"},
	{"ZREVRANGEBYSCORE", "r"},
	{"ZREVRANK", "r"},
	{"ZSCAN", "r"},
	{"ZSCORE", "r"},
}

var cmdFlagMap = make(map[string]int)

func init() {
	for _, row := range cmdTable {
		if row[1] == "b" {
			cmdFlagMap[row[0]] = CMD_FLAG_BLACK
		} else if row[1] == "r" {
			cmdFlagMap[row[0]] = CMD_FLAG_READONLY
		}
	}
}

func CmdFlag(cmd *resp.Command) int {
	return cmdFlagMap[cmd.Name()]
}
