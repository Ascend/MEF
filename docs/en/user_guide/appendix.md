# Appendix<a name="ZH-CN_TOPIC_0000001734129354"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:18:54.336Z pushedAt=2026-06-09T01:46:17.433Z -->

## Public Network Addresses<a name="ZH-CN_TOPIC_0000001722375573"></a>

The open-source code contains the following public network addresses.

> **NOTE**
>
>- Under the MEF Edge installation path, `libcrypto.so.1.1` (the dynamic library  of the open-source OpenSSL) includes the email address `appro\@openssl.org` as part of the license statement. This address is not actually used.
>- Under the MEF Edge installation path, `libc.so.6` (the dynamic library of the open-source edgecore) includes the email address `keld\@dkuug.dk` as part of the license statement. This address is not actually used.
>- Under the MEF Edge installation path, directories such as `software/edge_core/bin/edgecore` and `software/device_plugin/bin/device-plugin` contain open-source software used by the product. Some open-source software uses public network URLs and IPs in the following scenarios: 1) Website information is included when printing logs. 2) Warning messages. 3) Exception information. They are used for informational prompts only and are not accessed during actual operation. There is no security risk.
>- Under the MEF Edge installation path, `www.w3.org` contained in directory files such as `software/edge_installer/bin/edgectl` is the official website of the non-profit organization W3C, and there is no security risk.
>- Under the MEF Center installation path, `libcrypto.so.1.1` (the dynamic library of the open-source OpenSSL) includes the email address `appro\@openssl.org` as part of the license statement. This address is not actually used.
>- Under the MEF Center installation path, directories such as `edge-manager/edge-manager`, `installer/bin/MEF-center-upgrade`, `installer/bin/MEF-center-installer`, `cert-manager/cert-manager`, `nginx-manager/nginx/nginx-manager`, and `installer/bin/MEF-center-controller`:
>   - `www.w3.org` contained in the files is the official website of the non-profit organization W3C, and there is no security risk.
>   - The IPs are use in test cases, and there is no security risk.
>   - `https://ascend-ics-manager.mef-center.svc.cluster.local` and `https://ascend-ics-cert-manager.mef-center.svc.cluster.local` exist as built-in service access addresses, not public network addresses. They pose no security risk.

## Environment Variable Description<a name="ZH-CN_TOPIC_0000002099181517"></a>

| Variable Name | Source | Purpose |
|--|--|--|
| installed-module | Configured in deployment when installing/upgrading MEFCenter. | Installed modules. |
| POD_IP | Configured by kubelet based on container/host related status. | Pod IP address, used to obtain the server listening address within the container. |
| HOST_IP | Configured by kubelet based on container/host related status. | Host IP address, used to generate certificates. |
| KUBE_CLIENT_QPS | Configured in deployment when installing/upgrading MEFCenter. | Client-go request rate. |
| KUBE_CLIENT_BURST | Configured in deployment when installing/upgrading MEFCenter. | Client-go request burst rate. |
| API_SERVER_ENDPOINT | Configured in deployment when installing/upgrading MEFCenter. | k8s apiserver url |
| LOG_UPLOAD_CONCURRENCY | Configured in deployment when installing/upgrading MEFCenter. | Maximum concurrency for log upload. |
| EdgeMgrSvcPort | Configured in deployment when installing/upgrading MEFCenter. | Port for edge-manager listening. |
| AlarmMgrSvcPort | Configured in deployment when installing/upgrading MEFCenter. | Port for alarm-manager listening. |
| CertMgrSvcPort | Configured in deployment when installing/upgrading MEFCenter. | Port for cert-manager listening. |
| NginxSslPort | Configured in deployment when installing/upgrading MEFCenter. | Port for nginx-manager listening. |
| AuthPort | Configured in deployment when installing/upgrading MEFCenter. | Port for edge-manager listening (southbound authentication port). |
| WebsocketPort | Configured in deployment when installing/upgrading MEFCenter. | Port for edge-manager listening (southbound management port). |
| LD_LIBRARY_PATH | Existing environment variable in the system. | Dynamic library loading path. |

## User Information List<a name="ZH-CN_TOPIC_0000001722375493"></a>

**Table 1**  User information list

<table><thead align="left"><tr id="row1933934182920"><th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.1"><p id="p63392419291"><a name="p63392419291"></a><a name="p63392419291"></a>Username</p>
</th>
<th class="cellrowborder" valign="top" width="35%" id="mcps1.2.5.1.2"><p id="p11339114162912"><a name="p11339114162912"></a><a name="p11339114162912"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p733934192915"><a name="p733934192915"></a><a name="p733934192915"></a>Initial Password</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p5613193017"><a name="p5613193017"></a><a name="p5613193017"></a>Password Change Method</p>
</th>
</tr>
</thead>
<tbody><tr id="row1633917419290"><td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.1 "><p id="p163397412916"><a name="p163397412916"></a><a name="p163397412916"></a>MEFEdge</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.2 "><p id="p433934112918"><a name="p433934112918"></a><a name="p433934112918"></a><span id="ph3525163816271"><a name="ph3525163816271"></a><a name="ph3525163816271"></a>MEF Edge</span> running user for some processes. This user cannot log in.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p13395420290"><a name="p13395420290"></a><a name="p13395420290"></a>No initial password</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1417182710317"><a name="p1417182710317"></a><a name="p1417182710317"></a>-</p>
</td>
</tr>
<tr id="row126363469315"><td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.1 "><p id="p663620466311"><a name="p663620466311"></a><a name="p663620466311"></a>MEFCenter</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.2 "><p id="p46361469310"><a name="p46361469310"></a><a name="p46361469310"></a><span id="ph17253174072711"><a name="ph17253174072711"></a><a name="ph17253174072711"></a>MEF Center</span> service process running user. This user cannot log in.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p663613461316"><a name="p663613461316"></a><a name="p663613461316"></a>No initial password</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p18636946143115"><a name="p18636946143115"></a><a name="p18636946143115"></a>-</p>
</td>
</tr>
<tr id="row246494613212"><td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.1 "><p id="p1546454663215"><a name="p1546454663215"></a><a name="p1546454663215"></a>root</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.2 "><p id="p7465144673214"><a name="p7465144673214"></a><a name="p7465144673214"></a>Installation and running user for some processes of <span id="ph17809144104311"><a name="ph17809144104311"></a><a name="ph17809144104311"></a>MEF Edge</span> and <span id="ph1922216513327"><a name="ph1922216513327"></a><a name="ph1922216513327"></a>MEF Center</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p16465746123215"><a name="p16465746123215"></a><a name="p16465746123215"></a>For the default password, see <span id="ph58841112362"><a name="ph58841112362"></a><a name="ph58841112362"></a><a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100235027?idPath=23710424|251366513|22892968|252764743" target="_blank" rel="noopener noreferrer">Atlas Series Hardware Product Account List</a></span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p546524683213"><a name="p546524683213"></a><a name="p546524683213"></a>Execute the <strong id="b46051981449"><a name="b46051981449"></a><a name="b46051981449"></a>passwd <em id="i812695713434"><a name="i812695713434"></a><a name="i812695713434"></a>&lt;username&gt;</em> </strong> command as the root user to change.</p>
</td>
</tr>
</tbody>
</table>

**Table 2**  K8s users

|Username|Description|Initial Password|Password Change Method|
|--|--|--|--|
|MindXMEF|A normal user configured by MEF Center for authentication with the API server. Uses certificate authentication, no password.|None|-|

**K8s ServiceAccount<a name="section10713432131611"></a>**

**Table 3**  K8s ServiceAccount

|Account Name|Description|Initial Password|Password Change Method|
|--|--|--|--|
|default|<li>The default serviceaccount created with the mef-center namespace in K8s.</li><li>The default serviceaccount created with the mef-user namespace in K8s.</li>|None|-|

**Users in the inference image from the Dockerfile example<a name="section9469132314172"></a>**

| User | Initial Password | Password Change Method |
|--|--|--|
| HwHiAiUser | None | - |
| HwSysUser | None | - |
| HwDmUser | None | - |
| HwBaseUser | None | - |

**Users in the Ubuntu base image<a name="zh-cn_topic_0000001515257736_zh-cn_topic_0000001446965016_section158195363315"></a>**

| User | Initial Password | Password Change Method |
|--|--|--|
| root | None | - |
| daemon | None | - |
| bin | None | - |
| sys | None | - |
| sync | None | - |
| games | None | - |
| man | None | - |
| lp | None | - |
| mail | None | - |
| news | None | - |
| uucp | None | - |
| proxy | None | - |
| www-data | None | - |
| backup | None | - |
| list | None | - |
| irc | None | - |
| gnats | None | - |
| nobody | None | - |
| _apt | None | - |
| MEFCenter | None | - |
