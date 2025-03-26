# Copyright 2023 Adevinta

FROM golang:1.24-alpine AS builder

ARG TARGETOS TARGETARCH

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd cmd/vulcan-tracker && GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags musl . && cd -

FROM alpine:3.21

WORKDIR /flyway

RUN apk add --no-cache --update openjdk17-jre-headless bash gettext libc6-compat

ARG FLYWAY_VERSION=10.10.0

RUN wget -q https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && tar -xzf flyway-commandline-${FLYWAY_VERSION}.tar.gz --strip 1 \
    && rm flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && find ./drivers/ -type f | grep -Ev '(postgres|jackson)' | xargs rm \
    && chown -R root:root . \
    && ln -s /flyway/flyway /bin/flyway

WORKDIR /app

COPY db/sql /app/sql/

RUN mkdir -p /app/output

COPY --from=builder /app/cmd/vulcan-tracker/vulcan-tracker .

COPY config.toml .
COPY run.sh .

CMD [ "./run.sh" ]
