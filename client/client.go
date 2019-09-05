package main

import (
	pb ".."
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"strings"
	"sync"
)

func main() {

	var wg sync.WaitGroup

	commands := make(chan pb.CommandRequest, 100)

	displayQueue := make(chan string, 100)

	exitProgram := make(chan bool, 2)

	wg.Add(1)
	go outputHandler(displayQueue, &wg, exitProgram)

	wg.Add(1)
	go communicationHandler(commands, displayQueue, &wg, exitProgram)

	wg.Add(1)
	go inputHandler(commands, &wg, exitProgram)

	wg.Wait()
}

func inputHandler(commands chan<- pb.CommandRequest, wg *sync.WaitGroup, exitProgram chan<- bool) {

	defer wg.Done()

	for {
		reader := bufio.NewReader(os.Stdin)

		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")

		txt := strings.Split(text, " ")

		if strings.Compare(text, "EXIT") == 0 {
			exitProgram <- true
			return
		}

		params := ""
		if len(txt) > 1 {
			for _, v := range txt[1:] {
				params += v + " "
			}
		}
		params = strings.TrimSuffix(params, " ")

		commands <- pb.CommandRequest{Command: txt[0], Params:params}
	}
}

func outputHandler(displayQueue <-chan string, wg *sync.WaitGroup, exitProgram chan bool) {

	defer wg.Done()

	for {
		select {
		case inf := <-displayQueue:
			fmt.Println(inf)
		case <-exitProgram:
			fmt.Println("BYE!! ")
			exitProgram <- true
			return
		}
	}
}

func communicationHandler(commands <-chan pb.CommandRequest, displayQueue chan<- string, wg *sync.WaitGroup, exitProgram chan bool) {

	defer wg.Done()

	conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()
	cs := pb.NewCommandServiceClient(conn)
	srv, err := cs.CallCommand(context.Background())

	go func() {
		for {
			msg, err := srv.Recv()
			if err != nil {
				fmt.Printf("recv error", err)
				return
			}
			displayQueue <- "Result from command: " + msg.Command
			displayQueue <- msg.StandardOutput
			displayQueue <- msg.ErrorOutput
		}
	}()

	for {
		select {
		case comm := <-commands:
			_ = srv.Send(&comm)
		case <-exitProgram:
			exitProgram <- true
			_ = srv.CloseSend()
			return
		}
	}
}
