package silent

import (
	"fmt"
	"io"
	"log"
)

func Close(srv io.Closer) {
	if err := srv.Close(); err != nil {
		log.Printf("Error while closing: %s", err)
	}
}

func PanicOnErr(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			fmt.Printf("%s: %s\n", msg[0], err)
		}
		panic(err)
	}
}
