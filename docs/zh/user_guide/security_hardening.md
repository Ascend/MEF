# 安全加固<a name="ZH-CN_TOPIC_0000001722295485"></a>

## 加固须知<a name="ZH-CN_TOPIC_0000001674256298"></a>

本文中列出的安全加固措施为基本的加固建议项。用户应根据自身业务，重新审视整个系统的网络安全加固措施。用户应按照所在组织的安全策略进行相关配置，包括并不局限于软件版本、口令复杂度要求、安全配置（协议、加密套件、密钥长度等），权限配置、防火墙设置等。必要时可参考业界优秀加固方案和安全专家的建议。

安全加固涉及主机加固和容器应用加固，防止可能出现的安全隐患，用于保障设备和容器应用的安全，请用户根据实际需要进行安全加固操作。

外部下载的软件代码或程序可能存在风险，功能的安全性需由用户保证。

## 设备安全加固<a name="ZH-CN_TOPIC_0000001674256322"></a>

- 禁止root用户远程登录。

    设置方法：在“/etc/ssh/sshd\_config“文件中将"**PermitRootLogin**"参数设置为“no“。

- 使用Linux自带的ASLR（address space layout randomization）功能，增强漏洞攻击防护能力。

    设置方法：在“/proc/sys/kernel/randomize\_va\_space“文件中写入“2“。

- 将sudo命令中targetpw选项设置为默认要求输入目标用户的密码，防止增加sudo规则后，所有用户不需要输入root密码，就可以提权root账号执行系统命令，导致普通用户越权执行命令。该选项默认不添加，建议添加该选项。

    执行**cat /etc/sudoers \| grep -E "\^[\^#\]\*Defaults[[:space:\]\]\+targetpw"** 命令检查是否存在“Defaults targetpw“或“Defaults rootpw“配置项。如果不存在，请在“/etc/sudoers“文件的“\#Defaults specification”下添加“Defaults targetpw“或者“Defaults rootpw“配置项。

- 禁止普通用户或组通过所有命令提权到root用户。

    执行**cat /etc/sudoers**命令检查“/etc/sudoers“文件中是否存在“root ALL=(ALL:ALL)  ALL“和“root ALL=(ALL) ALL“之外的其他用户或组的“(ALL) ALL和(ALL:ALL) ALL“。如果存在，请根据实际业务场景确认是否需要，如果确认不需要，请删除，例如“user ALL=(ALL) ALL“、“%admin ALL=(ALL) ALL“或“%sudo ALL=(ALL:ALL) ALL“。

- 为保证能够生成安全的随机数，请确保操作系统支持getrandom系统调用（默认操作系统已支持）。

## 操作系统安全加固<a name="ZH-CN_TOPIC_0000001674416034"></a>

用户需按照所在组织的安全策略，及时更新安全补丁，并使用所在组织认可的软件版本。

### 防火墙配置<a name="ZH-CN_TOPIC_0000001674256330"></a>

操作系统安装后，若配置普通用户，可以通过在“/etc/login.defs“文件中新增“ALWAYS\_SET\_PATH“字段并设置为“yes“，防止越权操作。此外，为了防止普通用户通过**su root**继承环境变量从而提权，可以将服务器配置文件“/etc/default/su“中的配置参数“ALWAYS\_SET\_PATH“设置为“yes“。

K8s的安装部署阶段需要关闭防火墙，生产环境中安全的做法是在防火墙上配置K8s各组件和KubeEdge CloudCore通信的端口以及网络策略，具体配置方法请自行查看官方文档。

### 设置umask<a name="ZH-CN_TOPIC_0000001674415962"></a>

建议用户将主机（包括宿主机）和容器中的umask设置为027及其以上，提高安全性。

以设置umask为077为例，具体操作如下所示。

1. 以root用户登录服务器，编辑“/etc/profile“文件。

    ```bash
    vim /etc/profile
    ```

2. 在“/etc/profile“文件末尾加上**umask 077**，保存并退出。
3. 执行如下命令使配置生效。

    ```bash
    source /etc/profile
    ```

### 无属主文件安全加固<a name="ZH-CN_TOPIC_0000001722375537"></a>

因为官方Docker镜像与物理机上的操作系统存在差异，系统中的用户可能不能一一对应，导致物理机或容器运行过程中产生的文件变成无属主文件。

用户可以执行**find / -nouser -nogroup**命令，查找容器内或物理机上的无属主文件。根据文件的uid和gid创建相应的用户和用户组，或者修改已有用户的uid和用户组的gid来适配，赋予文件属主，避免无属主文件给系统带来安全隐患。

### 防DoS攻击<a name="ZH-CN_TOPIC_0000001674415914"></a>

可以通过添加白名单和调整服务组件并发参数大小等方式，防止资源被恶意请求占满。客户端维持连接的时间取决于服务器所设置的“**keepAlive**“相关参数，请根据实际业务合理设置TCP保活时间、探测次数和探测间隔。

为防止SYN攻击，根据实际业务需求，打开“**tcp\_syncookies**“，通过设置“**tcp\_max\_syn\_backlog**“重新调整SYN队列的长度，通过“**tcp\_synack\_retries**“和“**tcp\_syn\_retries**“重新定义SYN的重试次数。

## 容器安全加固<a name="ZH-CN_TOPIC_0000001722295405"></a>

**Docker容器运行安全加固<a name="section529216501938"></a>**

为保证容器安全运行，建议用户根据业务，配置如下加固项，具体操作方法请参考官方说明：

- 启用apparmor能力：运行容器时可指定apparmor文件，apparmor能够提供安全策略，保护Linux系统和应用程序。启用apparmor能力前需要先开启Linux内核的apparmor功能。
- 启用SELinux能力：运行容器时可指定SELinux配置，提高安全性。启用SELinux能力前需要使用“**--selinux-enabled**“配置在Docker守护进程中生效。
- 启用现场还原能力：需要启用“**--live-restore**“配置减少对Docker守护进程的依赖。
- 为容器设置系统资源的配额，避免容器占用过多的系统资源，导致资源耗尽。系统资源包括但不限于CPU、内存。
- 避免在容器中运行不可信的应用。
- 避免在容器中侦听不必要的端口。
- 为容器配置适当的CPU优先级。
- 将容器的根文件系统挂载为只读模式。
- 将传入的容器流量绑定到特定的主机接口，为容器的端口映射配置指定IP地址。
- 限制容器运行使用的文件句柄和fork进程数量。
- 容器服务对外侦听的业务接口启用认证和加密传输机制，保证业务数据不被窃取。
- 避免在容器中运行SSH服务端。
- 避免共享命名空间，包括：网络命名空间、UTS命名空间、user命名空间。
- 避免在容器中挂载docker.sock。
- 确保没有任何用户加入Docker用户组。
- 用户在使用创建容器或模板、更新容器或模板相关接口时，需谨慎配置环境变量、ConfigMap等参数，并确保使用安全镜像。避免通过环境变量、ConfigMap等传递敏感信息，以防止敏感数据泄露，或因配置不当而存在提权风险。建议用户结合自身业务，在使用数据前做好充分校验。

**容器应用日志安全加固<a name="section11787202519415"></a>**

如果容器应用中存在日志打印到标准输出，可能会出现容器应用日志绕接失败，导致磁盘空间耗尽。建议用户根据业务，配置“/etc/docker/daemon.json“配置文件中“**log-opts**“字段的“**max-size**“和“**max-file**“参数。该配置会在重启Docker后，对修改配置之后创建的容器应用生效。

参数说明：

- max-size：日志自动转储的最大转储文件大小。
- max-file：日志自动转储的最大转储文件数量。

**主机安全加固<a name="section5719132532217"></a>**

- 为容器创建单独分区。Docker的默认目录是“/var/lib/docker“，建议为Docker创建单独的磁盘分区，避免容器占用的磁盘容量和主机其他应用使用的磁盘容量相互影响。
- Docker主机必须进行加固。建议对运行Docker容器的主机进行安全加固，并定期进行漏洞扫描。
- 使用最新的Docker版本。建议及时更新Docker的版本，规避Docker软件的已知漏洞。
- 开启Docker守护进程和关键文件的审计功能。开启审计功能能够追溯攻击事件的根源，此功能开启后，会造成一定的性能影响，需要用户根据业务决定是否开启。

**Docker守护进程安全加固<a name="section1824414182320"></a>**

- 限制容器间的相互网络访问。Docker守护进程默认允许容器间相互网络通信，容易造成信息泄露，建议将-icc=false配置在Docker守护进程中。
- 防止开启守护进程的远程访问接口。不要使用Docker Remote API服务，且需要严格控制docker.sock文件的读写权限，只将必要的用户配置到Docker用户组中。若业务需要开启Docker Remote API服务，建议通过--authorization-plugin开启守护进程的细粒度访问策略控制。
- 限制容器的文件句柄数、fork进程数。建议将--default-ulimit 中 nofile和nproc参数配置在Docker守护进程中，避免fork炸弹或者耗尽文件句柄资源的情况，导致主机受到攻击。容器资源限制的数值需要根据业务评估，不合理的限额会导致容器无法运行。例如--default-ulimit nofile=64:64 --default-ulimit nproc=512:512，限制单个进程的文件句柄数为64，单个UID用户的fork进程数为512。
- 禁用用户空间代理。建议将--userland-proxy=false配置在Docker守护进程中，减小攻击面。
- 启用user namespace命名空间。启用后，会提供容器用户与主机用户的权限隔离，例如使用--userns-remap=default配置在Docker守护进程中。
- 避免使用aufs存储驱动，aufs是不受支持的驱动程序。
- 配置日志驱动。根据业务评估是否需要开启日志驱动。
- 确保Docker守护进程使用到的文件权限最小化。若Docker的配置文件被恶意利用，可能导致Docker守护进程出现异常行为。需要重点关注的文件和目录包括但不限于：

    /etc/docker/certs.d/，/etc/docker/daemon.json，/etc/default/docker，/usr/lib/systemd/system/docker.service，/etc/sysconfig/docker，/var/run/docker.sock，/etc/docker/，/usr/lib/systemd/system/docker.socket

## Kubernetes安全加固<a name="ZH-CN_TOPIC_0000001722375557"></a>

为保证环境安全运行，建议用户根据业务控制集群master节点登录权限，对环境中kubernetes的私钥文件及etcd中存储的认证凭据做好访问权限控制；不建议用户在后台直接操作kubernetes集群。

Kubernetes需要进行如下加固：

- kube-proxy加强：
    - 在"**kube-proxy**"启动参数中加入"**--nodeport-addresses**"参数。
    - 针对已经安装的K8s系统，通过如下命令修改kube-proxy的configmap。

        ```bash
        kubectl edit cm kube-proxy -n kube-system
        ```

    - 手动修改configmap中的"**nodePortAddresses**"参数为CIDR格式的节点IP。
    - 手动修改configmap中的"**healthzBindAddress**"参数为节点IP。
    - 配置再重启kube-proxy后生效。

- kube-apiserver加强：
    - 添加启动参数“--kubelet-certificate-authority“参数，配置kubelet CA证书路径，用于验证kubelet服务端证书有效性。
    - 修改启动参数“**--profiling**“的值设置为“false“，防止用户动态修改kube-apiserver日志级别。
    - 修改或增加启动参数“**--tls-cipher-suites**“，设置它的值如下，避免使用不安全的TLS加密套件带来风险：

        ```bash
        --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        ```

    - 修改或增加启动参数“**--tls-min-version**“，取值示例如--tls-min-version=VersionTLS13，用于配置apiserver时使用TLS1.3安全协议对通信进行加密。
    - 修改或增加启动参数“**--audit-policy-file**“，配置K8s的审计策略，具体配置可参考[Kubernetes官方文档](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/)。

- kube-controller加强：
    - 在启动参数“--controllers“内添加子项“-serviceaccount-token“，用于禁用命名空间默认服务账号，防止在安装运行MEF时，在mef-user和mef-center命名空间产生不需要使用的服务账号。

- kubelet加强：
    - 为防止单Pod占用过多进程数，可以开启SupportPodPidsLimit，并设置<b>--pod-max-pids</b>。在kubelet配置文件的KUBELET\_KUBEADM\_ARGS项增加--feature-gates=SupportPodPidsLimit=true --pod-max-pids=<max pid number\>。配置修改后，重启生效。具体可参考[Kubernetes官方文档](https://kubernetes.io/docs/concepts/policy/pid-limiting/)。
    - 配置启动参数--address或者修改启动配置文件中的address字段，设置值为主机IP。
    - 配置启动参数“**--tls-min-version**“或者修改启动配置文件中的“**tlsMinVersion**“字段，启动配置文件字段取值示例如tlsMinVersion: VersionTLS13，用于配置kubelet时使用TLS1.3安全协议对通信进行加密。
    - 修改或增加启动参数“**--tls-cipher-suites**“，设置它的值如下，避免使用不安全的TLS加密套件带来风险：

        ```bash
        --tls-cipher-
        suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        ```

        > [!NOTE] 说明  
        > K8s v1.19及以上版本支持TLS v1.3的加密套件，建议使用高版本的K8s时，加上TLS v1.3的加密套件。

- 若K8s集群使用的OS kernel内核版本大于等于4.6，安装完K8s后手动开启AppArmor或者SELinux。
- 为使推理服务pod的带宽限制生效，需要安装bandwidth插件到CNI bin目录中（默认为/opt/cni/bin），并修改CNI配置文件（默认在/etc/cni/net.d下），在plugins中加入bandwidth。

    ```json
    ...
        {
          "type": "bandwidth",
          "capabilities": {"bandwidth": true}
        }
    ...
    ```

- 其余安全加固内容可参考Kubernetes官方文档[Security](https://kubernetes.io/docs/concepts/security/)相关内容，也可以参考业界其他优秀加固方案。

## KubeEdge安全加固<a name="ZH-CN_TOPIC_0000001674415986"></a>

为保证环境安全运行，建议用户根据业务控制集群master节点登录权限，对环境中KubeEdge的CloudCore组件运行所需私钥文件及etcd中存储的认证凭据做好访问权限控制。

CloudCore的启动命令支持指定配置文件路径，通过设置配置文件IP参数为具体的IP，防止全0侦听。

```bash
cloudcore --config 配置文件路径
```

例如下面针对CloudCore服务的配置，使用了具体的IP（其中xx.xx.xx.xx表示具体的可访问的IP）：

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
