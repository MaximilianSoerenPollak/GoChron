VERSION=0.0

all:
	go build -ldflags "-X github.com/MaximilianSoerenPollak/zeit/z.VERSION=$(VERSION)"
