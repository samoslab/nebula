FROM karalabe/xgo-latest
MAINTAINER sliver <lijiangt@gmail.com>
RUN \
  sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
  apt-get clean && \
  apt-get update && \
  apt-get install -y flex bison
