FROM alpine:3.19@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd49f277b3b5c1d4b
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
WORKDIR /home/appuser
COPY --chown=appuser:appgroup app /home/appuser/app
ENTRYPOINT ["/home/appuser/app"]
