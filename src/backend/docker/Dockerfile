FROM alpine:latest
RUN addgroup -S app && adduser -S -D -H -s /sbin/nologin -G app app \
 && mkdir -p /opt/app/data/logs \
 && chown -R app:app /opt/app
WORKDIR /opt/app
COPY --chown=app:app ./store  /opt/app/store
COPY --chown=app:app ./dist   /opt/app/dist
COPY --chown=app:app ./assets /opt/app/assets
RUN chmod 500 /opt/app/store
USER app
EXPOSE 8082
VOLUME ["/opt/app/data"]
ENTRYPOINT ["/opt/app/store"]
