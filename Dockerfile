FROM golang:1.15
ARG EMAIL=devops@syncano.com
ENV ACME_VERSION=2.8.3 \
    LE_WORKING_DIR=/acme/home \
    LE_CONFIG_HOME=/acme/config \
    CERT_HOME=/acme/certs \
    GOPROXY=https://proxy.golang.org
WORKDIR /opt/build

RUN set -ex \
    && apt-get update && apt-get install --no-install-recommends -y \
        # env zip processing
        squashfs-tools \
        unzip \
        # pdf rendering
        wkhtmltopdf \
        fonts-freefont-ttf \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    \
    # Install acme.sh
    && wget https://github.com/Neilpang/acme.sh/archive/${ACME_VERSION}.zip \
    && unzip ${ACME_VERSION}.zip \
    && cd acme.sh-${ACME_VERSION} \
    && mkdir -p ${LE_WORKING_DIR} ${LE_CONFIG_HOME} ${CERT_HOME} \
    && ./acme.sh --install --nocron --home ${LE_WORKING_DIR} --config-home ${LE_CONFIG_HOME} --cert-home ${CERT_HOME} \
        --accountemail "${EMAIL}" --accountkey "/acme/config/account.key" \
    && ln -s ${LE_WORKING_DIR}/acme.sh /usr/bin/acme.sh \
    && cd .. \
    && rm -rf ${ACME_VERSION}.zip acme.sh-${ACME_VERSION}

COPY go.mod go.sum ./
RUN go mod download
