FROM golang:latest AS build-stage
WORKDIR /go/src/app
COPY . .
RUN make build

FROM gcr.io/distroless/base-debian11 AS production-stage
WORKDIR /
COPY --from=build-stage /go/src/app/bin/copatcher /
USER nonroot:nonroot
CMD ["/copatcher"]
