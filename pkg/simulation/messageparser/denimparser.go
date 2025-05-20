package Messageparser

import (
	Types "deniable-im/im-sim/pkg/simulation/types"
	"fmt"
	"strings"
)

func DenimParser(incoming string) (*Types.Msg, error) {
	err := fmt.Errorf("Failed to parse incoming string")

	splits := strings.Split(incoming, ":")
	if splits[0] == "" || splits[0] == "\n" {
		return nil, err
	}

	sender := splits[0]
	if len(sender) < 8 {
		return nil, err
	}

	if len(splits) < 2 {
		return nil, err
	}

	isDeniable := strings.Contains(strings.ToLower(sender), "deniable")

	if !isDeniable {
		sender = sender[8:]
		if sender == "\n" {
			return nil, err
		}
	}

	var msg *Types.Msg

	if isDeniable {
		sender = sender[9:]
		msg = &Types.Msg{To: "", From: sender, MsgContent: splits[1], IsDeniable: true}
	} else {
		msg = &Types.Msg{To: "", From: sender, MsgContent: splits[1], IsDeniable: false}
	}

	return msg, nil
}
