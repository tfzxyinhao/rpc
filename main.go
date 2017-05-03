package main

import (
	"flag"

	"sync"

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
				var w sync.WaitGroup
				w.Add(2)
				gservice.RegisterService(&w)
				go gservice.ServService(&w)
				w.Wait()
			}
		}
	}

}
