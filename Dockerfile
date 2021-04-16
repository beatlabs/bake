FROM golang:1.15 as builder

ARG GH_TOKEN

# Install Skim
RUN git config --global url."https://$GH_TOKEN@github.com/".insteadOf "https://github.com/" && \
    go get github.com/taxibeat/skim/cmd/skim && rm -rf /go/src/github.com/taxibeat/ && \
    git config --global --remove-section url."https://$GH_TOKEN@github.com/"

FROM golang:1.15

COPY --from=builder /go/bin/skim /go/bin/skim

RUN apt-get update && \
    apt-get install -y \
    --no-install-recommends \
    apt-transport-https \
    protobuf-compiler \
    unzip \
    ca-certificates \
    gnupg-agent \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

ENV APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE=1
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \	
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable" && \	
    apt-get -y update && \	
    apt-get install -y docker-ce \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# CGO is required by some modules like https://github.com/uber/h3-go
ENV CGO_ENABLED=1

# Required to access private modules
ENV GOPRIVATE=github.com/taxibeat/*

# Skim dependencies
ARG BUF_VERSION=0.24.0
ARG PROTOC_VERSION=3.13.0
ARG PROTODOC_VERSION=1.3.2

WORKDIR /usr/local/bin
RUN curl -sSLO https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-x86_64 && \
    mv buf-Linux-x86_64 buf-linux && \
    curl -sSLO https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/protoc-gen-buf-check-breaking-Linux-x86_64 && \
    mv protoc-gen-buf-check-breaking-Linux-x86_64 protoc-gen-buf-check-breaking && \
    curl -sSLO https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/protoc-gen-buf-check-lint-Linux-x86_64 && \
    mv protoc-gen-buf-check-lint-Linux-x86_64 protoc-gen-buf-check-lint && \
    chmod +x buf-linux protoc* && \
    curl -sSLO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d protoc && \
    chmod +x protoc/bin/* && \
    cp -r /usr/local/bin/protoc/include/google /usr/include/ && \
    chmod -R 755 /usr/include/google && \
    curl -sSLO https://github.com/pseudomuto/protoc-gen-doc/releases/download/v${PROTODOC_VERSION}/protoc-gen-doc-${PROTODOC_VERSION}.linux-amd64.go1.12.6.tar.gz && \
    tar xf protoc-gen-doc-${PROTODOC_VERSION}.linux-amd64.go1.12.6.tar.gz && \
    mv protoc-gen-doc-${PROTODOC_VERSION}.linux-amd64.go1.12.6/protoc-gen-doc . && \
    go get -u google.golang.org/protobuf/cmd/protoc-gen-go && \
    GOBIN=/ go install google.golang.org/protobuf/cmd/protoc-gen-go

WORKDIR /go

# Download and install mage file into bin path
ARG MAGE_VERSION=1.11.0
RUN wget -qc https://github.com/magefile/mage/releases/download/v${MAGE_VERSION}/mage_${MAGE_VERSION}_Linux-64bit.tar.gz -O - | tar -xz -C /usr/bin mage

# Download and install hadolint into bin path
ARG HADOLINT_VERSION=1.17.6
RUN wget -qO /usr/bin/hadolint https://github.com/hadolint/hadolint/releases/download/v${HADOLINT_VERSION}/hadolint-Linux-x86_64 && chmod +x /usr/bin/hadolint

# Download and install swag into bin path
ARG SWAG_VERSION=1.6.6
RUN wget -qc https://github.com/swaggo/swag/releases/download/v${SWAG_VERSION}/swag_${SWAG_VERSION}_Linux_x86_64.tar.gz -O - | tar -xz -C /usr/bin swag

# Download and install mark into bin path
ARG MARK_VERSION=5.6
RUN wget -qc https://github.com/kovetskiy/mark/releases/download/${MARK_VERSION}/mark_${MARK_VERSION}_Linux_x86_64.tar.gz -O - | tar -xz -C /usr/bin mark

# Download and install helm 3 into bin path
ARG HELM_VERSION=3.2.4
RUN wget -qc https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xz -C /tmp && mv /tmp/linux-amd64/helm /usr/bin && rm -rf /tmp/linux-amd

# Download and install golangci-lint into go bin path
ARG GOLANGCILINT_VERSION=1.33.0
RUN wget -qc https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -O - | /bin/sh -s -- -b "$(go env GOPATH)/bin" v${GOLANGCILINT_VERSION}

# Very permissive because we don't know what user the container will run as
RUN mkdir /home/beat && chmod 777 /home/beat
ENV HOME /home/beat

COPY bake.sh /home/beat/bake-default.sh

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["bash", "/entrypoint.sh"]
