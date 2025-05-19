package Behavior

import (
	Messagemaker "deniable-im/im-sim/pkg/simulation/messagemaker"
	Messageparser "deniable-im/im-sim/pkg/simulation/messageparser"
	Types "deniable-im/im-sim/pkg/simulation/types"
	"fmt"
	"math/rand"
	"strings"
	"time"

	fuzz "github.com/google/gofuzz"
)

type SimpleHumanTraits struct {
	Name              string
	SendProp          float64
	ResponseProb      float64
	DeniableProb      float64
	BurstModifier     float64
	DeniableBurstSize int32
	DeniableCount     int32
	User              *Types.SimUser
	nextMsgFunc       func(*SimpleHumanTraits) int
	randomizer        *rand.Rand
	nextSendTime      time.Time
}

func (sh *SimpleHumanTraits) GetBehaviorName() string {
	if sh == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Simple Human Traits Behavior with name %v", sh.Name)
	return b.String()
}

func (sh *SimpleHumanTraits) GetNextMessageTime() int {
	if sh == nil {
		return 0
	}

	next := sh.nextMsgFunc(sh)

	if sh.IsBursting() {
		return next
	}

	//Calculate the next time a message will be sent
	for {
		if sh.SendRegularMsg() {
			break
		}

		next += sh.nextMsgFunc(sh)
	}

	return next
}

func (sh *SimpleHumanTraits) GetRandomizer() *rand.Rand {
	return sh.randomizer
}

func (sh *SimpleHumanTraits) WillRespond(msg Types.Msg) bool {
	if sh == nil {
		return false
	}

	if msg.From == "Unknown Sender" {
		return false
	}

	return sh.randomizer.Float64() > (1.0 - sh.ResponseProb)
}

func (sh *SimpleHumanTraits) SendRegularMsg() bool {
	if sh == nil {
		return false
	}

	return sh.randomizer.Float64() > (1.0 - sh.SendProp)
}

func (sh *SimpleHumanTraits) SendDeniableMsg() bool {
	if sh == nil {
		return false
	}

	return sh.randomizer.Float64() > (1.0 - sh.DeniableProb)
}

func (sh *SimpleHumanTraits) GetResponseTime() int {
	if sh == nil {
		return 0
	}

	delta := time.Now().Sub(sh.nextSendTime)

	time := int32(delta)
	//Clause to avoid randomizer panicking. 1 ms difference is most likely not a problem
	if time < 1 {
		return 0
	}

	return int(sh.randomizer.Int31n((time)))
}

func (sh *SimpleHumanTraits) IncrementDeniableCount() {
	sh.DeniableCount += sh.DeniableBurstSize
}

func (sh *SimpleHumanTraits) IsBursting() bool {
	return sh.DeniableCount > 0
}

func (sh *SimpleHumanTraits) ParseIncoming(incoming string) (*Types.Msg, error) {
	return Messageparser.DenimParser(incoming)
}

func (sh *SimpleHumanTraits) MakeReply(msg Types.Msg) Types.Msg {
	response := Types.Msg{
		To:         msg.From,
		From:       msg.To,
		IsDeniable: msg.IsDeniable,
		MsgContent: Messagemaker.GetQuoteByIndexSafe(sh.randomizer.Int()),
	}

	if response.IsDeniable {
		sh.IncrementDeniableCount()
	} else {
		sh.DeniableCount -= 1
	}
	return Messagemaker.MakeDenimProtocolMessage(response)
}

func (sh *SimpleHumanTraits) MakeMessages() []Types.Msg {
	var msgs []Types.Msg

	//Deniable messages are made first to allow them to piggyback on the regular messages
	if sh.SendDeniableMsg() && len(sh.User.DeniableContactList) != 0 {
		den_target := sh.User.DeniableContactList[sh.randomizer.Intn(len(sh.User.DeniableContactList))]
		den_msg := Types.Msg{
			To:         den_target,
			From:       fmt.Sprintf("%v", sh.User.ID),
			MsgContent: Messagemaker.GetQuoteByIndexSafe(sh.randomizer.Int()),
			IsDeniable: true,
		}

		sh.IncrementDeniableCount()
		msgs = append(msgs, den_msg)
	}

	//Mayhaps make more than one regular message per call? Idk anymore, all of this is horrible to simulate
	reg_target := sh.User.RegularContactList[sh.randomizer.Intn(len(sh.User.RegularContactList))]
	reg_msg := Types.Msg{
		To:         reg_target,
		From:       fmt.Sprintf("%v", sh.User.ID),
		MsgContent: Messagemaker.GetQuoteByIndexSafe(sh.randomizer.Int()),
		IsDeniable: false,
	}

	msgs = append(msgs, reg_msg)

	for i, msg := range msgs {
		msgs[i] = Messagemaker.MakeDenimProtocolMessage(msg)
	}

	return msgs
}

func NewSimpleHumanTraits(
	name string,
	send_prop, response, deniable_prop, burst_mod float64,
	deniable_burst_size int32,
	next_func func(*SimpleHumanTraits) int,
	r *rand.Rand) *SimpleHumanTraits {
	return &SimpleHumanTraits{
		Name:              name,
		SendProp:          send_prop,
		ResponseProb:      response,
		DeniableProb:      deniable_prop,
		BurstModifier:     burst_mod,
		DeniableBurstSize: deniable_burst_size,
		nextMsgFunc:       next_func,
		randomizer:        r,
	}
}

func FuzzedNewSimpleHumanTraits(
	fuzzer fuzz.Fuzzer,
	next_func func(*SimpleHumanTraits) int,
	r *rand.Rand) *SimpleHumanTraits {
	var sh SimpleHumanTraits
	fuzzer.Fuzz(&sh)
	sh.nextMsgFunc = next_func
	sh.randomizer = r
	sh.DeniableBurstSize = sh.DeniableBurstSize % 10

	return &sh
}
