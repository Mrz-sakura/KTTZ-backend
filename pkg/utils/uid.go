package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenUniqueID(str string) string {
	unixTime := time.Now().Unix()
	rand.Seed(unixTime)
	return fmt.Sprintf("%08d_%s", unixTime%100000000+rand.Int63n(100), str)
}
