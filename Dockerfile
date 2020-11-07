FROM alpine
LABEL maintainer="master@rebeta.cn"

# 修正时区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 添加二进制程序到容器内
ADD gaea /

# 配置 phonedata 库的手机号归属地代码数据
ENV PHONE_DATA_DIR=/
ADD phone.dat /

# 容器启动命令
CMD ["/gaea"]
