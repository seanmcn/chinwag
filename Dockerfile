FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/whatsapp-analyse ./cmd/whatsapp-analyse

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/whatsapp-analyse /whatsapp-analyse
EXPOSE 8080
ENTRYPOINT ["/whatsapp-analyse", "--addr", "0.0.0.0:8080"]
