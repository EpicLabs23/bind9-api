FROM ubuntu/bind9:latest

ENV DEBIAN_FRONTEND=noninteractive

# Install Go + deps
RUN apt-get update && apt-get install -y \
    curl git bash \
    && rm -rf /var/lib/apt/lists/*

# Install bind9utils
RUN apt-get update && apt-get install -y bind9utils && rm -rf /var/lib/apt/lists/*

# Install vim
RUN apt-get update && apt-get install -y vim && rm -rf /var/lib/apt/lists/*

# Install Go
ENV GOLANG_VERSION=1.24.3
RUN curl -OL https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz \
    && rm go${GOLANG_VERSION}.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

# Install Air
RUN go install github.com/air-verse/air@latest
ENV PATH="/root/go/bin:${PATH}"

# Keep bind9 default entrypoint & cmd
EXPOSE 53/udp
EXPOSE 53/tcp
EXPOSE 8053/tcp
