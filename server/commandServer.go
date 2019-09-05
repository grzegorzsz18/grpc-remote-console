package main

import (
	pb ".."
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type CommandService struct{}

func createResponse(stdn string, err string, command string) pb.CommandResponse {
	return pb.CommandResponse{
		StandardOutput: stdn,
		ErrorOutput:    err,
		Command:        command,
	}
}

func (CommandService) CallCommand(client pb.CommandService_CallCommandServer) error {

	ctx := client.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

		}
		recv, err := client.Recv()
		if err != nil {
			fmt.Println("recv error")
			return nil
		}
		go func(request *pb.CommandRequest) {
			resp := executeCommand(request)
			_ = client.Send(&resp)
		}(recv)
	}
}

func executeCommand(command *pb.CommandRequest) pb.CommandResponse {
	comm := command.Command
	flagsString := command.Params
	flags := strings.Split(flagsString, " ")
	for i := range flags {
		flags[i] = strings.Replace(flags[i], " ", "", -1)
	}

	fmt.Println(comm, flags, len(flags))
	var cmd *exec.Cmd
	var out bytes.Buffer
	var stdErr bytes.Buffer
	if len(flags) > 1 {
		cmd = exec.Command(comm, flags...)
	} else {
		cmd = exec.Command(comm)
	}
	cmd.Stderr = &stdErr
	cmd.Stdout = &out

	err := cmd.Run()
	resultStdErr := ""
	if err != nil {
		log.Printf("error {}", err)
		resultStdErr = stdErr.String()
	}

	return createResponse(out.String(), resultStdErr, comm+" "+flagsString)
}
