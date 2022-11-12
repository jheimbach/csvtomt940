FROM alpine:3.14

WORKDIR /app

ENTRYPOINT ["/usr/bin/csvtomt940"]

COPY csvtomt940 /usr/bin/csvtomt940
RUN chmod +x /usr/bin/csvtomt940