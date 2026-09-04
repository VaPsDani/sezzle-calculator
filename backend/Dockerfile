# ---------- build ----------
FROM golang:1.23-alpine AS build

WORKDIR /src

# Copied on their own so that the dependency layer is only rebuilt when the
# module files change, not on every source edit.
COPY go.mod ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produces a static binary with no libc dependency, which is what
# lets the final stage be a bare Alpine with nothing installed in it.
# -trimpath removes local build paths, -s -w drop the debug symbols.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---------- runtime ----------
FROM alpine:3.20

# A container that runs as root runs as root on the host kernel too if it ever
# escapes its namespace, so the process gets an unprivileged user of its own.
RUN adduser --disabled-password --gecos "" --uid 10001 app

COPY --from=build /out/server /usr/local/bin/server

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]
