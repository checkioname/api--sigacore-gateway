package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijqlmnopqrstuvwxyz"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomInt(minValue, maxValue int64) int64 {
	return minValue + rand.Int63n(maxValue-minValue+1)
}

func randomString(n int) string {
	sb := strings.Builder{}
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// Custom implementations

func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "GBP", "BRL"}
	return currencies[rand.Intn(len(currencies))]
}

func RandomMoney() int64 {
	return RandomInt(1000, 100000)
}

func RandomOwner() string {
	return randomString(6)
}
