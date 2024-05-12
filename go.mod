module github.com/adrianliechti/llama

go 1.22

replace github.com/sashabaranov/go-openai v1.23.1 => github.com/adrianliechti/go-openai v0.0.0-20240322224346-47657b844843

require (
	github.com/coreos/go-oidc/v3 v3.10.0
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-chi/cors v1.2.1
	github.com/google/uuid v1.6.0
	github.com/sashabaranov/go-openai v1.23.1
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.63.2
	google.golang.org/protobuf v1.34.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-jose/go-jose/v4 v4.0.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240314234333-6e1732d8331c // indirect
)
