FROM golang:1.25

WORKDIR /opt/app

RUN touch /usr/bin/go-prisma && \
	echo "go run github.com/steebchen/prisma-client-go" > /usr/bin/go-prisma && \
	chmod +x /usr/bin/go-prisma

RUN go install github.com/melkeydev/go-blueprint@latest && \
	go install github.com/air-verse/air@latest 

RUN mkdir /.cache && chmod -R 777 /.cache
RUN chmod -R 1777 "$GOPATH"
