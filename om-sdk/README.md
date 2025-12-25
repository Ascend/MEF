# OM SDK

- [最新消息](#最新消息)
- [简介](#简介)
- [目录结构](#目录结构)
- [版本说明](#版本说明)
- [兼容性信息](#兼容性信息)
- [环境部署](#环境部署)
- [快速入门](#快速入门)
- [功能介绍](#功能介绍)
- [API参考](#API参考)
- [安全声明](#安全声明)
- [分支维护策略](#分支维护策略)
- [版本维护策略](#版本维护策略)
- [免责声明](#免责声明)
- [License](#License)
- [贡献声明](#贡献声明)
- [建议与交流](#建议与交流)

## 最新消息

- [2025.12.30]：版本发布

## 简介

OM SDK作为开发态组件，使能第三方合作伙伴基于昇腾AI处理器快速搭建智能边缘硬件管理平台，自定义构建设备运维系统，简化设备运维部署。

## 目录结构

关键目录如下，详细目录介绍参见[项目目录](docs/dir_structure.md)。

    om-sdk					        # 项目根目录
    ├── build				        # 构建相关目录
    ├── docs				        # 文档目录
    │   └── images				        # 图片目录
    └── src				                # 源码目录
        ├── om			                # OM SDK 组件代码
        └── om-web					# OM Web 前端组件代码

## 版本说明
OM SDK版本配套详情请参考：[版本配套说明](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/releasenote/edgereleasenote_0002.html)。

## 兼容性信息

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

## 环境部署

### 编译

1. 拉取om-sdk整体源码，例如放在/home目录下。
2. 进入/home/om-sdk/build目录
    ```shell
    cd /home/om-sdk/build
    ```
3. 修改组件版本配置文件service_config.ini中om-sdk-version字段值为所需编译版本，默认值如下：
    ```
    om-sdk-version=7.3.0
    ```
4. 执行以下命令，执行构建脚本：
    ```shell
    dos2unix *.sh && chmod +x *.sh
    ./build_all.sh
    ```
5. 执行完成后，可在/home/om-sdk/output目录下获取编译完成的软件包。

### 安装

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

## 快速入门

安装OM SDK后，可登录边缘管理系统进行基础操作、系统管理和数据配置。具体内容请参考昇腾社区《OM SDK使用》文档，"[新手入门](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkug/help_003.html)"章节。

- 用户登录
- 基础操作
- 首页
- 管理
- 设置

## 功能介绍

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

## API参考

API请参考"[RESTful接口](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkdg/omsdk_api01_0001.html)"和"[云边协同接口](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/omsdk/omsdkdg/omsdk_api02_0002.html)"。

## 安全声明

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
- 正式版本请参考正式release版本: https://gitcode.com/Ascend/om-sdk/releases。

## License

OM SDK以Mulan PSL v2许可证许可，对应许可证文本可查阅[LICENSE](LICENSE.md)。

## 贡献声明

1. 提交错误报告：如果您在OM SDK中发现了一个不存在安全问题的漏洞，请在OM SDK仓库中的Issues中搜索，以防该漏洞被重复提交，如果找不到漏洞可以创建一个新的Issues。如果发现了一个安全问题请不要将其公开，请参阅安全问题处理方式。提交错误报告时应该包含完整信息。
2. 安全问题处理：本项目中对安全问题处理的形式，请通过邮箱通知项目核心人员确认编辑。
3. 解决现有问题：通过查看仓库的Issues列表可以发现需要处理的问题信息, 可以尝试解决其中的某个问题。
4. 如何提出新功能：请使用Issues的Feature标签进行标记，我们会定期处理和确认开发。
5. 开始贡献：
    - Fork本项目的仓库
    - Clone到本地
    - 创建开发分支
    - 本地自测，提交前请通过所有的单元测试，包括为您要解决的问题新增的单元测试
    - 提交代码
    - 新建Pull Request
    - 代码检视，您需要根据评审意见修改代码，并重新提交更新。此流程可能涉及多轮迭代
    - 当您的PR获得足够数量的检视者批准后，Committer会进行最终审核
    - 审核和测试通过后，CI会将您的PR合并入到项目的主干分支

## 建议与交流

欢迎大家为社区做贡献。如果有任何疑问或建议，请提交[issues](https://gitcode.com/Ascend/om-sdk/issues)，我们会尽快回复。感谢您的支持。
