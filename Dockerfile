FROM golang:1.19
WORKDIR /app
COPY . ./
RUN make build
CMD ["/app/filsigner", "run"]
