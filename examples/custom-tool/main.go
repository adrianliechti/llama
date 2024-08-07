package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/adrianliechti/llama/pkg/jsonschema"
	"github.com/adrianliechti/llama/pkg/tool/custom"

	"google.golang.org/grpc"
)

func main() {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 6666))

	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	custom.RegisterToolServer(s, newServer())
	s.Serve(l)
}

type server struct {
	custom.UnsafeToolServer
}

func newServer() *server {
	return &server{}
}

func (s *server) Info(context.Context, *custom.InfoRequest) (*custom.Definition, error) {
	name := "kubectl"
	description := "invoke the Kubernetes CLI kubectl with the given arguments"

	schema, _ := json.Marshal(jsonschema.Definition{
		Type: "object",

		Properties: map[string]jsonschema.Definition{
			"args": {
				Type: "array",
				Items: &jsonschema.Definition{
					Type: "string",
				},
			},
		},

		Required: []string{"args"},
	})

	return &custom.Definition{
		Name:        name,
		Description: description,
		Schema:      string(schema),
	}, nil
}

func (s *server) Execute(ctx context.Context, r *custom.ExecuteRequest) (*custom.Result, error) {
	var input struct {
		Args []string `json:"args"`
	}

	if err := json.Unmarshal([]byte(r.Parameter), &input); err != nil {
		return nil, err
	}

	args := input.Args
	args = append(args, "-o", "wide")

	println("$ kubectl " + strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "kubectl", args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	result := map[string]any{
		"stdout": string(output),
		"stderr": "",
	}

	content, _ := json.Marshal(result)

	return &custom.Result{
		Content: string(content),
	}, nil
}
