#
# Copyright (c) 2018
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BUILDER_BASE=golang:1.13-alpine
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make && apk add bash git
COPY . .
RUN make cmd/support-logging/support-logging

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Cavium'

COPY --from=builder /bin/bash /bin/bash
COPY --from=builder /edgex-go/cmd/support-logging/Attribution.txt /
COPY --from=builder /edgex-go/cmd/support-logging/support-logging /
COPY --from=builder /edgex-go/cmd/support-logging/res/configuration.toml /res/configuration.toml
ENTRYPOINT ["/support-logging","-cp=consul.http://edgex-core-consul:8500","--registry","--confdir=/res"]
