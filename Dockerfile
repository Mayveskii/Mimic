FROM golang:1.22-bookworm AS builder

RUN apt-get update && apt-get install -y gcc make

WORKDIR /build

COPY go.mod go.sum ./
COPY core/ core/
COPY internal/ internal/
COPY cmd/ cmd/
COPY Makefile .

RUN make build

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/bin/mimic /usr/local/bin/mimic

ENTRYPOINT ["mimic"]
CMD ["serve"]
