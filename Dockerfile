FROM wacky6/cloud-dev

LABEL maintainer="Jiewei Qian <qjw@wacky.one>"

ENV WORKDIR="/cf-ddns-go/"

ARG GO_VER="1.15.2"
ARG VSC_EXTENSIONS="golang.go"

WORKDIR ${WORKDIR}

RUN    curl -fsSL https://golang.org/dl/go${GO_VER}.linux-${ARCH}.tar.gz | tar xz -C /usr/local/ \
    && echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile \
    && echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/zsh/zshenv \
    && /dev-env/install-extensions.sh ${VSC_EXTENSIONS}
