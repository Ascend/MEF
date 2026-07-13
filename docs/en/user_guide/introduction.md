# Introduction<a name="ZH-CN_TOPIC_0000001722295433"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:18:21.622Z pushedAt=2026-06-09T01:46:25.958Z -->

**Background<a name="section1277813591138"></a>**

With the evolution of AI technology, the demand for intelligent transformation in industries such as transportation, energy, and security has become increasingly strong, leading to more implementations of device-edge-cloud applications. Massive numbers of terminal devices generate data in real time, and centralized cloud computing struggles to accommodate the need for frequent data interaction in terms of bandwidth load, network latency, and data management costs, further highlighting the value of edge computing. The growth in edge computing demand inevitably brings an increase in the number of edge computing devices. How to manage a large number of edge devices and how to deploy edge computing applications to edge devices in batches have become key concerns for system administrators.

**Product Definition<a name="section141542218147"></a>**

MEF is a lightweight device-edge-cloud synergy enablement framework positioned for integration. As part of the Ascend inference solution, MEF is used for intelligent edge device enablement, providing edge node management and intelligent inference service (containerized application) management functions. Edge-cloud synergy management can be performed through MEF Edge and MEF Center, and users can integrate required functions through secondary development and integration with ISV (Independent Software Vendor) business platforms.

- MEF Edge is deployed on intelligent edge devices, responsible for interfacing with the central network management system, completing the deployment and management of intelligent inference services (containerized applications), and providing services for algorithm applications.
- MEF Center is deployed on general-purpose servers, responsible for batch management of edge nodes, service deployment, and system monitoring.

**Product Value<a name="section111014161153"></a>**

MEF provides benefits both on the service plane and management plane.

**Table 1** Product value

| Plane | Product Value |
|--|--|
| Service Plane | MEF features an open ecosystem and full-stack enablement, facilitating integration and lowering the barrier to entry for users. |
| Management Plane | MEF is extremely simple and easy to use, secure, and reliable. |

# MEF Architecture<a name="ZH-CN_TOPIC_0000001674256314"></a>

MEF relies on the open-source system KubeEdge to establish and manage the control link between MEF Center and MEF Edge. MEF Center provides RESTful interfaces for users, which can be integrated and called by other third-party applications, allowing them to access services through the interface.

**Figure 1** MEF architecture<a name="fig879538143318"></a>
![MEF architecture diagram](../figures/MEF-architecture.png)

- MEF Center is the central management software used by MEF to provide external interfaces for integration with ISV business platforms and for cloud-edge synergy with MEF Edge. This software integrates modules such as the node management module and the container management module, providing functions like node management services and containerized application management services.

    **Table 1** MEF Center module description

    |Module|Module Function|
    |--|--|
    |APIG (API Gateway)|Provides bidirectional authentication and RESTful interfaces for ISV business platforms. Used by ISV business platforms to invoke and use services provided by MEF Center.|
    |edge-manager|Edge node management module and containerized application management module. Manages edge node access and containerized applications running on nodes.|
    |cert-manager|Certificate management module. Used for unified management of internal and external certificates used by MEF.|
    |alarm-manager|Alarm management module. Used to manage alarms and events for MEF Edge and MEF Center.|

- MEF Edge is the edge management software that interfaces with MEF Center. MEF Edge primarily receives messages from MEF Center and collects and forwards relevant information to MEF Center, enabling functions such as software installation and upgrade, and full lifecycle management of containerized applications. At the same time, MEF has offline autonomy capabilities: when the link between the edge node where MEF Edge resides and the central node where MEF Center resides is interrupted, the inference service on the edge node is not interrupted; if the edge node restarts, the inference service can automatically recover after the edge node restart is complete.

    **Table 2** MEF Edge module description

    |Module|Module Function|
    |--|--|
    |edge-om|Main process module, including the upgrade module, etc.|
    |edge-main|Process module for interfacing MEF Edge and MEF Center.|
    |EdgeCore|The edge-side component of the open-source system KubeEdge. Responsible for container lifecycle management on edge nodes.|
    |Device-Plugin|Device discovery plugin for NPU (Ascend AI Processor).|

# Application Scenario<a name="ZH-CN_TOPIC_0000001674416014"></a>

The primary use cases of MEF include edge node management and containerized application management. By onboarding edge nodes from clusters into the MEF system, MEF enables unified management of edge node information, providing functions such as node onboarding, querying, modifying, and deleting node or node group information. As a fundamental feature of MEF, containerized application management handles the full lifecycle management of user applications. User applications are published as container images, and MEF manages these containerized application images, covering functions such as adding, deleting, modifying, and querying containerized applications.

MEF achieves cloud-edge synergy through the integration of MEF Edge with MEF Center, and externally interfaces with ISV business platforms via northbound interfaces to manage edge nodes and containerized applications.

**Figure 1** MEF integration mode<a name="fig193001916163619"></a>
![](../figures/MEF-integration-mode.png "MEF-integration-mode")

**Procedure of Connecting MEF Edge to MEF Center<a name="section1194622405510"></a>**

**Figure 2** Cloud-edge collaboration between MEF Edge and MEF Center<a name="fig1825153453711"></a>
![](../figures/cloud-edge-collaboration-between-MEF-Edge-and-MEF-Center.png "cloud-edge-collaboration-between-MEF-Edge-and-MEF-Center")

The cloud-edge collaboration procedure for integrating MEF Edge with MEF Center mainly includes: installing MEF, secondary development and integration of MEF, and managing edge nodes and containerized applications. Installing MEF is divided into the preparation and installation of MEF Center and MEF Edge. After users complete secondary development such as customization modifications, MEF is integrated with the developer platform, externally interfacing with the ISV business platform through the northbound interface, and internally implementing cloud-edge integration between MEF Center and MEF Edge. Management of edge nodes and containerized applications is then carried out through the ISV business platform.

# Supported Product Models and OSs<a name="ZH-CN_TOPIC_0000001674415942"></a>

**Table 1** Product list supported by the integration mode of MEF Edge with MEF Center

<table><thead align="left"><tr id="row919217382919"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.1"><p id="p1261719119133"><a name="p1261719119133"></a><a name="p1261719119133"></a>Installation Node</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.2"><p id="p1719213812914"><a name="p1719213812914"></a><a name="p1719213812914"></a>Software</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.3"><p id="p1319216381792"><a name="p1319216381792"></a><a name="p1319216381792"></a>Product Form</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.4"><p id="p16827128141310"><a name="p16827128141310"></a><a name="p16827128141310"></a>Software Architecture</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.6.1.5"><p id="p101929381791"><a name="p101929381791"></a><a name="p101929381791"></a>OS</p>
</th>
</tr>
</thead>
<tbody><tr id="row719216381197"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.1 "><p id="p861721181320"><a name="p861721181320"></a><a name="p861721181320"></a>Management Node</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.2 "><p id="p14192123816914"><a name="p14192123816914"></a><a name="p14192123816914"></a><span id="ph24772113106"><a name="ph24772113106"></a><a name="ph24772113106"></a>MEF Center</span></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="p10192113813919"><a name="p10192113813919"></a><a name="p10192113813919"></a>General-Purpose Server</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.4 "><p id="p98271884135"><a name="p98271884135"></a><a name="p98271884135"></a><span id="ph119331722101310"><a name="ph119331722101310"></a><a name="ph119331722101310"></a>AArch64</span> and <span id="ph1274682034217"><a name="ph1274682034217"></a><a name="ph1274682034217"></a>x86_64</span></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.5 "><p id="p144746191013"><a name="p144746191013"></a><a name="p144746191013"></a><span id="ph1052114212228"><a name="ph1052114212228"></a><a name="ph1052114212228"></a>Ubuntu</span> 20.04</p>
<p id="p121928381599"><a name="p121928381599"></a><a name="p121928381599"></a><span id="ph123381496136"><a name="ph123381496136"></a><a name="ph123381496136"></a>openEuler</span> 22.03</p>
</td>
</tr>
<tr id="row219203820914"><td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.6.1.1 "><p id="p1029253817131"><a name="p1029253817131"></a><a name="p1029253817131"></a>Compute Node</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.6.1.2 "><p id="p12192538898"><a name="p12192538898"></a><a name="p12192538898"></a><span id="ph1061942718105"><a name="ph1061942718105"></a><a name="ph1061942718105"></a>MEF Edge</span></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.3 "><p id="p14245463518"><a name="p14245463518"></a><a name="p14245463518"></a><span id="text6139155151210"><a name="text6139155151210"></a><a name="text6139155151210"></a>Atlas 200I A2 Acceleration Module</span></p>
<p id="p14579121816595"><a name="p14579121816595"></a><a name="p14579121816595"></a><span id="ph103373574589"><a name="ph103373574589"></a><a name="ph103373574589"></a>Atlas 200I DK A2 Developer Kit</span></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.4 "><p id="p827871281317"><a name="p827871281317"></a><a name="p827871281317"></a><span id="ph150212122159"><a name="ph150212122159"></a><a name="ph150212122159"></a>AArch64</span></p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.6.1.5 "><p id="p12072214101"><a name="p12072214101"></a><a name="p12072214101"></a><span id="ph9796114014252"><a name="ph9796114014252"></a><a name="ph9796114014252"></a>openEuler</span> 22.03</p>
<p id="p2066465711339"><a name="p2066465711339"></a><a name="p2066465711339"></a><span id="ph1428314031119"><a name="ph1428314031119"></a><a name="ph1428314031119"></a>Ubuntu</span> 22.04</p>
</td>
</tr>
<tr id="row11488658181115"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p168761453654"><a name="p168761453654"></a><a name="p168761453654"></a><span id="ph1485920011616"><a name="ph1485920011616"></a><a name="ph1485920011616"></a>Atlas 500 Pro Intelligent Edge Server (Model 3000)</span> (with <span id="ph616663925417"><a name="ph616663925417"></a><a name="ph616663925417"></a>Atlas 300I Pro Inference Card</span>)</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p73216165513"><a name="p73216165513"></a><a name="p73216165513"></a><span id="ph0951163552"><a name="ph0951163552"></a><a name="ph0951163552"></a>AArch64</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1785164564"><a name="p1785164564"></a><a name="p1785164564"></a><span id="ph4292409202"><a name="ph4292409202"></a><a name="ph4292409202"></a>openEuler</span> 22.03</p>
</td>
</tr>
</tbody>
</table>
