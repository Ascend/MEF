# MEF Edge构建说明

## 说明
MEF Edge支持MEF_Edge_SDK和MEF_Edge两种构建形态，对应构建包分别为`Ascend-mindxedge-mefedgesdk_{version}_linux-aarch64.zip`和`Ascend-mindxedge-mefedge_{version}_linux-aarch64.zip`。
两种构建形态分别用于支持不同设备形态，如下表所示。

表1 MEF Edge支持的产品形态表

<table>
    <tr>
        <th>软件包</th>
        <th>适配的产品形态</th>
        <th>软件架构</th>
    </tr>
    <tr>
        <td rowspan="2">Ascend-mindxedge-mefedgesdk_{version}_linux-aarch64.zip</td>
        <td>Atlas 200I A2 加速模块<br>Atlas 200I DK A2 开发者套件</td>
        <td>AArch64</td>
    </tr>
    <tr>
        <td>Atlas 500 Pro 智能边缘服务器（型号 3000）（插Atlas 300I Pro 推理卡A300I Pro 推理卡）</td>
        <td>AArch64</td>
    </tr>
    <tr>
        <td>Ascend-mindxedge-mefedge_{version}_linux-aarch64.zip</td>
        <td>Atlas 500 A2智能小站</td>
        <td>AArch64</td>
    </tr>
</table>

## 构建

如果希望单独构建MEF Edge软件包，可参考本章节。下面以Ubuntu20.04系统为例，介绍如何通过源码编译生成MEF软件包。

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

### 编译

1. 拉取MEF整体源码，例如放在/home目录下。
2. 修改组件版本配置文件/home/MEF/build/service_config.ini中mef-version字段值为所需编译版本，默认值如下：
   ```
   mef-version=7.3.0
   ```
3. 进入MEF Edge构建子目录：/home/MEF/src/mef-edge/build
   ```shell
   cd /home/MEF/src/mef-edge/build
   ```
4. 执行以下命令，执行构建依赖准备脚本：
   ```shell
   dos2unix *.sh && chmod +x *.sh
   ./prepare_dependency.sh
   ```
5. 执行以下命令，执行构建脚本：
   ```shell
   ./build.sh -p <产品名称>
   ```
   其中参数-p用于指定产品名称，用于构建MEF_Edge_SDK或MEF_Edge软件包时指定，示例如下：
   ```shell
   # 构建MEF_Edge_SDK包
   ./build.sh -p MEF_Edge_SDK
   # 构建MEF_Edge包
   ./build.sh -p MEF_Edge
   ```
6. 执行完成后，可在/home/MEF/src/mef-edge/output目录下获取编译完成的软件包，注意：MEF Edge仅支持AArch64架构，请在AArch64架构下编译MEF Edge软件包。

### 注意事项
- 由于软件构建过程中使用了glibc 2.34源码编译，建议使用Ubuntu 20.04系统进行编译构建，避免系统上的glibc版本过高导致的不兼容问题。
- 如果在编译过程中遇到问题，请检查错误日志并确保所有依赖库和工具都已正确安装。

## 测试用例

MEF Edge测试用例执行方法参考如下，注：测试用例需要在x86_64架构环境下执行，由于部分测试用例使用了gomonkey进行运行时函数打桩，该技术依赖底层架构相关的汇编指令和GO编译器行为，需要在x86_64架构运行。

1. 执行测试用例前，请参考[依赖准备](#依赖准备)章节进行测试环境准备
2. 安装依赖用于统计测试覆盖率和生成可视化报告，若最新版本依赖与Golang版本不兼容，可自行安装兼容的版本
   ```shell
   go install github.com/axw/gocov/gocov@latest
   go install github.com/matm/gocov-html/cmd/gocov-html@latest
   go install gotest.tools/gotestsum@latest
   export PATH=$PATH:$(go env GOPATH)/bin
   ```
3. 执行MEF Edge的测试用例
   ```shell
   cd /home/MEF/src/mef-edge/build
   dos2unix *.sh && chmod +x *.sh
   bash prepare_dependency.sh
   # 执行MEF_Edge_SDK对应测试用例
   bash test.sh MEF_Edge_SDK
   # 执行MEF_Edge对应测试用例
   bash test.sh MEFEdge_A500
   ```