FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini

# nobody 65534:65534
USER 65534:65534


COPY  smd /
COPY smd-loader /
COPY smd-init /

CMD [ "/smd" ]

ENTRYPOINT [ "/sbin/tini", "--" ]