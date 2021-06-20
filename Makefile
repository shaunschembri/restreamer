build:
	@ go build -o restreamer -ldflags="-s -w" cmd/restreamer/main.go

lint:
	@ ~/go/bin/golangci-lint run --fix --deadline=10s

snapshot:
	@ goreleaser --snapshot --skip-publish --rm-dist

tag:
	@ git tag -f -a ${TAG} -m ${TAG}
	@ git push -f origin ${TAG}