package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/adrianliechti/llama/pkg/provider/custom"

	"google.golang.org/grpc"
)

func main() {
	l, err := net.Listen("tcp", ":50051")

	if err != nil {
		panic(err)
	}

	s := &Server{}

	server := grpc.NewServer()

	custom.RegisterCompleterServer(server, s)

	if err := server.Serve(l); err != nil {
		panic(err)
	}
}

type Server struct {
	custom.UnsafeCompleterServer
}

func (*Server) Complete(r *custom.CompletionRequest, stream custom.Completer_CompleteServer) error {
	for _, m := range r.Messages {
		println(m.Role.String() + ": " + m.Content)
	}

	println("")
	println("")

	last := r.Messages[len(r.Messages)-1]

	completion := &custom.Completion{
		Id: fmt.Sprintf("%d", time.Now().UnixMicro()),

		Message: &custom.Message{
			Role:    custom.Role_ROLE_ASSISTANT,
			Content: strings.ToUpper(last.Content) + "!!!",
		},
	}

	return stream.Send(completion)
}
