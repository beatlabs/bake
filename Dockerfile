FROM golang:1.15-alpine

RUN apk add --no-cache ca-certificates wget bash git docker-cli tar gcc musl-dev yarn

# CGO is required by some modules like https://github.com/uber/h3-go
ENV CGO_ENABLED=1

# Required to access private modules
ENV GOPRIVATE=github.com/taxibeat/*

# Download and install mage file into bin path
RUN wget -qc https://github.com/magefile/mage/releases/download/v1.9.0/mage_1.9.0_Linux-64bit.tar.gz -O - | tar -xz -C /usr/bin mage

# Download and install hadolint into bin path
RUN wget -qO /usr/bin/hadolint https://github.com/hadolint/hadolint/releases/download/v1.17.6/hadolint-Linux-x86_64 && chmod +x /usr/bin/hadolint

# Download and install swag into bin path
RUN wget -qc https://github.com/swaggo/swag/releases/download/v1.6.6/swag_1.6.6_Linux_x86_64.tar.gz -O - | tar -xz -C /usr/bin swag

# Download and install mark into bin path
RUN wget -qc https://github.com/mantzas/mark/releases/download/v0.9.0/mark-linux-x64.tar.gz -O - | tar -xz -C /usr/bin mark

# Download and install helm 3 into bin path
RUN wget -qc https://get.helm.sh/helm-v3.2.4-linux-amd64.tar.gz -O - | tar -xz -C /tmp && mv /tmp/linux-amd64/helm /usr/bin && rm -rf /tmp/linux-amd

# Download and install golangci-lint into go bin path
RUN wget -qc https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -O - | /bin/sh -s -- -b $(go env GOPATH)/bin v1.33.0

# Download and install goveralls - go coveralls client
RUN go get github.com/mattn/goveralls

# Very permissive because we don't know what user the container will run as
RUN mkdir /home/beat && chmod 777 /home/beat
ENV HOME /home/beat

COPY bake.sh /home/beat/bake-default.sh

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["bash", "/entrypoint.sh"]
