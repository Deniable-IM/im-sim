package ipv4

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"deniable-im/im-sim/internal/utils/fn"
)

func IPv4AddressSpace(iprange string) (map[string]struct{}, error) {
	mask, err := strconv.Atoi(strings.Split(iprange, "/")[1])
	if err != nil {
		return nil, fmt.Errorf("Network assiged ip mask not valid: %w", err)
	}

	highestAvailableIP := uint32(math.Pow(2, (32-float64(mask))) - 2)
	lowestAvailableIP := IPv4StringToDecimal(strings.Split(iprange, "/")[0])

	addressSet := make(map[string]struct{})
	for i := lowestAvailableIP; i < lowestAvailableIP+highestAvailableIP; i++ {
		addressSet[IPv4DecimalToString(i)] = struct{}{}
	}

	return addressSet, nil
}

func IPv4StringToDecimal(ipv4 string) uint32 {
	ip := fn.Map(strings.Split(ipv4, "."),
		func(e string) uint32 {
			n, _ := strconv.Atoi(e)
			return uint32(n)
		})
	a := ip[0] << 24
	b := ip[1] << 16
	c := ip[2] << 8
	d := ip[3] ^ 1
	return a ^ b ^ c ^ d
}

func IPv4DecimalToString(ipv4 uint32) string {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, ipv4)
	return fmt.Sprintf("%d.%d.%d.%d",
		bytes[0],
		bytes[1],
		bytes[2],
		bytes[3],
	)
}
