# TODO !! use an explicit version here, copy the cloud approach for updating docker image via CLI
FROM alpine:3.22.1
RUN apk add --no-cache docker docker-cli-compose
RUN addgroup -S app && adduser -S -D -H -s /sbin/nologin -G app app \
 && mkdir -p /opt/app/data \
 && chown -R app:app /opt/app
WORKDIR /opt/app
COPY --chown=app:app ./store  /opt/app/store
COPY --chown=app:app ./assets /opt/app/assets
RUN chmod 500 /opt/app/store
# TODO !! expose port 80 later
USER app
EXPOSE 8082
VOLUME ["/opt/app/data"]
ENTRYPOINT ["/opt/app/store"]
