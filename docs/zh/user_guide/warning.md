# 告警处理<a name="ZH-CN_TOPIC_0000001674415906"></a>

## 告警事件说明<a name="ZH-CN_TOPIC_0000001674415930"></a>

- 告警：从某个时间点开始出现的问题。
- 事件：在过去某个时间点发生的特定事件。

<table><thead align="left"><tr id="row5323212152618"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="p15323121210269"><a name="p15323121210269"></a><a name="p15323121210269"></a>告警类型</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="p1732341222611"><a name="p1732341222611"></a><a name="p1732341222611"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="p9323151242611"><a name="p9323151242611"></a><a name="p9323151242611"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row1732321220263"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p1532321212269"><a name="p1532321212269"></a><a name="p1532321212269"></a>云边协同告警事件</p>
<p id="p7323151218262"><a name="p7323151218262"></a><a name="p7323151218262"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p33231412192613"><a name="p33231412192613"></a><a name="p33231412192613"></a>通用云边协同告警事件。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p4323191213267"><a name="p4323191213267"></a><a name="p4323191213267"></a><a href="#通用告警事件">通用告警事件</a></p>
</td>
</tr>
<tr id="row7323131282612"><td class="cellrowborder" valign="top" headers="mcps1.1.4.1.1 "><p id="p183231712122613"><a name="p183231712122613"></a><a name="p183231712122613"></a><span id="ph133231212142610"><a name="ph133231212142610"></a><a name="ph133231212142610"></a>MEF Center</span>云边协同告警事件。</p>
<p id="p232317121261"><a name="p232317121261"></a><a name="p232317121261"></a>在<span id="ph1232331218265"><a name="ph1232331218265"></a><a name="ph1232331218265"></a>MEF Edge</span>对接<span id="ph1732311126262"><a name="ph1732311126262"></a><a name="ph1732311126262"></a>MEF Center</span>进行云边协同时出现的告警或事件。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.4.1.2 "><p id="p16323151222617"><a name="p16323151222617"></a><a name="p16323151222617"></a><a href="#MEF Center云边协同告警事件">MEF Center云边协同告警事件</a></p>
</td>
</tr>
<tr id="row123241412142615"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p1832361211261"><a name="p1832361211261"></a><a name="p1832361211261"></a><span id="ph203231912162615"><a name="ph203231912162615"></a><a name="ph203231912162615"></a>MEF Center</span>设备告警事件</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p820825516497"><a name="p820825516497"></a><a name="p820825516497"></a><span id="ph120819554495"><a name="ph120819554495"></a><a name="ph120819554495"></a>MEF Center</span>设备和第三方管理平台、软件仓库、或镜像仓库对接时产生的告警事件。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p16324101242611"><a name="p16324101242611"></a><a name="p16324101242611"></a><a href="#MEF Center设备告警事件">MEF Center设备告警事件</a></p>
</td>
</tr>
</tbody>
</table>

## 通用告警事件<a id="通用告警事件"></a>

### 0x00131001 Docker引擎异常（紧急告警）<a name="ZH-CN_TOPIC_0000001722295449"></a>

**告警解释<a name="zh-cn_topic_0190357109_section10542121642"></a>**

告警描述：Docker引擎异常。

当Docker引擎未正常运行时产生此告警；Docker引擎恢复正常运行时告警清除。

产生此告警的模块为：MEF Edge。

**告警属性<a name="zh-cn_topic_0190357109_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x00131001|紧急|是|

**对系统的影响<a name="zh-cn_topic_0190357109_section85431015414"></a>**

无法部署和运行容器。

**可能原因<a name="zh-cn_topic_0190357109_section95431611040"></a>**

Docker引擎未正常运行。

**处理步骤<a name="section1420215111613"></a>**

1. 登录设备CLI界面，执行**systemctl restart docker**命令重启Docker引擎，查看告警是否清除。
2. 若告警无法清除，请联系华为技术支持。

### 0x00131011  MEF Edge日志空间满（一般告警）<a name="ZH-CN_TOPIC_0000001674415902"></a>

**告警解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

告警描述：MEF Edge日志空间即将写满。

当MEF Edge日志和日志转储文件目录已占用空间达到80%以上产生此告警；当已占用空间低于此阈值时，告警消除。

产生此告警的模块为：MEF Edge。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x00131011|一般|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

MEF Edge无法记录日志；edgecore无法正常启动。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

MEF Edge日志目录存储空间不足。

**处理步骤<a name="section1817111521104"></a>**

1. 检查MEF Edge日志目录是否有足够的存储空间。
    - 执行**df** **-lh**命令，查看各分区的占用率。当分区占用率超过80%时，会上报告警。

        例如，查询/var/alog目录下各文件的大小。

        显示信息如下：表示/var/alog目录占用率为80%，产生告警。

        ```text
        Euler:/var/alog # df -lh
        Filesystem      Size  Used Avail Use% Mounted on
        /dev/root       5.9G  1.8G  3.8G  32% /
        devtmpfs        5.5G     0  5.5G   0% /dev
        tmpfs           5.7G     0  5.7G   0% /dev/shm
        tmpfs           5.7G  650M  5.1G  12% /run
        tmpfs           4.0M     0  4.0M   0% /sys/fs/cgroup
        tmpfs           5.7G   10M  5.7G   1% /tmp
        tmpfs           128M     0  128M   0% /var/IEF
        tmpfs           128M  118M   11M  92% /var/alog
        tmpfs           128M     0  128M   0% /var/dlog
        tmpfs           128M  128M     0 100% /var/log
        tmpfs           128M  320K  128M   1% /var/plog
        /dev/mmcblk0p5  974M   55M  853M   6% /home/log
        /dev/mmcblk0p4  974M  332M  576M  37% /home/data
        /dev/mmcblk0p7  2.9G  286M  2.5G  11% /home/package
        /dev/mmcblk0p6  2.0G  175M  1.7G  10% /usr/local/mindx
        ```

    - 执行**du -sh **_目录_**/\***命令，查看分区占用率超过80%的目录下各文件的大小。

        例如，执行**du -sh /var/alog/\***，查询/var/alog目录下各文件的大小。

        显示信息如下：

        ```text
        Euler:~ # du -sh /var/alog/*
        90M	/var/alog/big_file
        28M	/var/alog/MEFEdge_log
        ```

2. 备份数据后处理目录空间不足的文件，释放空间。
    - 执行**mv**命令，移走对应目录中的文件。
    - 执行**rm**命令，删除对应目录中的文件。

3. 若上述操作不能解决该告警，请联系华为技术支持。

### 0x00131012 NPU异常（紧急告警）<a name="ZH-CN_TOPIC_0000001722375465"></a>

**告警解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

告警描述：NPU芯片处于不健康状态。

当NPU芯片健康状态不正常时，产生此告警；当NPU芯片恢复健康时，告警清除。

产生此告警的模块为：MEF Edge。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x00131012|紧急|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

推理容器无法部署或运行。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

NPU芯片处于不健康状态。

**处理步骤<a name="section1817111521104"></a>**

1. 登录设备，使用**npu-smi info**命令检查不健康的NPU信息。
2. 请联系华为技术支持并提供对应信息。

### 0x00131004 容器状态重启事件（正常事件）<a name="ZH-CN_TOPIC_0000001722375545"></a>

**事件解释<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section14304185863417"></a>**

事件描述：容器应用状态重启。

当MEF Edge检测到通过网管部署的容器应用重启时，产生此事件。

产生此事件的模块为：MEF Edge。

**事件属性<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section123741058103416"></a>**

**表 1**  事件信息

|事件ID|事件级别|可自动清除|
|--|--|--|
|0x00131004|正常|否|

**对系统的影响<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section17423131514189"></a>**

可能导致系统容器应用状态异常。

**可能原因<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section13451932141819"></a>**

容器应用异常退出或配置变更导致容器重启。

**处理步骤<a name="section18190182917212"></a>**

1. 登录设备，排查容器重启原因。
    - 是，处理完毕。
    - 否，执行[2](#zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341)。

2. <a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341"></a>请联系华为技术支持。

### 0x00131014 数据库异常（严重告警）<a name="ZH-CN_TOPIC_0000001722295521"></a>

**告警解释<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section14304185863417"></a>**

告警描述：MEF Edge数据库异常。

当MEF Edge检测到数据库格式异常时，产生此告警。

产生此告警的模块为：MEF Edge。

**告警属性<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section123741058103416"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x00131014|严重|是|

**对系统的影响<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section17423131514189"></a>**

边缘系统将停止运行或运行不正常。

**可能原因<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section13451932141819"></a>**

MEF Edge数据库格式错误。

**处理步骤<a name="zh-cn_topic_0000001182494244_section6849845132315"></a>**

1. 登录设备，检查数据库格式是否正常。
    - 是，处理完毕。
    - 否，执行[2](#zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341)。

2. <a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341"></a>请联系华为技术支持。

## MEF Center云边协同告警事件<a id="MEF Center云边协同告警事件"></a>

### 0x00131013  MEF Center根证书超期（严重告警）<a name="ZH-CN_TOPIC_0000001674256230"></a>

**告警解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

告警描述：MEF Edge设备中的MEF Center根证书（用于签发MEF Center服务证书）即将过期或已经过期。

当MEF Edge设备中的MEF Center根证书小于证书超期告警时间阈值时，产生此告警；更新为有效证书后，此告警消失。

产生此告警的模块为：MEF Edge。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x00131013|严重|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

MEF Center根证书可能超期。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

MEF Center根证书即将超期或已经过期。

**处理步骤<a name="section1817111521104"></a>**

1. 重新获取MEF Center根证书和云边认证token，并在MEF Edge设备上进行网管配置，具体操作请参见[MEF Center和MEF Edge认证对接](./usage.md#mef-center和mef-edge认证对接)。
2. 如果告警未自动清除，请联系华为技术支持并提供对应信息。

### 0x01000004 签发MEF Edge服务证书的根证书即将过期（一般告警）<a name="ZH-CN_TOPIC_0000001722375589"></a>

告警描述：MEF Center设备中的用于签发MEF Edge服务证书的根证书即将过期或已经过期，产生告警。

- 当MEF Center设备中的根证书即将过期（24小时 < 证书超期时间 < 15天）后，产生此告警，且MEF Edge设备重新申请签发服务证书，更新流程结束后告警消失。
- 当MEF Center设备中的根证书已经过期或证书超期时间小于等于24小时，将强制更新MEF Center设备中的该证书，告警消失。

产生此告警的模块为：MEF Center。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x01000004|一般|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

MEF Edge和MEF Center可能无法正常对接。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

用于签发MEF Edge服务证书的根证书即将过期或已经过期。

**处理步骤<a name="section1817111521104"></a>**

1. MEF Center证书管理模块检测到根证书需要更新并且成功触发更新流程。
2. 如果告警未自动清除，请联系华为技术支持。

### 0x01000005  MEF Center根证书即将过期（一般告警）<a name="ZH-CN_TOPIC_0000001722375477"></a>

**告警解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

告警描述：MEF Center设备中的MEF Center根证书（用于签发MEF Center服务证书）即将过期或已经过期，产生告警。

- 当MEF Center设备中的MEF Center根证书即将过期（24小时 < 证书超期时间 < 15天）后，产生此告警，且MEF Edge设备完成MEF Center根证书更新，更新流程结束后告警消失。
- 当MEF Center设备中的MEF Center根证书已经过期或证书超期时间小于等于24小时，将强制更新MEF Center设备中的该证书，告警消失。

产生此告警的模块为：MEF Center。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x01000005|一般|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

MEF Edge和MEF Center可能无法正常对接。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

MEF Center根证书（用于签发MEF Center服务证书）即将过期或已经过期。

**处理步骤<a name="section1817111521104"></a>**

1. MEF Center证书管理模块检测到根证书需要更新并且成功触发更新流程。
2. 如果告警未自动清除，请联系华为技术支持。

### 0x01000006 签发MEF Edge服务证书的根证书自动更新流程异常（严重事件）<a name="ZH-CN_TOPIC_0000001722375473"></a>

**事件解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

事件描述：MEF Center设备中的用于签发MEF Edge服务证书的根证书更新后，仍存在未成功申请签发对应服务证书的MEF Edge设备。

产生此事件的模块为：MEF Center。

**事件属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  事件信息

|事件ID|告警级别|可自动清除|
|--|--|--|
|0x01000006|严重|否|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

更新服务证书失败的MEF Edge设备无法与MEF Center再次对接。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

MEF Edge由于网络中断或者其他原因未能正常完成签发MEF Edge服务证书的申请。

**处理步骤<a name="section1817111521104"></a>**

1. 参考[MEF Edge自动更新MEF Center根证书失败](./troubleshooting.md#mef-edge自动更新mef-center根证书失败)，重新获取MEF Center根证书和云边认证token，并在MEF Edge设备上进行网管配置。
2. 如果告警未自动清除，请联系华为技术支持并提供对应信息。

### 0x01000007  MEF Center根证书自动更新流程异常（严重事件）<a name="ZH-CN_TOPIC_0000001674415998"></a>

**事件解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

事件描述：MEF Center设备中的MEF Center根证书（用于签发MEF Center服务证书）更新后，仍存在未成功更新该MEF Center根证书的MEF Edge设备。

产生此事件的模块为：MEF Center。

**事件属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  事件信息

|事件ID|告警级别|可自动清除|
|--|--|--|
|0x01000007|严重|否|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

更新MEF Center根证书失败的MEF Edge设备无法与MEF Center再次对接。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

MEF Edge由于网络中断或者其他原因未能正常完成MEF Center根证书（用于签发MEF Center服务证书）的更新。

**处理步骤<a name="section1817111521104"></a>**

1. 参考[MEF Edge自动更新MEF Center根证书失败](./troubleshooting.md#mef-edge自动更新mef-center根证书失败)，重新获取MEF Center根证书和云边认证token，并在MEF Edge设备上进行网管配置。
2. 如果告警未自动清除，请联系华为技术支持并提供对应信息。

## MEF Center设备告警事件<a id="MEF Center设备告警事件"></a>

### 0x01000001第三方管理平台根证书超期（严重告警）<a name="ZH-CN_TOPIC_0000001722295473"></a>

**告警解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

告警描述：MEF Center设备中的第三方管理平台根证书即将过期或已经过期。

当证书有效期小于证书超期告警时间阈值时，产生此告警；更新为有效证书后，此告警消失。

产生此告警的模块为：MEF Center。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x01000001|严重|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

证书过期后，第三方管理平台无法访问MEF Center。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

证书即将超期或已经过期。

**处理步骤<a name="section1817111521104"></a>**

1. 检查证书是否过期或即将超期并重新导入有效证书。
2. 如果告警未自动清除，请联系华为技术支持并提供对应信息。

### 0x01000002软件仓根证书超期（严重告警）<a name="ZH-CN_TOPIC_0000001722375505"></a>

**告警解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

告警描述：MEF Center设备中的软件仓根证书即将过期或已经过期。

当证书有效期小于证书超期告警时间阈值时，产生此告警；更新为有效证书后，此告警消失。

产生此告警的模块为：MEF Center。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x01000002|严重|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

证书过期后，MEF Center无法访问软件仓库。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

证书即将超期或已经过期。

**处理步骤<a name="section1817111521104"></a>**

1. 检查证书是否过期或即将超期并重新导入有效证书。
2. 如果告警未自动清除，请联系华为技术支持并提供对应信息。

### 0x01000003镜像仓根证书超期（严重告警）<a name="ZH-CN_TOPIC_0000001674256222"></a>

**告警解释<a name="zh-cn_topic_0176114050_section10542121642"></a>**

告警描述：MEF Center设备中的镜像仓根证书即将过期或已经过期。

当证书有效期小于证书超期告警时间阈值时，产生此告警；更新为有效证书后，此告警消失。

产生此告警的模块为：MEF Center。

**告警属性<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**表 1**  告警信息

|告警ID|告警级别|可自动清除|
|--|--|--|
|0x01000003|严重|是|

**对系统的影响<a name="zh-cn_topic_0176114050_section85431015414"></a>**

证书过期后，MEF Center无法访问镜像仓库。

**可能原因<a name="zh-cn_topic_0176114050_section95431611040"></a>**

证书即将超期或已经过期。

**处理步骤<a name="section1817111521104"></a>**

1. 检查证书是否过期或即将超期并重新导入有效证书。
2. 如果告警未自动清除，请联系华为技术支持并提供对应信息。
