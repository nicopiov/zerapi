FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/zerapi .

FROM scratch

COPY --from=build /bin/zerapi /zerapi

WORKDIR /data

EXPOSE 8080

ENTRYPOINT ["/zerapi"]
CMD ["serve", "--host", "0.0.0.0", "/data/data.json"]
