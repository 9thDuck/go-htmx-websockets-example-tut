package utils

import (
	"log"
)

func ThrowOnError(name string, err error) {
	if err != nil {
		log.Fatalf("%s error\nerror details:%v", name, err)
	}
}
