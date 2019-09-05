FROM golang AS builder
WORKDIR /screepsplus
COPY . .
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    packr2 build -o screepsplus
RUN echo "screepsplus:x:1000:1000::/:/sbin/nologin" > passwd \
  && echo "screepsplus:x:1000:" > group

FROM scratch
COPY --from=builder /screepsplus/passwd /screepsplus/group /etc/
COPY --from=builder /screepsplus/screepsplus /
USER screepsplus
ENTRYPOINT ["/screepsplus"]
