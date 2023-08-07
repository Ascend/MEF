FROM ubuntu:22.04 as build
ARG UID
ARG GID
RUN groupadd MEFCenter -g ${GID}
RUN useradd -d /home/MEFCenter -u ${UID} -g ${GID} -m -s  /usr/sbin/nologin MEFCenter
WORKDIR /home/MEFCenter

COPY nginx .
COPY --chown=MEFCenter:MEFCenter ./lib/kmc-lib/* ./lib/

RUN chmod 750 /home/MEFCenter &&\
    chmod 500 -R /home/MEFCenter/* &&\
    chmod 700 -R /home/MEFCenter/conf &&\
    chmod 400 -R /home/MEFCenter/conf/* &&\
    chmod 400 -R /home/MEFCenter/html/* &&\
    chown MEFCenter:MEFCenter -R /home/MEFCenter/* &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> /home/MEFCenter/.bashrc

# MEFCenter is used as the default user of the container
USER MEFCenter