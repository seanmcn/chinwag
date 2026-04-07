FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
ENV GOWORK=off
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/chinwag ./cmd/chinwag

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/chinwag /chinwag
ENTRYPOINT ["/chinwag"]
