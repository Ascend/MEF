FROM ubuntu:18.04 as build
RUN useradd -d /home/hwMindX -u 9000 -m -s  /usr/sbin/nologin hwMindX
WORKDIR /home/hwMindX
COPY edge-manager .
COPY --chown=hwMindX:hwMindX ./lib/* ./lib/

RUN chown hwMindX:hwMindX -R /home/hwMindX/* &&\
    chmod 750 /home/hwMindX &&\
    chmod 500 -R /home/hwMindX/* &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

ENV LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:/home/hwMindX/lib
# hwMindX is used as the default user of the container
USER hwMindX