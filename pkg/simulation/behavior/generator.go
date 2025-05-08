package Behavior

import (
	"fmt"
	"math/rand"
)

const MAX_MIN_DIFF = 0.2

// Based on user 90, 44, 9, 80?, 82, 99, 16, 31, 0, 69, 58
// 58: send: 0.375, resp: 0.452
// 69: send: 0.584, resp: 0.334
// 00: send: 0.751, resp: 0.387
// 31: send: 0.285, resp: 0.464
// 16: send: 0.182, resp: 0.249
// 99: send: 0.200, resp: 0.432
// 82: send: 0.380, resp: 0.843
// 80: send: 0.357, resp: 0.735
// 09: send: 0.181, resp: 0.180
// 44: send: 0.184, resp: 0.279
// 58: send: 0.161, resp: 0.380

func GenerateRealisticSimpleHumanTraits(count int, r *rand.Rand, nfunc func(*SimpleHumanTraits) float64) []*SimpleHumanTraits {

	traits := make([]*SimpleHumanTraits, count)
	goodSendValues := []float64{0.375, 0.584, 0.751, 0.285, 0.182, 0.200, 0.380, 0.357, 0.181, 0.184, 0.161}
	var goodSendAvg float64
	for _, v := range goodSendValues {
		goodSendAvg += v
	}
	goodSendAvg /= float64(len(goodSendValues))

	goodReplyValues := []float64{0.452, 0.334, 0.387, 0.464, 0.249, 0.432, 0.843, 0.735, 0.180, 0.279, 0.380}
	var goodReplyAvg float64
	for _, v := range goodReplyValues {
		goodReplyAvg += v
	}
	goodReplyAvg /= float64(len(goodReplyValues))

	fmt.Printf("Avg send: %v \n", goodSendAvg)
	fmt.Printf("Avg reply: %v \n", goodReplyAvg)

	for i := range traits {
		var send, reply float64
		var rand_param *rand.Rand
		if r != nil {
			send = r.Float64()*MAX_MIN_DIFF + (goodSendAvg - MAX_MIN_DIFF/2)
			reply = r.Float64()*MAX_MIN_DIFF + (goodReplyAvg - MAX_MIN_DIFF/2)
			rand_param = r
		} else {
			send = rand.Float64()*MAX_MIN_DIFF + (goodSendAvg - MAX_MIN_DIFF/2)
			reply = rand.Float64()*MAX_MIN_DIFF + (goodReplyAvg - MAX_MIN_DIFF/2)
			rand_param = rand.New(rand.NewSource(rand.Int63()))
		}
		fmt.Printf("Pre increment send: %v, reply: %v \n", send, reply)
		for reply <= send {
			println("Incrementing reply...")
			reply = rand_param.Float64()*0.1 + reply
		}
		fmt.Printf("Post increment send: %v, reply: %v \n", send, reply)

		traits[i] = NewSimpleHumanTraits(fmt.Sprintf("%v", i), send, reply, 0.1, 0.1, 10, nfunc, rand_param)
	}

	return traits
}
