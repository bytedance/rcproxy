package proxy

import (
	"bytes"
	"fmt"

	"github.com/collinmsn/resp"
)

const (
	NumSlots                   = 16384
	CLUSTER_SLOTS_START        = 0
	CLUSTER_SLOTS_END          = 1
	CLUSTER_SLOTS_SERVER_START = 2
)

// ServerGroup根据cluster slots和ReadPrefer得出
type ServerGroup struct {
	write string
	read  []string
}

type SlotTable struct {
	serverGroups []*ServerGroup
	// a cheap way to random select read backend
	counter uint32
}

func NewSlotTable() *SlotTable {
	st := &SlotTable{
		serverGroups: make([]*ServerGroup, NumSlots),
	}
	return st
}

func (st *SlotTable) WriteServer(slot int) string {
	return st.serverGroups[slot].write
}

func (st *SlotTable) ReadServer(slot int) string {
	st.counter += 1
	readServers := st.serverGroups[slot].read
	return readServers[st.counter%uint32(len(readServers))]
}

func (st *SlotTable) SetSlotInfo(si *SlotInfo) {
	for i := si.start; i <= si.end; i++ {
		st.serverGroups[i] = &ServerGroup{
			write: si.write,
			read:  si.read,
		}
	}
}

type SlotInfo struct {
	start int
	end   int
	write string
	read  []string
}

func NewSlotInfo(data *resp.Data) *SlotInfo {
	/*
	   cluster slots array element example
	   1) 1) (integer) 10923
	      2) (integer) 16383
	      3) 1) "10.4.17.164"
	         2) (integer) 7705
	      4) 1) "10.4.17.164"
	         2) (integer) 7708
	*/
	si := &SlotInfo{
		start: int(data.Array[CLUSTER_SLOTS_START].Integer),
		end:   int(data.Array[CLUSTER_SLOTS_END].Integer),
	}
	for i := CLUSTER_SLOTS_SERVER_START; i < len(data.Array); i++ {
		host := string(data.Array[i].Array[0].String)
		if len(host) == 0 {
			host = "127.0.0.1"
		}
		port := int(data.Array[i].Array[1].Integer)
		node := fmt.Sprintf("%s:%d", host, port)
		if i == CLUSTER_SLOTS_SERVER_START {
			si.write = node
		} else {
			si.read = append(si.read, node)
		}
	}
	return si
}

func Key2Slot(key string) int {
	buf := []byte(key)
	if pos := bytes.IndexByte(buf, '{'); pos != -1 {
		pos += 1
		if pos2 := bytes.IndexByte(buf[pos:], '}'); pos2 > 0 {
			slot := CRC16(buf[pos:pos+pos2]) % NumSlots
			return int(slot)
		}
	}
	slot := CRC16(buf) % NumSlots
	return int(slot)
}
