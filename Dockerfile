FROM golang:1.26-bookworm@sha256:f5b1ea492ad3e465d57c79dcea9393c2f0710ca2dce326078691851fd4307737 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/golden-path-service \
    ./cmd/golden-path-service

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6

LABEL org.opencontainers.image.title="Vaipex Golden Path Service" \
      org.opencontainers.image.description="Community reference service for the Vaipex Golden Path" \
      org.opencontainers.image.source="https://github.com/vaipexlabs/platform-lab-01-golden-path"

COPY --from=build --chown=65532:65532 /out/golden-path-service /golden-path-service

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/golden-path-service"]
