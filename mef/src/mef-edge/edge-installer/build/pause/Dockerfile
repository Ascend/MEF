FROM scratch
FROM scratch
ARG ARCH
ARG LD_PATH
ADD bin/pause /pause
COPY /lib/ld.so ${LD_PATH}
COPY /lib/libc.so /lib/libc.so.6
ENTRYPOINT ["/pause"]