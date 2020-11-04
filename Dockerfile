FROM alpine
LABEL maintainer="master@rebeta.cn"

ADD gaea_test /

CMD ["/gaea"]
