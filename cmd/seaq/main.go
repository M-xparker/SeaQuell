package main

import (
	"bufio"
	"fmt"
	"github.com/MattParker89/seaquell/machine"
	"github.com/MattParker89/seaquell/parse"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Kill, os.Interrupt)

	text := make(chan string)
	bio := bufio.NewReader(os.Stdin)
	go func(t chan string) {
		for {
			text, _ := bio.ReadString('\n')
			t <- text
		}
	}(text)

	m := machine.New("test.db")
	defer m.Close()

	fmt.Print(">")
	for {
		select {
		case <-ch:
			fmt.Println("close")
			return
		case t := <-text:
			rom := parse.Generate(t)
			res := m.Exec(rom)
			if len(res) > 0 {
				r := res[0]
				for _, col := range r.Columns {
					fmt.Printf("|%10s|", col)
				}
				fmt.Print("\n")
			}
			for _, r := range res {
				for _, d := range r.Data {
					fmt.Printf("|%10v|", d)
				}
				fmt.Print("\n")
			}

			fmt.Print(">")
		default:
		}

	}

}
