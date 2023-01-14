FROM ubuntu:22.04 as build
RUN useradd -d /home/MEFCenter -u 8000 -m -s  /usr/sbin/nologin MEFCenter
COPY --chown=MEFCenter:MEFCenter ./lib /home/cert_mgr_lib
WORKDIR /home/MEFCenter
COPY cert-manager .

RUN chown MEFCenter:MEFCenter -Rh /home/MEFCenter/* /home/cert_mgr_lib &&\
    chmod 750 /home/MEFCenter &&\
    chmod 500 -R /home/MEFCenter/* &&\
    chmod 700  /home/cert_mgr_lib &&\
    chmod 400 -R /home/cert_mgr_lib/* &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'export LD_LIBRARY_PATH=/home/cert_mgr_lib' >> /home/MEFCenter/.bashrc && \
    echo 'source /etc/profile' >> /home/MEFCenter/.bashrc

ENV LD_LIBRARY_PATH=/home/cert_mgr_lib
# MEFCenter is used as the default user of the container
USER MEFCenter