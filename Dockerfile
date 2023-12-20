FROM ubuntu:22.04 as builder
ARG VERSION
ARG GO_VERSION
ENV GOOS=linux CGO_ENABLED=1
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone && \
	apt update && apt upgrade -y && apt install -y make git gcc g++ ca-certificates curl && \
	bash -c "source <(curl -L https://go-install.netlify.app/install.sh) -v ${GO_VERSION}" && \
	ln -sf /usr/local/go/bin/go /usr/local/bin

RUN	go install github.com/aura-studio/syncloud@${VERSION}

FROM ubuntu:22.04
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone
COPY --from=builder /root/go/bin/syncloud /syncloud

ENTRYPOINT ["/syncloud"]