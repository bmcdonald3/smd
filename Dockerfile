FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini



COPY  smd /
COPY smd-loader /
COPY smd-init /
RUN mkdir /persistent_migrations
COPY migrations/* /persistent_migrations/

EXPOSE 27779

# nobody 65534:65534
USER 65534:65534

CMD [ "/smd"]

ENTRYPOINT [ "/sbin/tini", "--" ]
