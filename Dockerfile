FROM centos:7

COPY liveness /liveness

RUN chmod +x /liveness

ENTRYPOINT ["/liveness"]
