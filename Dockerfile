FROM golang:1.13.1 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM gcr.io/distroless/static:nonroot 
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
