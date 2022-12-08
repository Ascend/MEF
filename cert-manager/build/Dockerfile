FROM ubuntu:20.04 as build
RUN useradd -d /home/hwMindX -u 9000 -m -s  /usr/sbin/nologin hwMindX
COPY lib /home/cert_mgr_lib
WORKDIR /home/hwMindX
COPY cert-manager .

RUN chown hwMindX:hwMindX -Rh /home/hwMindX/* /home/cert_mgr_lib &&\
    chmod 750 /home/hwMindX &&\
    chmod 500 -R /home/hwMindX/* &&\
    chmod 700  /home/cert_mgr_lib &&\
    chmod 400 -R /home/cert_mgr_lib/* &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'export LD_LIBRARY_PATH=/home/cert_mgr_lib' >> /home/hwMindX/.bashrc && \
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

ENV LD_LIBRARY_PATH=/home/cert_mgr_lib
# hwMindX is used as the default user of the container
USER hwMindX