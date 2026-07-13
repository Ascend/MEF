# Security Hardening<a name="ZH-CN_TOPIC_0000001722295485"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:18:55.845Z pushedAt=2026-06-09T01:46:29.795Z -->

## Hardening Notes<a name="ZH-CN_TOPIC_0000001674256298"></a>

The security hardening measures listed in this document are basic hardening recommendations. Users should re-examine the network security hardening measures for the entire system based on their own business needs. Users should perform relevant configurations according to the security policies of their organization, including but not limited to software versions, password complexity requirements, security configurations (protocols, cipher suites, key lengths, etc.), permission configurations, and firewall settings. When necessary, refer to industry best hardening practices and recommendations from security experts.

Security hardening involves host hardening and containerized application hardening to prevent potential security risks and ensure the security of devices and containerized applications. Users should perform security hardening operations based on actual needs.

Software code or programs downloaded from external sources may pose risks, and users are responsible for ensuring the security of their functionality.

## Device Security Hardening<a name="ZH-CN_TOPIC_0000001674256322"></a>

- Disable remote login for the root user.

    Configuration method: Set the `PermitRootLogin` parameter to `no` in the `/etc/ssh/sshd_config` file.

- Use the Linux built-in ASLR (address space layout randomization) feature to enhance vulnerability attack protection.

    Configuration method: Write `2` to the `/proc/sys/kernel/randomize_va_space` file.

- Set the targetpw option in the sudo command to require the target user's password by default. This prevents all users from being able to escalate privileges to the root account and execute system commands without entering the root password after adding a sudo rule, which would lead to unauthorized command execution by ordinary users. This option is not added by default; it is recommended to add it.

    Run the `cat /etc/sudoers | grep -E "^[^#]*Defaults[[:space:\]\]\+targetpw` command to check if the `Defaults targetpw` or `Defaults rootpw` configuration entry exists. If it does not exist, add the `Defaults targetpw` or `Defaults rootpw` configuration entry under the `#Defaults specification` in the `/etc/sudoers` file.

- Prohibit ordinary users or groups from escalating privileges to the root user through all commands.

    Run the `cat /etc/sudoers` command to check if the `/etc/sudoers` file contains `(ALL) ALL and (ALL:ALL) ALL` entries for users or groups other than `root ALL=(ALL:ALL)  ALL` and `root ALL=(ALL) ALL`. If they exist, confirm whether they are needed based on the actual service scenario. If they are not needed, delete them, for example, `user ALL=(ALL) ALL`, `%admin ALL=(ALL) ALL`, or `%sudo ALL=(ALL:ALL) ALL`.

- To ensure the generation of secure random numbers, make sure the operating system supports the getrandom system call (supported by default in the operating system).

## OS Security Hardening<a name="ZH-CN_TOPIC_0000001674416034"></a>

Users must promptly apply security patches and use software versions approved by their organization in accordance with the organization's security policies.

### Configuring Firewalls<a name="ZH-CN_TOPIC_0000001674256330"></a>

After the operating system is installed, if a regular user is configured, you can prevent unauthorized access by adding the `ALWAYS_SET_PATH` field in the `/etc/login.defs` file and setting it to `yes`. In addition, to prevent regular users from inheriting environment variables through `su root` and thereby performing privilege escalation, you can set the configuration parameter `ALWAYS_SET_PATH` in the server configuration file `/etc/default/su` to `yes`.

The firewall needs to be disabled during the K8s installation and deployment phase. In a production environment, the secure practice is to configure the ports and network policies for communication between K8s components and KubeEdge CloudCore on the firewall. For specific configuration methods, refer to the official documentation.

### Setting umask<a name="ZH-CN_TOPIC_0000001674415962"></a>

It is recommended that users set the umask on the host (including the physical machine) and in containers to 027 or higher to improve security.

The following example shows the specific steps for setting umask to 077.

1. Log in to the server as the root user and edit the `/etc/profile` file.

    ```bash
    vim /etc/profile
    ```

2. Add **umask 077** to the end of the `/etc/profile` file, then save and exit.
3. Run the following command to apply the configuration.

    ```bash
    source /etc/profile
    ```

### Performing Security Hardening on Ownerless Files<a name="ZH-CN_TOPIC_0000001722375537"></a>

Due to differences between official Docker images and the operating system on the physical machine, users in the system may not have a one-to-one correspondence, causing files generated during the operation of the physical machine or container to become ownerless files.

Users can run the **find / -nouser -nogroup** command to search for ownerless files in the container or on the physical machine. Create corresponding users and user groups based on the file's uid and gid, or modify the uid of an existing user and the gid of a user group to match, thereby assigning an owner to the files and preventing ownerless files from posing security risks to the system.

### Defensing Against DoS Attacks<a name="ZH-CN_TOPIC_0000001674415914"></a>

You can prevent resources from being exhausted by malicious requests by adding a whitelist and adjusting the concurrency parameters of service components. The duration a client maintains a connection depends on the `keepAlive` related parameters set on the server. Set the TCP keepalive time, probe count, and probe interval appropriately based on your actual business needs.

To prevent SYN attacks, enable `tcp_syncookies` based on actual business requirements. Adjust the length of the SYN queue by setting `tcp_max_syn_backlog`, and redefine the number of SYN retries using `tcp_synack_retries` and   `tcp_syn_retries`.

## Container Security Hardening<a name="ZH-CN_TOPIC_0000001722295405"></a>

**Docker Container Runtime Security Hardening<a name="section529216501938"></a>**

To ensure the secure operation of containers, it is recommended that users configure the following hardening items based on their business needs. For specific operation methods, refer to the Official Documentation:

- Enable AppArmor Capability: When running a container, you can specify an AppArmor file. AppArmor provides security policies to protect the Linux system and applications. Before enabling the AppArmor capability, you need to enable the AppArmor feature in the Linux kernel.
- Enable SELinux Capability: When running a container, you can specify an SELinux configuration to improve security. Before enabling the SELinux capability, you need to use the `--selinux-enabled` configuration to take effect in the Docker daemon.
- Enable Live Restore Capability: You need to enable the `--live-restore` configuration to reduce the dependency on the Docker daemon.
- Set resource quotas for containers to prevent them from consuming excessive system resources, which could lead to resource exhaustion. System resources include but are not limited to CPU and memory.
- Avoid running untrusted applications in containers.
- Avoid listening on unnecessary ports in containers.
- Configure appropriate CPU priority for containers.
- Mount the container's root filesystem as read-only.
- Bind incoming container traffic to a specific host interface, and specify an IP address for the container's port mapping configuration.
- Limit the number of file handles and forked processes used by the running container.
- Enable authentication and encrypted transmission mechanisms for the service interfaces that the container service listens on externally, ensuring that service data is not stolen.
- Avoid running an SSH server inside containers.
- Avoid sharing namespaces, including: network namespace, UTS namespace, and user namespace.
- Avoid mounting docker.sock inside containers.
- Ensure that no users are added to the Docker user group.
- When using APIs related to creating containers or templates, or updating containers or templates, users must carefully configure parameters such as environment variables and ConfigMaps, and ensure the use of secure images. Avoid passing sensitive information through environment variables, ConfigMaps, etc., to prevent sensitive data leakage or privilege escalation risks due to improper configuration. Users are advised to perform thorough validation of data before use, based on their own business needs.

**Security Hardening on Containerized Application Logs<a name="section11787202519415"></a>**

If a containerized application has logs printed to standard output, containerized application log rotation may fail, leading to disk space exhaustion. Users are advised to configure the `max-size` and `max-file` parameters in the `log-opts` field of the `/etc/docker/daemon.json` Configuration File based on their business needs. This configuration will take effect for containerized applications created after the configuration is modified, following a Docker restart.

Parameter Description:

- max-size: The maximum size of a log file before it is rotated.
- max-file: The maximum number of log files retained during automatic log rotation.

**Host Security Hardening<a name="section5719132532217"></a>**

- Create a separate partition for containers. The default directory for Docker is `/var/lib/docker`. It is recommended to create a separate disk partition for Docker to prevent the disk capacity consumed by containers and that consumed by other host applications from affecting each other.
- The Docker host must be hardened. It is recommended to perform security hardening on the host running Docker containers and conduct regular vulnerability scans.
- Use the latest Docker version. It is recommended to update the Docker version in a timely manner to avoid known vulnerabilities in the Docker software.
- Enable auditing for the Docker daemon and critical files. Enabling auditing helps trace the root cause of attack events. However, this feature may cause some performance impact, and users need to decide whether to enable it based on their business requirements.

**Security Hardening on the Docker Daemon<a name="section1824414182320"></a>**

- Restrict inter-container network communication. By default, the Docker daemon allows network communication between containers, which can easily lead to information leakage. It is recommended to configure `-icc=false` in the Docker daemon.
- Prevent enabling the daemon's remote access interface. Do not use the Docker Remote API service, and strictly control the read and write permissions of the `docker.sock` file, adding only necessary users to the Docker user group. If business needs require enabling the Docker Remote API service, it is recommended to enable fine-grained access policy control for the daemon through `--authorization-plugin`.
- Limit the number of file handles and forked processes for containers. It is recommended to configure the `nofile` and `nproc` parameters in `--default-ulimit` within the Docker daemon to prevent fork bombs or exhaustion of file handle resources, which could lead to host compromise. The values for container resource limits must be evaluated based on business needs, as unreasonable limits may prevent containers from running. For example, `--default-ulimit nofile=64:64 --default-ulimit nproc=512:512` limits the number of file handles for a single process to 64 and the number of forked processes for a single UID user to 512.
- Disable the userland proxy. It is recommended to configure `--userland-proxy=false` in the Docker daemon to reduce the attack surface.
- Enable the user namespace. Once enabled, it provides permission isolation between container users and host users, for example, by configuring `--userns-remap=default` in the Docker daemon.
- Avoid using the aufs storage driver, as aufs is an unsupported driver.
- Configure the logging driver. Evaluate whether to enable the logging driver based on business needs.
- Ensure that the file permissions used by the Docker daemon are minimized. If the Docker configuration files are maliciously exploited, it may cause abnormal behavior in the Docker daemon. Key files and directories to focus on include but are not limited to:

    `/etc/docker/certs.d/`, `/etc/docker/daemon.json`, `/etc/default/docker`, `/usr/lib/systemd/system/docker.service`, `/etc/sysconfig/docker`, `/var/run/docker.sock`, `/etc/docker/`, and `/usr/lib/systemd/system/docker.socket`

## Kubernetes Security Hardening<a name="ZH-CN_TOPIC_0000001722375557"></a>

To ensure a secure operating environment, it is recommended that users control login permissions to the cluster master nodes based on business needs, enforce access control on the Kubernetes private key files and the authentication credentials stored in etcd within the environment; it is not recommended for users to directly operate the Kubernetes cluster from the backend.

Kubernetes requires the following Hardening measures:

- kube-proxy Hardening:
    - Add the `--nodeport-addresses` parameter to the `kube-proxy` startup parameters.
    - For an already installed K8s system, modify the kube-proxy configmap using the following command.

        ```bash
        kubectl edit cm kube-proxy -n kube-system
        ```

    - Manually modify the `nodePortAddresses` parameter in the configmap to the node IP in CIDR format.
    - Manually modify the `healthzBindAddress` parameter in the configmap to the node IP.
    - The configuration takes effect after restarting kube-proxy.

- kube-apiserver Hardening:
    - Add the startup parameter `--kubelet-certificate-authority` to configure the path to the kubelet CA certificate, which is used to verify the validity of the kubelet server certificate.
    - Modify the startup parameter `--profiling` and set its value to `false` to prevent users from dynamically changing the kube-apiserver log level.
    - Modify or add the startup parameter `--tls-cipher-suites` and set its value as follows to avoid risks caused by using insecure TLS Cipher Suites:

        ```bash
        --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        ```

    - Modify or add the startup parameter `--tls-min-version`, with a value example such as --tls-min-version=VersionTLS13, to configure the apiserver to use the TLS 1.3 security protocol for communication encryption.
    - Modify or add the startup parameter **--audit-policy-file** to configure the K8s audit policy. For details on the configuration, see the [Kubernetes Official Documentation](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/).

- kube-controller Hardening:
    - Add the sub-item **-serviceaccount-token** within the startup parameter **--controllers** to disable the default service account for the namespace, preventing the creation of unnecessary service accounts in the mef-user and mef-center namespaces when installing and running MEF.

- kubelet Hardening:
    - To prevent a single Pod from consuming an excessive number of processes, you can enable SupportPodPidsLimit and set <b>--pod-max-pids</b>. Add --feature-gates=SupportPodPidsLimit=true --pod-max-pids=\<max pid number\> to the KUBELET\_KUBEADM\_ARGS item in the kubelet configuration file. The modification takes effect after a restart. For details, see the [Kubernetes Official Documentation](https://kubernetes.io/docs/concepts/policy/pid-limiting/).
    - Configure the startup parameter `--address` or modify the `address` field in the startup configuration file, setting the value to the host IP.
    - Configure the startup parameter `--tls-min-version` or modify the `tlsMinVersion` field in the startup configuration file. An example value for the startup configuration file field is `tlsMinVersion: VersionTLS13`, used to configure kubelet to encrypt communications using the TLS 1.3 security protocol.
    - Modify or add the startup parameter `--tls-cipher-suites` and set its value as follows to avoid risks from using insecure TLS Cipher Suites:

        ```bash
        --tls-cipher-
        suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        ```

        > [!NOTE]
        >
        > K8s v1.19 and later versions support TLS v1.3 Cipher Suites. It is recommended to include TLS v1.3 Cipher Suites when using a higher version of K8s.

- If the OS kernel version used by the K8s cluster is 4.6 or later, manually enable AppArmor or SELinux after installing K8s.
- For the bandwidth limit of the inference service pod to take effect, you need to install the bandwidth plugin into the CNI bin directory (default: /opt/cni/bin), modify the CNI configuration file (default location: /etc/cni/net.d), and add bandwidth to the plugins.

    ```json
    ...
        {
          "type": "bandwidth",
          "capabilities": {"bandwidth": true}
        }
    ...
    ```

- For other security hardening content, refer to the relevant sections in the Kubernetes official documentation [Security](https://kubernetes.io/docs/concepts/security/), or consult other excellent hardening solutions in the industry.

## KubeEdge Security Hardening<a name="ZH-CN_TOPIC_0000001674415986"></a>

To ensure the secure operation of the environment, it is recommended that users control the login permissions for the cluster master nodes based on business needs, and enforce access permission control for the private key files required by the KubeEdge CloudCore component and the authentication credentials stored in etcd.

The CloudCore startup command supports specifying a configuration file path. By setting the IP parameter in the configuration file to a specific IP address, you can prevent listening on all zeros.

```bash
cloudcore --config Configuration File Path
```

For example, the following configuration for the CloudCore service uses a specific IP (where xx.xx.xx.xx represents a specific accessible IP):

```text
...
modules:
cloudHub:
advertiseAddress:
- xx.xx.xx.xx
enable: true
https:
address: xx.xx.xx.xx
enable: true
port: 10002
nodeLimit: 1000
...
websocket:
address: xx.xx.xx.xx
enable: true
port: 10000
...
router:
address: xx.xx.xx.xx
enable: false
port: 9443
restTimeout: 60
...
```
