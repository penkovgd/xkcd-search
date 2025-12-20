lint: protolint golint

protobuf:
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/words/words.proto
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/update/update.proto
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/search/search.proto

protolint:
	protolint .

golint:
	golangci-lint run -E gocritic -v ./...

test:
	CGO_ENABLED=1 go test -race -coverprofile cover.out \
		$(shell go list ./... | egrep -v 'yadro.com/course/(proto|api$$|update$$|words$$|search$$|search/core/mocks)')
	go tool cover -func=cover.out | tail -1
	go tool cover -html cover.out -o cover.html
