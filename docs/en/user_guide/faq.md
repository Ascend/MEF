# FAQ<a name="ZH-CN_TOPIC_0000001674256326"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:17:41.798Z pushedAt=2026-06-09T01:46:21.978Z -->

## MEF Center Installation Fails When Importing the Base Image<a name="ZH-CN_TOPIC_0000001749571521"></a>

**Symptom<a name="section13754951112814"></a>**

The user fails to install MEF Center by importing the local Ubuntu:22.04 base image using the `docker load` command.

**Root Cause Analysis<a name="section77551626153617"></a>**

When the Docker version in the MEF Center installation environment is 23.0 or later, Docker enables BuildKit by default as the image building tool to build images and re-obtains the dependent base images through the image registry. When MEF Center builds the dependent base image Ubuntu:22.04, if the image registry configured in the environment or the Docker public image registry is unavailable, and the base image is obtained through offline import, the MEF Center component image build will fail, resulting in the failure of MEF Center installation.

**Solution<a name="section157021112307"></a>**

Before installing MEF Center, run the following command to disable the Docker BuildKit feature by setting an environment variable.

```bash
export DOCKER_BUILDKIT=0
```

## MEF Edge Starts or Stops Repeatedly<a id="mef-edge-repeated-start-or-stop-alarm"></a>

**Symptom<a name="section281692625610"></a>**

After the MEF Edge software has started or stopped, executing the `start` or `stop` command repeatedly will return a warning prompt. An example of the echo is shown below.

- Repeated start

    ```text
    warning: component [edge-om] is already started!
    warning: component [edge-main] is already started!
    warning: component [edgecore] is already started!
    warning: component [device-plugin] is already started!
    ```

- Repeated stop

    ```text
    warning: component [edge-om] is already stopped!
    warning: component [edge-main] is already stopped!
    warning: component [edgecore] is already stopped!
    warning: component [device-plugin] is already stopped!
    ```

**Solution<a name="section3191818161414"></a>**

This prompt will not appear if the start or stop command is not executed repeatedly.

## Restoring the MEF Center Upgrade Environment After Forced Termination<a name="ZH-CN_TOPIC_0000001722295465"></a>

**Symptom<a name="section195476587165"></a>**

During the MEF Center upgrade, if the upgrade is forcibly terminated due to device power-on/power-off or other exceptions, you need to clear the image and node labels to restore the environment.

**Solution<a name="section5415101941714"></a>**

1. Log in to the device environment.
2. Run the following command to restore the upgrade environment.

    ```bash
    run.sh start
    ```

## Restoring the MEF Edge Upgrade Environment After Forced Termination<a name="ZH-CN_TOPIC_0000001722295469"></a>

**Symptom<a name="section2448162514186"></a>**

During the MEF Edge upgrade, the upgrade is forcibly terminated due to device power-on/power-off or other exceptions. In this case, the environment needs to be restored.

**Solution<a name="section172121345201815"></a>**

1. Log in to the device environment.
2. If the residual file "/home/data/mefedge/unpack/edge_installer" exists, delete it.
3. Check the current MEF Edge version. If the upgrade was unsuccessful, perform the upgrade again.
    - Run the following command to enter the installation directory.

        ```bash
        cd MEFEdge installation path/MEFEdge/software
        ```

    - Run the following command to view the MEF Edge version in the version.xml file.

        ```bash
        cat version.xml
        ```

4. If the MEF Edge version in version.xml is the target version, run the following command to restore the upgrade environment.

    ```bash
    run.sh start
    ```

## Node Pressure Eviction Mechanism Causes MEF Center to Run Abnormally<a id="ZH-CN_TOPIC_0000001780849121"></a>

**Symptom<a name="section13754951112814"></a>**

When deploying and running MEF Center, K8s triggers the node pressure eviction mechanism due to insufficient node resources such as memory, disk, and PID, causing MEF Center services to fail to run normally.

**Root Cause Analysis<a name="section77551626153617"></a>**

The node pressure eviction mechanism caused MEF Center-related images to be evicted.

**Solution<a name="section157021112307"></a>**

1. Clean up space to ensure sufficient capacity, and ensure that the remaining disk space of "/var/lib/docker" is no less than 10%.
2. Uninstall MEF Center, then reinstall MEF Center and restart the MEF Center service.
    - If uninstalling MEF Center fails, perform the uninstall operation again.
    - If restarting MEF Center fails, run **docker images** to check for evicted images, navigate to the corresponding image path "_MEF Center installation path_/MEF-Center/mef-center/images/_module name_/image", manually execute <b>docker load -i _image name_</b>, and then restart MEF Center to recover.

## MEF Edge Log Refreshing<a name="ZH-CN_TOPIC_0000001850152197"></a>

**Symptom<a name="section13754951112814"></a>**

After MEF Edge successfully connects to CloudCore, edge_main generates a large number of messages, causing log flooding.

```text
edgecore proxy receive msg router: {Source: Destination:EdgeCore Option:query Resource:default/node/***}, route: {Source:edge_main Group:resource Operation:response Resource:default/node/***}
```

**Root Cause Analysis<a name="section77551626153617"></a>**

The kube-controller-manager has not assigned a CIDR to the node, and edgecore will keep querying the node status until a CIDR is assigned.

**Solution<a name="section157021112307"></a>**

1. Log in to the host where MEF Center is installed.
2. Modify the kube-controller-manager configuration file and configure the startup parameters of kube-controller-manager: cluster-cidr and allocate-node-cidrs.

    The kube-controller-manager configuration file is usually located at /etc/kubernetes/manifests/kube-controller-manager.yaml

    Example:

    ```text
    --cluster-cidr=192.168.0.0/16
    --allocate-node-cidrs=true
    ```
