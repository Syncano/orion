FROM alpine:3.11

ARG EMAIL=devops@syncano.com
ENV ACME_VERSION=2.8.3 \
    LE_WORKING_DIR=/acme/home \
    LE_CONFIG_HOME=/acme/config \
    CERT_HOME=/acme/certs

RUN set -ex \
    && addgroup -S orion \
    && adduser -S -G orion orion \
    && apk --no-cache add \
        # env zip processing
        squashfs-tools \
        unzip \
        # openssl support
        ca-certificates \
        openssl \
        curl \
        # pdf rendering
        wkhtmltopdf \
        ttf-freefont \
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

COPY orion /usr/local/bin/orion

WORKDIR /home/orion
USER orion
