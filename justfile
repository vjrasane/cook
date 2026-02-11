hooks:
    @devenv tasks run devenv:git-hooks:run

build-listonic:
    cd listonic-cli && CGO_ENABLED=0 go build -o listonic .

test-listonic:
    cd listonic-cli && go test -v ./...

listonic *args:
    @cd listonic-cli && go run . {{args}}
