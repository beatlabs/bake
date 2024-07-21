FROM golang:1.22
ARG TARGETARCH
RUN echo Building bake image for ${TARGETARCH} architecture

RUN apt-get update && \
  apt-get install -y \
  --no-install-recommends \
  unzip \
  ca-certificates \
  gnupg-agent \
  software-properties-common

RUN curl -fsSL https://get.docker.com | sh

RUN rm -rf /var/lib/apt/lists/*

# CGO is required by some modules like https://github.com/uber/h3-go
ENV CGO_ENABLED=1

WORKDIR /go

# Download and install mage file into bin path
ARG MAGE_VERSION=1.15.0
RUN case "${TARGETARCH}" in \
  "amd64")  MAGE_ARCH=64bit  ;; \
  "arm64")  MAGE_ARCH=ARM64  ;; \
  esac && \
  wget -qc "https://github.com/magefile/mage/releases/download/v${MAGE_VERSION}/mage_${MAGE_VERSION}_Linux-${MAGE_ARCH}.tar.gz" -O - | tar -xz -C /usr/bin mage

# Download and install hadolint into bin path
ARG HADOLINT_VERSION=2.12.0
RUN case "${TARGETARCH}" in \
  "amd64")  HADOLINT_ARCH=x86_64  ;; \
  "arm64")  HADOLINT_ARCH=arm64  ;; \
  esac && \
  wget -qO /usr/bin/hadolint "https://github.com/hadolint/hadolint/releases/download/v${HADOLINT_VERSION}/hadolint-Linux-${HADOLINT_ARCH}" && chmod +x /usr/bin/hadolint

# Download and install helm 3 into bin path
ARG HELM_VERSION=3.15.3
RUN case "${TARGETARCH}" in \
  "amd64")  HELM_ARCH=amd64  ;; \
  "arm64")  HELM_ARCH=arm64  ;; \
  esac && \
  wget -qc "https://get.helm.sh/helm-v${HELM_VERSION}-linux-${HELM_ARCH}.tar.gz" -O - | tar -xz -C /tmp && mv "/tmp/linux-${HELM_ARCH}/helm" /usr/bin && rm -rf "/tmp/linux-${HELM_ARCH}"

# Download and install golangci-lint into go bin path
ARG GOLANGCILINT_VERSION=1.59.1
RUN wget -qc https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -O - | /bin/sh -s -- -b "$(go env GOPATH)/bin" v${GOLANGCILINT_VERSION}

# Download and install promtool
# https://prometheus.io/download/
ARG PROMTOOL_VERSION=2.53.1
RUN case "${TARGETARCH}" in \
  "amd64")  PROMTOOL_ARCH=amd64  ;; \
  "arm64")  PROMTOOL_ARCH=arm64  ;; \
  esac && \
  wget -qc "https://github.com/prometheus/prometheus/releases/download/v${PROMTOOL_VERSION}/prometheus-${PROMTOOL_VERSION}.linux-${PROMTOOL_ARCH}.tar.gz" -O - | tar -xz -C /tmp && mv "/tmp/prometheus-${PROMTOOL_VERSION}.linux-${PROMTOOL_ARCH}/promtool" /usr/bin && rm -rf "/tmp/prometheus-${PROMTOOL_VERSION}.linux-${PROMTOOL_ARCH}"

# Restore permissions as per https://hub.docker.com/_/golang
RUN chmod 777 -R /go

# Very permissive because we don't know what user the container will run as
RUN mkdir /home/beat && chmod 777 /home/beat
ENV HOME=/home/beat

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["bash", "/entrypoint.sh"]
