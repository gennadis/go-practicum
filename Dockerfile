### Build
FROM golang:1.21-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download -x

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/shortener/main.go

### Final
FROM scratch

COPY --from=build /bin/server /bin/

CMD [ "./bin/server" ]
