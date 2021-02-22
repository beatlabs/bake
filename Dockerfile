FROM golang:1.15

RUN apt-get -y update && \
    apt-get install -y \
    apt-transport-https \	
    ca-certificates \
    gnupg-agent \
    software-properties-common	

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \	
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable" && \	
    apt-get -y update && \	
    apt-get install -y docker-ce

# CGO is required by some modules like https://github.com/uber/h3-go
ENV CGO_ENABLED=1

# Required to access private modules
ENV GOPRIVATE=github.com/taxibeat/*

# expect a build-time variable
ARG GH_TOKEN

# Access to Beat private repos
RUN git config --global url."https://$GH_TOKEN@github.com/".insteadOf "https://github.com/"

# Download and install skim - proto schema registry tool
RUN go get github.com/taxibeat/skim/cmd/skim

# Remove token from config after accessing private repos.
RUN git config --global --remove-section url."https://$GH_TOKEN@github.com/"
# Download and install mage file into bin path
RUN wget -qc https://github.com/magefile/mage/releases/download/v1.11.0/mage_1.11.0_Linux-64bit.tar.gz -O - | tar -xz -C /usr/bin mage

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

# Very permissive because we don't know what user the container will run as
RUN mkdir /home/beat && chmod 777 /home/beat
ENV HOME /home/beat

COPY bake.sh /home/beat/bake-default.sh

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["bash", "/entrypoint.sh"]
