FROM alpine
LABEL maintainer="master@rebeta.cn"

ADD gaea /

CMD ["/gaea"]
