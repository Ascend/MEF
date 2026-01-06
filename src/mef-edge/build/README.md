# MEF Edge构建说明

## 说明
MEF Edge支持MEF_Edge_SDK和MEF_Edge两种构建形态，对应构建包分别为Ascend-mindxedge-mefedgesdk_{version}\_linux-aarch64.zip和Ascend-mindxedge-mefedge_{version}_linux-aarch64.zip。
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

如果希望单独构建MEF Edge软件包，可参考本章节。

### 依赖准备
执行MEF Edge编译前，请保证环境上存在以下依赖：
golang、gcc、zip、dos2unix、git、autoconf、automake、libtool、libc-dev、cmake、build-essential，其中golang版本为1.22.1。

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
