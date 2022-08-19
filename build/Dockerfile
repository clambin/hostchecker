FROM alpine

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

COPY bld/hostchecker-$TARGETOS-$TARGETARCH /app/hostchecker

EXPOSE 8080
ENTRYPOINT ["/app/hostchecker"]
CMD []
