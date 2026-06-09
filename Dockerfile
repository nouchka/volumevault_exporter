FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/volumevault_exporter ./cmd/volumevault_exporter

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/volumevault_exporter /volumevault_exporter
EXPOSE 9780
ENTRYPOINT ["/volumevault_exporter"]
