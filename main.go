package main

import (
	"flag"

	"github.com/tfzxyinhao/rpc/gservice"
)

func main() {
	t := flag.String("t", "s", "-t <s|c>")
	flag.Parse()

	if t != nil {
		switch *t {
		case "c":
			gservice.ClientTestService()
		case "c1":
			gservice.ClientTestServiceDirect()
		case "s":
			{
				gservice.RegisterService()
				gservice.ServService()
			}
		}
	}

}
