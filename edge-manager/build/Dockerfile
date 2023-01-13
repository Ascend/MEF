FROM ubuntu:22.04 as build
RUN useradd -d /home/MEFCenter -u 8000 -m -s  /usr/sbin/nologin MEFCenter
WORKDIR /home/MEFCenter
COPY edge-manager .
COPY --chown=MEFCenter:MEFCenter ../../lib/kmc_lib/* ./lib/

RUN chown MEFCenter:MEFCenter -R /home/MEFCenter/* &&\
    chmod 750 /home/MEFCenter &&\
    chmod 500 -R /home/MEFCenter/* &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> /home/MEFCenter/.bashrc

ENV LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:/home/MEFCenter/lib
# MEFCenter is used as the default user of the container
USER MEFCenter