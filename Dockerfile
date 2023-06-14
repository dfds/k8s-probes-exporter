FROM golang:1.20-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY cmds /app/cmds

RUN go build -o /app/app /app/cmds/server.go

FROM golang:1.20-alpine

COPY --from=build /app/app /app/app

CMD [ "/app/app" ]