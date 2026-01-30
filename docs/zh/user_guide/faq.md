# FAQ<a name="ZH-CN_TOPIC_0000001674256326"></a>

## 通过导入基础镜像的方式安装MEF Center失败<a name="ZH-CN_TOPIC_0000001749571521"></a>

**问题现象描述<a name="section13754951112814"></a>**

用户通过**docker load**命令导入本地的Ubuntu:22.04基础镜像的方式安装MEF Center失败。

**原因分析<a name="section77551626153617"></a>**

当MEF Center安装环境的Docker版本为23.0及以上版本时，Docker会默认启用BuildKit作为镜像构建工具对镜像进行构建，并通过镜像仓库重新获取依赖的基础镜像。MEF Center构建依赖的基础镜像Ubuntu:22.04时，如果环境配置的镜像仓库或Docker公共镜像仓库不可用，且通过离线导入的方式获取基础镜像时，会导致MEF Center组件镜像构建失败，从而导致安装MEF Center失败。

**解决措施<a name="section157021112307"></a>**

安装MEF Center前，执行以下命令，以设置环境变量方式关闭Docker的BuildKit功能。

```bash
export DOCKER_BUILDKIT=0
```

## MEF Edge重复启动或停止告警<a id="mef-edge重复启动或停止告警"></a>

**问题描述<a name="section281692625610"></a>**

MEF Edge软件已经启动或停止后，重复执行启动或停止的命令会返回告警提示。回显示例如下。

- 重复启动

    ```text
    warning: component [edge-om] is already started!
    warning: component [edge-main] is already started!
    warning: component [edgecore] is already started!
    warning: component [device-plugin] is already started!
    ```

- 重复停止

    ```text
    warning: component [edge-om] is already stopped!
    warning: component [edge-main] is already stopped!
    warning: component [edgecore] is already stopped!
    warning: component [device-plugin] is already stopped!
    ```

**解决措施<a name="section3191818161414"></a>**

非重复执行启动或停止情况下则不会出现该提示。

## MEF Center升级环境被强制终止后恢复环境操作<a name="ZH-CN_TOPIC_0000001722295465"></a>

**问题描述<a name="section195476587165"></a>**

升级MEF Center软件时，因设备上下电或其他异常场景导致升级强制终止后需要清除镜像并清空节点标签恢复环境。

**解决措施<a name="section5415101941714"></a>**

1. 登录设备环境。
2. 执行以下命令，恢复升级环境。

    ```bash
    run.sh start 
    ```

## MEF Edge升级环境被强制终止后恢复环境操作<a name="ZH-CN_TOPIC_0000001722295469"></a>

**问题描述<a name="section2448162514186"></a>**

升级MEF Edge软件时，因设备上下电或其他异常场景导致升级强制终止后需要恢复环境。

**解决措施<a name="section172121345201815"></a>**

1. 登录设备环境。
2. 如果存在残留的文件“/home/data/mefedge/unpack/edge\_installer“，删除该文件。
3. 检查当前MEF Edge版本，如未升级成功，需重新升级。
    - 执行以下命令，进入安装目录

        ```bash
        cd MEFEdge安装路径/MEFEdge/software
        ```

    - 执行以下命令，查看version.xml文件中MEF Edge的版本。

        ```bash
        cat version.xml
        ```

4. 若version.xml中的MEF Edge版本为目标版本，执行以下命令，恢复升级环境。

    ```bash
    run.sh start
    ```

## 节点压力驱逐机制导致MEF Center运行异常<a id="ZH-CN_TOPIC_0000001780849121"></a>

**问题现象描述<a name="section13754951112814"></a>**

K8s因部署和运行MEF Center时节点内存、磁盘、PID等资源不足，触发节点压力驱逐机制，导致MEF Center服务无法正常运行。

**原因分析<a name="section77551626153617"></a>**

节点压力驱逐机制使MEF Center相关镜像被驱逐。

**解决措施<a name="section157021112307"></a>**

1. 清理空间，保证空间足够，且保证“/var/lib/docker“剩余磁盘空间不低于10%。
2. 卸载MEF Center，再重新安装MEF Center并重启MEF Center服务。
    - 若卸载MEF Center失败，请再次执行卸载操作。
    - 若重启MEF Center失败，请执行**docker images**查看被压力驱逐的镜像，并进入对应的镜像路径下“_MEF Center安装路径_/MEF-Center/mef-center/images/_模块名称_/image“，手动执行**docker load -i  _镜像名称_**后，再重启MEF Center即可恢复。

## MEF Edge日志刷屏<a name="ZH-CN_TOPIC_0000001850152197"></a>

**问题现象描述<a name="section13754951112814"></a>**

MEF Edge成功对接CloudCore后，edge\_main产生大量消息，造成日志刷屏。

```text
edgecore proxy receive msg router: {Source: Destination:EdgeCore Option:query Resource:default/node/***}, route: {Source:edge_main Group:resource Operation:response Resource:default/node/***}
```

**原因分析<a name="section77551626153617"></a>**

kube-controller-manager未对节点分配CIDR，而edgecore会持续查询节点状态，直到CIDR被分配为止。

**解决措施<a name="section157021112307"></a>**

1. 登录安装MEF Center的主机。
2. 修改kube-controller-manager配置文件，配置kube-controller-manager的启动参数：cluster-cidr和allocate-node-cidrs。

    kube-controller-manager配置文件通常位于/etc/kubernetes/manifests/kube-controller-manager.yaml

    示例：

    ```text
    --cluster-cidr=192.168.0.0/16 
    --allocate-node-cidrs=true
    ```
