package Messagemaker

import (
	Types "deniable-im/im-sim/pkg/simulation/types"
	"fmt"
)

func MakeDenimProtocolMessage(msg Types.Msg) Types.Msg {
	result := Types.Msg{
		To:         msg.To,
		From:       msg.From,
		IsDeniable: msg.IsDeniable,
	}

	if msg.IsDeniable {
		result.MsgContent = fmt.Sprintf("denim:%v:%v", msg.To, msg.MsgContent)
	} else {
		result.MsgContent = fmt.Sprintf("send:%v:%v", msg.To, msg.MsgContent)
	}

	return result
}
