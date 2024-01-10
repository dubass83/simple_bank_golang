package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "qazwsxedcrfvtgbyhnujmikolp"

func init() {
	rand.NewSource(time.Now().UnixNano())
}

// RandomInt return random int64 between min and max values
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString return random string of given length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomOwner generate a random Owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generate a random amount of Money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generate a random currency code
func RandomCurrency() string {
	cur := []string{
		USD,
		EUR,
		UAH,
	}
	return cur[rand.Intn(len(cur))]
}

func RandomEmail() string {
	return fmt.Sprintf("%s@example.com", RandomString(7))
}
