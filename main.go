package main

import (
	"flag"
	"fmt"

	oi "github.com/reiver/go-oi"
	telnet "github.com/reiver/go-telnet"

	termbox "github.com/nsf/termbox-go"
)

func main() {
	host := flag.String("h", "localhost", "host")
	port := flag.String("p", "23", "port")
	flag.Parse()

	// Init termbox to use the getchar
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	termbox.SetCursor(0, 0)
	termbox.HideCursor()

	if err := telnet.DialToAndCall(*host+":"+*port, caller{}); err != nil {
		fmt.Println(err)
	}
}

type caller struct{}

func (c caller) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
	quit := make(chan struct{}, 1)
	defer close(quit)

	// Write to telnet server
	readBlocker := make(chan struct{}, 1)
	defer close(readBlocker)
	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if ev.Key == termbox.KeyCtrlC {
					quit <- struct{}{}
					return
				}
				if isASCII(ev.Ch) {
					fmt.Printf("%c", ev.Ch)
					oi.LongWrite(w, []byte{byte(ev.Ch)})
					readBlocker <- struct{}{}
				}
			}
		}
	}()

	// Read from telnet server
	go func() {
		var buffer [1]byte
		p := buffer[:]
		for {
			<-readBlocker
			r.Read(p)
			fmt.Printf("%c", p[0])
		}
	}()

	<-quit
}

func isASCII(r rune) bool {
	return r <= '~'
}
