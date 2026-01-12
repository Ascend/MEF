# MEF

- [最新消息](#最新消息)
- [简介](#简介)
- [目录结构](#目录结构)
- [版本说明](#版本说明)
- [兼容性信息](#兼容性信息)
- [环境部署](#环境部署)
- [编译流程](#编译流程)
- [测试用例](#测试用例)
- [快速入门](#快速入门)
- [功能介绍](#功能介绍)
- [API参考](#API参考)
- [FAQ](#FAQ)
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

MEF是一款定位为被集成的轻量化端边云协同使能框架。用于智能边缘设备使能，提供边缘节点管理、边缘推理应用生命周期管理等边云协同能力。可通过MEF
Edge和MEF Center进行边云协同管理，用户可通过二次开发，对接ISV（Independent Software Vendor）业务平台，集成所需功能。

- MEF Edge部署在智能边缘设备上，负责与中心网管对接，完成智能推理业务（容器应用）的部署和管理，为算法应用提供服务。
- MEF Center部署在通用服务器上，负责对边缘节点实现批量管理、业务部署和系统监测。

## 目录结构

关键目录如下，详细目录介绍参见[项目目录](docs/dir_structure.md)。

    MEF					        # 项目根目录
    ├── build				        # 构建相关目录
    ├── docs				        # 文档目录
    │   └── images				        # 图片目录
    └── src				                # 源码目录
        ├── common-utils				# 公共工具库
        ├── device-plugin				# 设备插件组件
        ├── mef-center				# MEFCenter 中心组件代码
        └── mef-edge				# MEFEdge 边缘组件代码

## 版本说明

MEF版本配套详情请参考：[版本配套说明](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/releasenote/edgereleasenote_0002.html)。

## 兼容性信息

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

## 环境部署

### 获取软件包

可通过本项目获取正式版本软件包，或自行编译构建两种方式获取软件包。
- 正式版本软件包获取: https://gitcode.com/Ascend/MEF/releases
- 自行编译构建软件包请参考[编译流程](#编译流程)章节

### 安装

在安装和使用前，用户需要了解安装须知、进行安装环境准备，具体内容请参考昇腾社区文档，“[安装MEF](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0006.html)”章节。

![MEF安装流程图](docs/images/mef_install.png)

- 安装部署MEF Center
    - 以root用户登录准备安装MEF Center的设备环境
    - 将软件包上传至设备任意路径下（建议该目录权限为root且其他用户不可写）
        - 执行以下命令，解压软件包
          ```shell
          unzip Ascend-mindxedge-mefcenter_{version}_linux-{arch}.zip
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
          unzip Ascend-mindxedge-mefedgesdk_{version}_linux-aarch64.zip
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

## 编译流程

本节以Ubuntu20.04系统为例，介绍如何通过源码编译生成MEF软件包。

### 依赖准备

- 执行MEF编译前，需保证系统上安装了必要的编译工具和依赖库，参考安装命令如下：
  ```shell
  apt-get update
  apt-get -y install texinfo gawk libffi-dev zlib1g-dev libssl-dev openssl sqlite3 libsqlite3-dev libnuma-dev numactl libpcre2-dev bison flex build-essential automake autoconf libtool rpm dos2unix libc-dev lcov pkg-config sudo tar git wget unzip zip docker.io python-is-python3 iputils-ping
  ```
- 除上述工具和依赖外，还需安装golang、cmake，版本要求如下，建议通过源码安装。

表2 依赖版本要求

| 依赖名称        | 版本建议      | 获取建议 |
|:------------|:----------| :--- |
| Golang | 1.22.1    | 建议通过获取源码包编译安装。 |
| CMake  | 3.16.5及以上 | 建议通过获取源码包编译安装。 |

### 编译MEF

1. 拉取MEF整体源码，例如放在/home目录下。
2. 进入/home/MEF/build目录
    ```shell
    cd /home/MEF/build
    ```
3. 修改组件版本配置文件service_config.ini中mef-version字段值为所需编译版本，默认值如下：
    ```
    mef-version=7.3.0
    ```
4. 执行以下命令，执行构建脚本：
    ```shell
    dos2unix *.sh && chmod +x *.sh
    ./build_all.sh
    ```
5. 执行完成后，可在/home/MEF/output目录下获取编译完成的软件包，注意：根据MEF Center和MEF Edge对不同架构的支持情况，AArch64架构下将编译MEF Center和MEF Edge软件包，x86_64架构下将仅编译MEF Center软件包。

### 注意事项
- 由于软件支持的运行操作系统包括Ubuntu 20.04，且软件构建过程中使用了glibc 2.34源码编译，建议使用Ubuntu 20.04系统进行编译构建，避免系统上的glibc版本过高导致的不兼容问题。
- 如果在编译过程中遇到问题，请检查错误日志并确保所有依赖库和工具都已正确安装。

## 测试用例

测试用例执行方法参考如下，注：测试用例需要在x86_64架构环境下执行，由于部分测试用例使用了gomonkey进行运行时函数打桩，该技术依赖底层架构相关的汇编指令和GO编译器行为，需要在x86_64架构运行。

1. 执行测试用例前，请参考[依赖准备](#依赖准备)章节进行测试环境准备
2. 安装依赖用于统计测试覆盖率和生成可视化报告，若最新版本依赖与Golang版本不兼容，可自行安装兼容的版本
   ```shell
   go install github.com/axw/gocov/gocov@latest
   go install github.com/matm/gocov-html/cmd/gocov-html@latest
   go install gotest.tools/gotestsum@latest
   export PATH=$PATH:$(go env GOPATH)/bin
   ```
3. 本项目下包含多个模块，各个模块的测试用例可分别执行，执行参考如下：
- 执行common-utils模块下测试用例
  ```shell
  cd /home/MEF/src/common-utils/build
  dos2unix *.sh && chmod +x *.sh
  bash test.sh
  ```
- 执行device-plugin模块下测试用例
  ```shell
  cd /home/MEF/src/device-plugin/build
  dos2unix *.sh && chmod +x *.sh
  bash test.sh
  ```
- 执行mef-center模块下测试用例
  ```shell
  cd /home/MEF/src/mef-center/build
  dos2unix *.sh && chmod +x *.sh
  bash test.sh
  ```
- 执行mef-edge模块下测试用例
  ```shell
  cd /home/MEF/src/mef-edge/build
  dos2unix *.sh && chmod +x *.sh
  bash prepare_dependency.sh
  bash test.sh MEF_Edge_SDK
  ```

## 快速入门

云边协同的应用流程主要包括安装MEF、二次开发集成和管理边缘节点及容器应用三部分，具体内容请参考昇腾社区文档，“[使用指导](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0025.html)”章节。

![MEF应用流程图](docs/images/mef_application.png)

## 功能介绍

可通过MEF Edge和MEF Center进行边云协同管理，用户可通过二次开发，对接ISV（Independent Software Vendor）业务平台，集成所需功能。

- MEF Edge部署在智能边缘设备上，负责与中心网管对接，完成智能推理业务（容器应用）的部署和管理，为算法应用提供服务。
- MEF Center部署在通用服务器上，负责对边缘节点实现批量管理、业务部署和系统监测。

表3 MEF组件功能介绍

| 功能类型                                                                                               | 详细功能介绍                                                                           |
|:---------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------|
| [节点管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0007.html)   | <ul><li>支持对节点组进行创建、查询、修改和删除操作。</li><li>支持对节点进行纳管、添加、修改、删除、查询等操作。</li></ul>       |
| [容器应用管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0029.html) | <ul><li>支持对容器应用进行创建、查询、更新、删除等操作。</li><li>支持容器应用部署到节点组、从节点组卸载、从单个节点上卸载。</li></ol> |
| [日志收集](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0041.html)   | <ul><li>支持收集并导出MEF Edge的日志，实现MEF Edge的日志排查，设备状态监测。</li></ol>                     |
| [配置管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0052.html)   | <ul><li>支持导入、查询、删除根证书。</li><li>支持导入吊销列表。</li><li>支持配置镜像下载信息等。</li></ol>          |
| [告警管理](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0046.html)   | <ul><li>支持查询告警或事件信息。</li></ol>                                                   |
| [软件升级](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0061.html)   | <ul><li>支持通过MEF Center软件升级接口进行MEF Edge的在线升级、同版本升级和版本回退。</li></ol>                |
| [北向接口](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0000.html)   | <ul><li>提供APIG服务，实现接受外部访问、对北向接口限流及转发功能。</li></ol>                                |

## API参考

API参考详见：[接口参考](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefapiref_0000.html)。

## FAQ

相关FAQ详见：[FAQ](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0108.html)。

## 安全声明

- 请参考[安全加固建议](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0088.html)对系统进行安全加固。
- 安全加固建议中的安全加固措施为基本的加固建议项。用户应根据自身业务，重新审视整个系统的网络安全加固措施。用户应按照所在组织的安全策略进行相关配置，包括并不局限于软件版本、口令复杂度要求、安全配置（协议、加密套件、密钥长度等），权限配置、防火墙设置等。必要时可参考业界优秀加固方案和安全专家的建议。
- 安全加固涉及主机加固和容器应用加固，防止可能出现的安全隐患，用于保障设备和容器应用的安全，请用户根据实际需要进行安全加固操作。
- 外部下载的软件代码或程序可能存在风险，功能的安全性需由用户保证。
- 通信矩阵详见：[通信矩阵](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/commumatrix/Communication_matrix_0001.html)
- 公网地址详见：[公网地址](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0115.html)
- 环境变量说明详见：[环境变量说明](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0116.html)
- 用户信息列表详见：[用户信息列表](https://www.hiascend.com/document/detail/zh/mindedge/72rc1/mef/mefug/mefug_0117.html)

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

## 免责声明

- 本代码仓库中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布之前，这些分支不应被用于任何生产环境或依赖关键业务的项目中。请务必仅使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考正式release版本: https://gitcode.com/Ascend/MEF/releases

## License

MEF以Mulan PSL v2许可证许可，对应许可证文本可查阅[LICENSE](LICENSE.md)。

## 贡献声明

1. 提交错误报告：如果您在MEF中发现了一个不存在安全问题的漏洞，请在MEF仓库中的Issues中搜索，以防该漏洞被重复提交，如果找不到漏洞可以创建一个新的Issues。如果发现了一个安全问题请不要将其公开，请参阅安全问题处理方式。提交错误报告时应该包含完整信息。
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

欢迎大家为社区做贡献。如果有任何疑问或建议，请提交[issues](https://gitcode.com/Ascend/MEF/issues)，我们会尽快回复。感谢您的支持。
