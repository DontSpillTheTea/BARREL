FROM golang:1.25-alpine
WORKDIR /app
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
COPY rules/ /rules/
EXPOSE 8080
CMD ["go", "run", "./cmd/server"]
