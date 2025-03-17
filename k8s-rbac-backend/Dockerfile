FROM golang:1.19 as builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o k8s-rbac-backend

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/k8s-rbac-backend .
EXPOSE 8080
CMD ["./k8s-rbac-backend"]