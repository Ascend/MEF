# MindEdge

- [最新消息](#最新消息)
- [简介](#简介)
- [目录结构](#目录结构)
- [版本说明](#版本说明)
- [MindEdge Framework (MEF)](#mindedge-framework-mef)
    - [兼容性信息](#mef-兼容性信息)
    - [环境部署](#mef-环境部署)
    - [快速入门](#mef-快速入门)
    - [功能介绍](#mef-功能介绍)
    - [API参考](#mef-api参考)
    - [FAQ](#mef-faq)
    - [安全声明](#mef-安全声明)
- [OM SDK](#om-sdk)
    - [兼容性信息](#omsdk-兼容性信息)
    - [环境部署](#omsdk-环境部署)
    - [快速入门](#omsdk-快速入门)
    - [功能介绍](#omsdk-功能介绍)
    - [API参考](#omsdk-api参考)
    - [安全声明](#omsdk-安全声明)
- [分支维护策略](#分支维护策略)
- [版本维护策略](#版本维护策略)
- [免责声明](#免责声明)
- [License](#License)
- [建议与交流](#建议与交流)

## 最新消息

- [2025.12.15]：MindEdge版本发布

## 简介

MindEdge提供边缘AI业务基础组件管理和边缘AI业务容器的全生命周期管理能力，同时提供节点看管、日志采集等统一运维能力和严格的安全可信保障，使能客户快速构建边缘AI业务。包含MindEdge Framework (MEF)和OM SDK两大组件。

- MindEdge Framework (MEF)作为被集成的轻量化端边云协同使能框架，提供边缘节点管理、边缘推理应用生命周期管理等边云协同能力。
- OM SDK作为开发态组件，使能第三方合作伙伴基于昇腾AI处理器快速搭建智能边缘硬件管理平台，自定义构建设备运维系统，简化设备运维部署。

![架构图](docs/images/framework.png)

## 目录结构

关键目录如下，详细目录介绍参见[项目目录](docs/dir_structure.md)。

	mind-edge				        # 项目根目录
    ├── build				        # 构建相关目录
    ├── common-utils				# 公共工具库
    ├── device-plugin				# 设备插件组件
    ├── mef-center					# MEFCenter 中心组件代码
    ├── mef-edge					# MEFEdge 边缘组件代码
    ├── om-sdk					# OM SDK 组件代码
    └── om-web					# OM Web 前端组件代码

## 版本说明
MindEdge版本配套详情请参考：[版本配套说明](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/releasenote/edgereleasenote_0002.html)。

## <span id="mindedge-framework-mef">MindEdge Framework (MEF)</span>

### <span id="mef-兼容性信息">兼容性信息</span>

表1 MEF支持的产品形态和OS清单表
<table>
    <tr>
        <th>安装节点</th>
        <th>软件</th>
        <th>产品形态</th>
        <th>软件架构</th>
        <th>操作系统</th>
    </tr>
    <tr>
        <td>管理节点</td>
        <td>MEF Center</td>
        <td>通用服务器</td>
        <td>AArch64<br>x86_64</td>
        <td>Ubuntu 20.04<br>OpenEuler 22.03</td>
    </tr>
    <tr>
        <td rowspan="2">计算节点</td>
        <td rowspan="2">MEF Edge</td>
        <td>Atlas 200I A2 加速模块<br>Atlas 200I DK A2 开发者套件</td>
        <td>AArch64</td>
        <td>OpenEuler 22.03<br>Ubuntu 22.04</td>
    </tr>
    <tr>
        <td>Atlas 500 Pro 智能边缘服务器（型号 3000）（插Atlas 300I Pro 推理卡A300I Pro 推理卡）</td>
        <td>AArch64</td>
        <td>OpenEuler 22.03</td>
    </tr>
</table>

### <span id="mef-环境部署">环境部署</span>

在安装和使用前，用户需要了解安装须知、环境准备，具体内容请参考昇腾社区《MindEdge Framework用户指南》文档，"[安装MindEdge Framework](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0006.html)”章节。

![MEF安装流程图](docs/images/mef_install.png)

- 安装部署MEF Center
    - 以root用户登录准备安装MEF Center的设备环境
    - 将软件包上传至设备任意路径下（建议该目录权限为root且其他用户不可写）
        - 执行以下命令，解压软件包
          ```shell
          tar -zxvf Ascend-mindxedge-mefcenter_{version}_linux-{arch}.tar.gz
          ```
    - 安装MEF Center
        - 进入安装路径。
          ```shell
          cd 软件包上传路径/installer
          ```
        - 执行以下命令，安装MEF Center
          ```shell
          ./install.sh
          ```
        - 回显示例如下，表示MEF Center安装成功
          ```shell
          install MEF center success
          ```
    - 启动MEF Center
        - 执行以下命令，进入MEF Center所在路径
          ```shell
          cd 安装路径/MEF-Center/mef-center
          ```
        - 执行以下命令，启动MEF Center所有模块
          ```shell
          ./run.sh start
          ```
        - 回显示例如下，表示操作执行成功
          ```shell
          start all component successful
          ```

- 安装部署MEF Edge
    - 以root用户登录准备安装MEF Edge的设备环境
    - 将获取到的软件包上传至设备任意路径下（该目录须为root属主，且目录权限为属组及其他用户不可写）
        - 执行以下命令，解压软件包
          ```shell
          tar -zxvf Ascend-mindxedge-mefedgesdk_{version}_linux-aarch64.tar.gz
          ```
    - 安装MEF Edge
        - MEF Edge软件的安装可以选择默认安装和指定路径安装
            - 默认安装
              ```shell
              ./install.sh
              ```
            - 指定路径安装
              ```shell
              ./install.sh --install_dir=安装路径 --log_dir=日志路径 --log_backup_dir=日志转储路径
              ```
        - 回显示例如下，表示MEF Edge安装成功
          ```shell
          install MEFEdge success
          ```
    - 启动MEF Edge
        - 执行以下命令，进入run.sh所在路径
          ```shell
          cd 安装目录/MEFEdge/software/
          ```
        - 执行以下命令，启动MEF Edge
          ```shell
          ./run.sh start
          ```
        - 回显示例如下，表示启动命令执行成功
          ```shell
          Execute [start] command success!
          ```

### <span id="mef-快速入门">快速入门</span>

云边协同的应用流程主要包括安装MEF、二次开发集成和管理边缘节点及容器应用三部分，具体内容请参考昇腾社区《MindEdge Framework用户指南》文档，“[使用指导](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0025.html)”章节。

![MEF应用流程图](docs/images/mef_application.png)

### <span id="mef-功能介绍">功能介绍</span>

可通过MEF Edge和MEF Center进行边云协同管理，用户可通过二次开发，对接ISV（Independent Software Vendor）业务平台，集成所需功能。
- MEF Edge部署在智能边缘设备上，负责与中心网管对接，完成智能推理业务（容器应用）的部署和管理，为算法应用提供服务。
- MEF Center部署在通用服务器上，负责对边缘节点实现批量管理、业务部署和系统监测。

表2 MEF组件功能介绍

| 功能类型                                                                                               | 详细功能介绍                                                                           |
|:---------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------|
| [节点管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0007.html)   | <ul><li>支持对节点组进行创建、查询、修改和删除操作。</li><li>支持对节点进行纳管、添加、修改、删除、查询等操作。</li></ul>       |
| [容器应用管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0029.html) | <ul><li>支持对容器应用进行创建、查询、更新、删除等操作。</li><li>支持容器应用部署到节点组、从节点组卸载、从单个节点上卸载。</li></ol> |
| [日志收集](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0041.html)   | <ul><li>支持收集并导出MEF Edge的日志，实现MEF Edge的日志排查，设备状态监测。</li></ol>                     |
| [配置管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0052.html)   | <ul><li>支持导入、查询、删除根证书。</li><li>支持导入吊销列表。</li><li>支持配置镜像下载信息等。</li></ol>          |
| [告警管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0046.html)   | <ul><li>支持查询告警或事件信息。</li></ol>                                                   |
| [软件升级](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0061.html)   | <ul><li>支持通过MEF Center软件升级接口进行MEF Edge的在线升级、同版本升级和版本回退。</li></ol>                |
| [北向接口](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0000.html)   | <ul><li>提供APIG服务，实现接受外部访问、对北向接口限流及转发功能。</li></ol>                                |

### <span id="mef-api参考">API参考</span>

API参考详见：[接口参考](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0000.html)。

### <span id="mef-faq">FAQ</span>

相关FAQ详见：[FAQ](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0108.html)。

### <span id="mef-安全声明">安全声明</span>

- 请参考[安全加固建议](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0088.html)对系统进行安全加固。
- 安全加固建议中的安全加固措施为基本的加固建议项。用户应根据自身业务，重新审视整个系统的网络安全加固措施。用户应按照所在组织的安全策略进行相关配置，包括并不局限于软件版本、口令复杂度要求、安全配置（协议、加密套件、密钥长度等），权限配置、防火墙设置等。必要时可参考业界优秀加固方案和安全专家的建议。
- 安全加固涉及主机加固和容器应用加固，防止可能出现的安全隐患，用于保障设备和容器应用的安全，请用户根据实际需要进行安全加固操作。
- 外部下载的软件代码或程序可能存在风险，功能的安全性需由用户保证。
- 通信矩阵详见：[通信矩阵](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/commumatrix/Communication_matrix_0001.html)
- 公网地址详见：[公网地址](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0115.html)
- 环境变量说明详见：[环境变量说明](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0116.html)
- 用户信息列表详见：[用户信息列表](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0117.html)


## <span id="om-sdk">OM SDK</span>

### <span id="omsdk-兼容性信息">兼容性信息</span>

表3 OM SDK支持的产品和产品所支持的操作系统
<table>
    <tr>
        <th>产品名称</th>
        <th>操作系统</th>
    </tr>
    <tr>
        <td>Atlas 200I A2 加速模块（RC模式）</td>
        <td rowspan="2">OpenEuler 22.03<br>Ubuntu 22.04</td>
    </tr>
    <tr>
        <td>Atlas 200I DK A2 开发者套件</td>
    </tr>
</table>

### <span id="omsdk-环境部署">环境部署</span>

在安装和使用前，用户需要了解安装须知、环境准备，具体内容请参考昇腾社区《OM SDK使用》文档，"[安装部署](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkug/omsdkag_0004.html)"章节。

- 获取软件包
- 准备安装环境
- 通过命令行安装
    - 将软件包上传到环境任意目录下（如“/home”）
    - 在软件包目录下，执行以下命令，创建临时目录om_install
       ```shell
       mkdir om_install
       ```
    - 执行以下命令，解压tar.gz软件包
       ```shell
       tar -zxf om-sdk.tar.gz -C om_install
       ```
    - 执行以下命令，为安装脚本添加可执行权限
       ```shell
       chmod +x om_install/install.sh
       ```
    - 执行以下命令，安装软件包
       ```shell
       om_install/install.sh
       ```
    - 回显示例如下，表示安装完成
       ```shell
       check install environment success
       prepare service file success
       executing install success
       start service success
       Install MindXOM success, MindXOM service is ready.
       ```  

### <span id="omsdk-快速入门">快速入门</span>

安装OM SDK后，可登录边缘管理系统进行基础操作、系统管理和数据配置。具体内容请参考昇腾社区《OM SDK使用》文档，"[新手入门](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkug/help_003.html)"章节。

- 用户登录
- 基础操作
- 首页
- 管理
- 设置

### <span id="omsdk-功能介绍">功能介绍</span>

边缘管理系统支持对边缘设备进行初始化配置、硬件监测、软件安装、系统运维等功能；同时还支持与SmartKit软件、华为FusionDirector管理软件对接，实现集中式运维管理。具体内容请参考昇腾社区《OM SDK使用》文档，"[Web功能介绍](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkug/help_001.html)"章节。

表4 边缘管理系统功能介绍

| 功能类型 | 详细功能介绍                                                                                                                            |
|:-----|:----------------------------------------------------------------------------------------------------------------------------------|
| 硬件管理 | <ul><li>硬件信息查询</li><li>硬件故障检测</li></ul>                                                                                           |
| 软件管理 | <ul><li>系统OS、驱动固件升级</li><li>软件信息查询</li><li>一键式开局和免软调上线</li><li>OM SDK的安装和升级</li></ol>                                             |
| 时间管理 | <ul><li>系统时区、系统时间配置</li><li>支持NTP从服务器同步时间</li></ol>                                                                               |
| 网络管理 | <ul><li>支持ETH、WiFi、LTE等多种网络设备配置</li><li>支持手动配置系统网口的IP、端口、VLAN、网关、DNS</li><li>支持DHCP从Server端获取系统IP</li></ol>                       |
| 存储管理 | <ul><li>支持查询和配置本地存储</li><li>查询系统分区、存储容量和分区健康状态</li><li>支持配置、查询NFS存储系统，如NFS挂载，容量显示，连接健康状态</li></ol>                                |
| 用户管理 | <ul><li>支持密码有效期，登录规则、弱口令设置、查询，支持用户密码修改</li><li>支持用户可定制化的安全策略，支持客户可信根导入</li><li>支持Web证书导入、查询和有效期检查</li></ol>                       |
| 系统监测 | <ul><li>支持告警上报，告警屏蔽、历史告警查询、支持当前告警显示</li><li>支持客户增量设备、关键进程的告警集成显示、管理</li><li>系统支持安全日志、操作日志、运行日志、黑匣子记录，支持日志收集、查询、远程syslog</li></ol> |
| 北向接口 | <ul><li>系统功能支持FusionDirector集中纳管协议，支持RESTful开放接口，RESTful满足服务器北向接口标准</li></ol>                                                     |

### <span id="omsdk-api参考">API参考</span>

API请参考"[RESTful接口](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkdg/omsdk_api01_0001.html)"和"[云边协同接口](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkdg/omsdk_api02_0002.html)"。

### <span id="omsdk-安全声明">安全声明</span>

- 请参考[安全配置和加固](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkug/omsdkag_0019.html)对系统进行安全加固。
- 安全加固建议中的安全加固措施为基本的加固建议项。用户应根据自身业务，重新审视整个系统的网络安全加固措施。
- 外部下载的软件代码或程序可能存在风险，功能的安全性需由用户保证。
- 公网地址详见：[公网地址](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkug/omsdkag_0035.html)
- 用户信息列表详见：[用户信息列表](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkug/omsdkag_0036.html)


## 分支维护策略

版本分支的维护阶段如下：

| 状态          | 时间     | 说明                                                      |
|-------------|--------|---------------------------------------------------------|
| 计划          | 1-3个月  | 计划特性                                                    |
| 开发          | 3个月    | 开发新特性并修复问题，定期发布新版本                                      | 
| 维护          | 3-12个月 | 常规分支维护3个月，长期支持分支维护12个月。对重大BUG进行修复，不合入新特性，并视BUG的影响发布补丁版本 | 
| 生命周期终止（EOL） | N/A    | 分支不再接受任何修改                                              |

## 版本维护策略

| 版本       | 维护策略 | 当前状态 | 发布日期       | 后续状态                 | EOL日期      |
|----------|------|------|------------|----------------------|------------|
| master   | 长期支持 | 开发   | 在研分支，不发布   | 2025-10-27           | -          |
| v7.3.0   | 长期支持 | 开发   | 在研分支，未发布   | 2025-10-27           | -          |

## 免责声明

- 本代码仓库中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布之前，这些分支不应被用于任何生产环境或依赖关键业务的项目中。请务必仅使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考MindEdge正式release版本: https://gitcode.com/Ascend/mind-edge/releases。

## License

MindEdge以Mulan PSL v2许可证许可，对应许可证文本可查阅[LICENSE](LICENSE.md)。

## 建议与交流

欢迎大家为社区做贡献。如果有任何疑问或建议，请提交[issues](https://gitcode.com/Ascend/mind-edge/issues)，我们会尽快回复。感谢您的支持。
