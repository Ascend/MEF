FROM scratch
FROM scratch
ARG ARCH
ARG LD_PATH
ARG BUILD_TIME
LABEL build.time=${BUILD_TIME}
ADD bin/pause /pause
COPY /lib/ld.so ${LD_PATH}
COPY /lib/libc.so /lib/libc.so.6
ENTRYPOINT ["/pause"]