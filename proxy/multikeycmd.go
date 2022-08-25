package proxy

import (
	"bufio"
	"bytes"

	"github.com/collinmsn/resp"
	log "github.com/ngaut/logging"
)

const (
	MGET = iota
	MSET
	DEL
)

var (
	OK_DATA *resp.Data
)

func init() {
	OK_DATA = &resp.Data{
		T:      resp.T_SimpleString,
		String: []byte("OK"),
	}
}

/*
multi key cmd被拆分成numKeys个子请求按普通的pipeline request发送，最后在写出response时进行合并
当最后一个子请求的response到来时，整个multi key cmd完成，拼接最终response并写出

只要有一个子请求失败，都认定整个请求失败
多个子请求共享一个request sequence number

请求的失败包含两种类型：1、网络失败，比如读取超时，2，请求错误，比如本来该在A机器上，请求到了B机器上，表现为response type为error
*/
type MultiKeyCmd struct {
	cmd               *resp.Command
	cmdType           int
	numSubCmds        int
	numPendingSubCmds int
	subCmdRsps        []*PipelineResponse
}

func NewMultiKeyCmd(cmd *resp.Command, numSubCmds int) *MultiKeyCmd {
	mc := &MultiKeyCmd{
		cmd:               cmd,
		numSubCmds:        numSubCmds,
		numPendingSubCmds: numSubCmds,
	}
	switch cmd.Name() {
	case "MGET":
		mc.cmdType = MGET
	case "MSET":
		mc.cmdType = MSET
	case "DEL":
		mc.cmdType = DEL
	default:
		panic("not multi key command")
	}
	mc.subCmdRsps = make([]*PipelineResponse, numSubCmds)
	return mc
}

func (mc *MultiKeyCmd) OnSubCmdFinished(rsp *PipelineResponse) {
	mc.subCmdRsps[rsp.ctx.subSeq] = rsp
	mc.numPendingSubCmds--
}

func (mc *MultiKeyCmd) Finished() bool {
	return mc.numPendingSubCmds == 0
}

func (mc *MultiKeyCmd) CoalesceRsp() *PipelineResponse {
	plRsp := &PipelineResponse{}
	var rsp *resp.Data
	switch mc.CmdType() {
	case MGET:
		rsp = &resp.Data{T: resp.T_Array, Array: make([]*resp.Data, mc.numSubCmds)}
	case MSET:
		rsp = OK_DATA
	case DEL:
		rsp = &resp.Data{T: resp.T_Integer}
	default:
		panic("invalid multi key cmd name")
	}
	for i, subCmdRsp := range mc.subCmdRsps {
		if subCmdRsp.err != nil {
			rsp = &resp.Data{T: resp.T_Error, String: []byte(subCmdRsp.err.Error())}
			break
		}
		reader := bufio.NewReader(bytes.NewReader(subCmdRsp.rsp.Raw()))
		data, err := resp.ReadData(reader)
		if err != nil {
			log.Errorf("re-parse response err=%s", err)
			rsp = &resp.Data{T: resp.T_Error, String: []byte(err.Error())}
			break
		}
		if data.T == resp.T_Error {
			rsp = data
			break
		}
		switch mc.CmdType() {
		case MGET:
			rsp.Array[i] = data
		case MSET:
		case DEL:
			rsp.Integer += data.Integer
		default:
			panic("invalid multi key cmd name")
		}
	}
	plRsp.rsp = resp.NewObjectFromData(rsp)
	return plRsp
}

func (mc *MultiKeyCmd) CmdType() int {
	return mc.cmdType
}

func IsMultiCmd(cmd *resp.Command) (multiKey bool, numKeys int) {
	multiKey = true
	switch cmd.Name() {
	case "MGET":
		numKeys = len(cmd.Args) - 1
	case "MSET":
		numKeys = (len(cmd.Args) - 1) / 2
	case "DEL":
		numKeys = len(cmd.Args) - 1
	default:
		multiKey = false
	}
	return
}
