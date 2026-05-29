FROM golang:1.25

WORKDIR /opt/app


RUN curl -sSf https://atlasgo.sh | sh && \
	go install github.com/melkeydev/go-blueprint@latest && \
	go install github.com/air-verse/air@latest 

RUN mkdir /.cache && chmod -R 777 /.cache
RUN chmod -R 1777 "$GOPATH"

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
