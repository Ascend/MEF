# 附录<a name="ZH-CN_TOPIC_0000001734129354"></a>

## 公网地址<a name="ZH-CN_TOPIC_0000001722375573"></a>

开源代码包含的公网地址请参考[MEF 7.3.0 MEF公网地址.xlsx](../resource/MEF%207.3.0%20MEF公网地址.xlsx)。

> [!NOTE] 说明 
> 
>- MEF Edge的安装路径下，libcrypto.so.1.1为开源软件OpenSSL的动态链接库，其中的电子邮件地址“appro@openssl.org”是许可声明的一部分，实际并未使用。
>- MEF Edge的安装路径下，libc.so.6为开源软件edgecore的动态链接库，其中的电子邮件地址“keld@dkuug.dk”是许可声明的一部分，实际并未使用。
>- MEF Edge的安装路径下，“software/edge\_core/bin/edgecore“、"software/device\_plugin/bin/device-plugin", 等目录中为产品使用的开源软件，部分开源软件在以下场景中使用公网URL和IP：1. 在打印日志时会附带网址信息。2. warning提示。3. 异常信息。以上内容均用于信息提示，实际运行时不会访问该网址，没有安全风险。
>- MEF Edge的安装路径下,"software/edge\_installer/bin/edgectl"等目录文件中包含的www.w3.org为非盈利组织W3C官网，没有安全风险。
>- MEF Center的安装路径下，libcrypto.so.1.1为开源软件OpenSSL的动态链接库，其中的电子邮件地址“appro@openssl.org”是许可声明的一部分，实际并未使用。
>- MEF Center的安装路径下，“edge-manager/edge-manager”、"installer/bin/MEF-center-upgrade"、"installer/bin/MEF-center-installer"、"cert-manager/cert-manager"、“nginx-manager/nginx/nginx-manager”、"installer/bin/MEF-center-controller"等目录：
>    - 文件包含的www.w3.org为非盈利组织W3C官网，没有安全风险。
>    - IP为测试用例中设置，没有安全风险。
>    - 存在 https://ascend-ics-manager.mef-center.svc.cluster.local 和 https://ascend-ics-cert-manager.mef-center.svc.cluster.local，为内置访问服务地址，非公网地址，没有安全风险。

## 环境变量说明<a name="ZH-CN_TOPIC_0000002099181517"></a>

|变量名|来源|作用|
|--|--|--|
|installed-module|安装/升级MEFCenter时，在deployment中配置。|已安装的模块。|
|POD_IP|kubelet根据容器/主机相关状态配置。|pod IP，用于在容器内获取服务器侦听地址。|
|HOST_IP|kubelet根据容器/主机相关状态配置。|主机IP，用于生成证书。|
|KUBE_CLIENT_QPS|安装/升级MEFCenter时，在deployment中配置。|client go请求速率。|
|KUBE_CLIENT_BURST|安装/升级MEFCenter时，在deployment中配置。|client go请求爆发速率。|
|API_SERVER_ENDPOINT|安装/升级MEFCenter时，在deployment中配置。|k8s apiserver url|
|LOG_UPLOAD_CONCURRENCY|安装/升级MEFCenter时，在deployment中配置。|日志上传最大并发数。|
|EdgeMgrSvcPort|安装/升级MEFCenter时，在deployment中配置。|edge-manager组件侦听的端口。|
|AlarmMgrSvcPort|安装/升级MEFCenter时，在deployment中配置。|alarm-manager组件侦听的端口。|
|CertMgrSvcPort|安装/升级MEFCenter时，在deployment中配置。|cert-manager组件侦听的端口。|
|NginxSSlPort|安装/升级MEFCenter时，在deployment中配置。|nginx-manager组件侦听的端口。|
|AuthPort|安装/升级MEFCenter时，在deployment中配置。|edge-manager组件侦听的端口（南向认证端口）。|
|WebsocketPort|安装/升级MEFCenter时，在deployment中配置。|edge-manager组件侦听的端口（南向管理端口）。|
|LD_LIBRARY_PATH|系统中已存在的环境变量。|动态库加载路径。|

## 用户信息列表<a name="ZH-CN_TOPIC_0000001722375493"></a>

**表 1**  用户信息列表

<table><thead align="left"><tr id="row1933934182920"><th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.1"><p id="p63392419291"><a name="p63392419291"></a><a name="p63392419291"></a>用户名</p>
</th>
<th class="cellrowborder" valign="top" width="35%" id="mcps1.2.5.1.2"><p id="p11339114162912"><a name="p11339114162912"></a><a name="p11339114162912"></a>描述</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p733934192915"><a name="p733934192915"></a><a name="p733934192915"></a>初始密码</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p5613193017"><a name="p5613193017"></a><a name="p5613193017"></a>密码修改方法</p>
</th>
</tr>
</thead>
<tbody><tr id="row1633917419290"><td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.1 "><p id="p163397412916"><a name="p163397412916"></a><a name="p163397412916"></a>MEFEdge</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.2 "><p id="p433934112918"><a name="p433934112918"></a><a name="p433934112918"></a><span id="ph3525163816271"><a name="ph3525163816271"></a><a name="ph3525163816271"></a>MEF Edge</span>部分进程运行用户，该用户不可登录</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p13395420290"><a name="p13395420290"></a><a name="p13395420290"></a>无初始密码</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1417182710317"><a name="p1417182710317"></a><a name="p1417182710317"></a>-</p>
</td>
</tr>
<tr id="row126363469315"><td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.1 "><p id="p663620466311"><a name="p663620466311"></a><a name="p663620466311"></a>MEFCenter</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.2 "><p id="p46361469310"><a name="p46361469310"></a><a name="p46361469310"></a><span id="ph17253174072711"><a name="ph17253174072711"></a><a name="ph17253174072711"></a>MEF Center</span>服务进程运行用户，该用户不可登录</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p663613461316"><a name="p663613461316"></a><a name="p663613461316"></a>无初始密码</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p18636946143115"><a name="p18636946143115"></a><a name="p18636946143115"></a>-</p>
</td>
</tr>
<tr id="row246494613212"><td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.1 "><p id="p1546454663215"><a name="p1546454663215"></a><a name="p1546454663215"></a>root</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.2 "><p id="p7465144673214"><a name="p7465144673214"></a><a name="p7465144673214"></a><span id="ph17809144104311"><a name="ph17809144104311"></a><a name="ph17809144104311"></a>MEF Edge</span>和<span id="ph1922216513327"><a name="ph1922216513327"></a><a name="ph1922216513327"></a>MEF Center</span>安装和部分进程运行用户</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p16465746123215"><a name="p16465746123215"></a><a name="p16465746123215"></a>默认密码请参见<span id="ph58841112362"><a name="ph58841112362"></a><a name="ph58841112362"></a>《<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100235027?idPath=23710424|251366513|22892968|252764743" target="_blank" rel="noopener noreferrer">Atlas 系列硬件产品 账户清单</a>》</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p546524683213"><a name="p546524683213"></a><a name="p546524683213"></a>以root用户执行<strong id="b46051981449"><a name="b46051981449"></a><a name="b46051981449"></a>passwd <em id="i812695713434"><a name="i812695713434"></a><a name="i812695713434"></a>&lt;username&gt;</em> </strong>命令修改</p>
</td>
</tr>
</tbody>
</table>

**表 2**  K8s用户

|用户名|描述|初始密码|密码修改方法|
|--|--|--|--|
|MindXMEF|MEF Center配置的与API server认证的普通用户。使用证书认证，无密码。|无|-|

**K8s的ServiceAccount<a name="section10713432131611"></a>**

**表 3**  k8s服务账号（ServiceAccount）

|账号名|说明|初始密码|密码修改方法|
|--|--|--|--|
|default|<li>K8s内随mef-center命名空间创建的默认serviceaccount。</li><li>K8s内随mef-user命名空间创建的默认serviceaccount。</li>|无|-|

**Dockerfile示例中推理镜像中的用户<a name="section9469132314172"></a>**

|用户|初始密码|密码修改方法|
|--|--|--|
|HwHiAiUser|无|-|
|HwSysUser|无|-|
|HwDmUser|无|-|
|HwBaseUser|无|-|

**ubuntu基础镜像中的用户<a name="zh-cn_topic_0000001515257736_zh-cn_topic_0000001446965016_section158195363315"></a>**

|用户|初始密码|密码修改方法|
|--|--|--|
|root|无|-|
|daemon|无|-|
|bin|无|-|
|sys|无|-|
|sync|无|-|
|games|无|-|
|man|无|-|
|lp|无|-|
|mail|无|-|
|news|无|-|
|uucp|无|-|
|proxy|无|-|
|www-data|无|-|
|backup|无|-|
|list|无|-|
|irc|无|-|
|gnats|无|-|
|nobody|无|-|
|_apt|无|-|
|MEFCenter|无|-|
