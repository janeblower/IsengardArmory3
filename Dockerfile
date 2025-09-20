### BUILD ISENGARDARMORY MULTIARCH START ###
FROM --platform=$BUILDPLATFORM golang:1.25.0-alpine AS builder
COPY . /opt/src
WORKDIR /opt/src
ARG TARGETARCH
# Step for multiarch build with docker buildx
ENV GOARCH=$TARGETARCH
# Build isengardarmory
RUN apk add --update g++ \
&& go mod tidy \
&& go clean -i -r -cache \
&& go build -ldflags '-w -s' --o isengard-armory .
### BUILD ISENGARDARMORY MULTIARCH END ###


### BUILD MAIN IMAGE START ###
FROM alpine
WORKDIR /app
COPY --from=builder /opt/src/isengard-armory /app/isengard-armory
COPY ./static /app/static
COPY ./templates /app/templates
CMD ["./isengard-armory"]
### BUILD MAIN IMAGE end ###