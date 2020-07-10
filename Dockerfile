FROM golang:1.14

RUN apt-get -y update

# Install docker binary (required on mac/windows)
RUN apt-get install -y \
    apt-transport-https \
    ca-certificates \
    gnupg-agent \
    software-properties-common
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable" && \
    apt-get -y update && \
    apt-get install -y docker-ce

# Download and install mage file into bin path
RUN wget -c https://github.com/magefile/mage/releases/download/v1.9.0/mage_1.9.0_Linux-64bit.tar.gz -O - | tar -xz -C /usr/bin mage

# Download and install hadolint into bin path
RUN wget -O /usr/bin/hadolint https://github.com/hadolint/hadolint/releases/download/v1.17.6/hadolint-Linux-x86_64 && chmod +x /usr/bin/hadolint

# Download and install swag into bin path
RUN wget -c https://github.com/swaggo/swag/releases/download/v1.6.6/swag_1.6.6_Linux_x86_64.tar.gz -O - | tar -xz -C /usr/bin swag

# Download and install mark into bin path
RUN wget -c https://github.com/mantzas/mark/releases/download/v0.9.0/mark-linux-x64.tar.gz -O - | tar -xz -C /usr/bin mark

# Download and install golangci-lint into go bin path
RUN wget -c https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -O - | /bin/sh -s -- -b $(go env GOPATH)/bin v1.28.2

# Very permissive because we don't know what user the container will run as
RUN mkdir /home/beat && chmod 777 /home/beat
ENV HOME /home/beat

COPY bake.sh /home/beat/bake-default.sh

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["bash", "/entrypoint.sh"]
