package common

import (
	"os"
)

func MustGetenv(name string) string {
	ret := os.Getenv(name)
	if ret == "" {
		Fatal(name + " not set")
	}
	return ret
}
