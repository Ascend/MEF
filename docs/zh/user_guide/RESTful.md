# 接口参考<a name="ZH-CN_TOPIC_0000001674256246"></a>

用户通过对MEF Center提供的RESTful接口进行调用，可实现MEF相关功能的使用。

## 概述<a name="ZH-CN_TOPIC_0000001527201136"></a>

本文档详细的描述了MEF Center  RESTful接口的相关说明。指导用户通过调用RESTful接口实现对MEF功能的使用。

Redfish是一种基于HTTPS服务的管理标准，利用RESTful接口实现设备管理。每个HTTPS操作都以UTF-8编码的JSON的形式，提交或返回一个资源。就像Web应用程序向浏览器返回HTML一样，RESTful接口会通过同样的传输机制（HTTPS），以JSON的形式向客户端返回数据，本文档中通过JSON格式传入的参数名不区分大小写。

> [!NOTE] 说明  
> MEF提供的所有RESTful接口仅支持串行调用。

**返回状态码<a name="zh-cn_topic_0000001082477678_zh-cn_topic_0178823233_section37953311"></a>**

**表 1**  状态码说明

|状态码|说明|
|--|--|
|200|请求成功。|
|400|请求非法，客户端侧发生错误并返回错误消息。|
|401|未认证，请求要求用户的身份认证。|
|404|访问请求资源不存在。|
|405|请求方法非法。如使用错误的方法类型进行请求。|
|423|当前资源被锁定。|
|429|发送请求过多，超出频次限制。|
|499|客户端断开了连接。|
|503|由于服务器维护或者过载，服务器当前无法处理请求。|

**Base URL说明<a name="section20778152719278"></a>**

所有接口调用均基于统一的Base URL格式：https://{ip}:{port}，具体的接口路径请参见各接口的详细说明。

**表 2**  Base URL组成说明

|名称|是否必选|说明|取值要求|
|--|--|--|--|
|ip|必选|登录设备的IP地址|IPv4地址。|
|port|必选|调用具体模块接口的端口号|-|

**请求头参数<a name="section13745311172616"></a>**

本文档涉及到的请求头参数说明如下。

**表 3**  请求头

|参数名|是否必选|参数说明|取值要求|
|--|--|--|--|
|Content-Type|必选|媒体类型|字符串，取值为application/json，在请求头中携带。|

## 功能介绍<a name="ZH-CN_TOPIC_0000001526721304"></a>

当前MEF Center包含的模块有nginx-manager模块、edge-manager模块、cert-manager和alarm-manager模块，接口参考按照各模块提供的功能进行接口介绍。

- nginx-manager模块提供APIG服务，实现接受外部访问、对北向接口限流及转发功能。
- edge-manager模块包含节点管理接口和容器应用管理接口，实现对边缘节点及节点上运行的容器应用的管理。
    - 支持对节点组进行创建、查询、修改和删除操作。
    - 支持对节点进行纳管、添加、删除、查询等操作。
    - 支持对容器应用进行创建、查询、更新、删除和部署等操作。

- cert-manager模块包含配置接口，支持导入、查询、删除根证书，并进行证书签发。
- alarm-manager模块提供对告警或事件信息的查询。

## 组件版本接口<a name="ZH-CN_TOPIC_0000001526881264"></a>

### 查询edge-manager版本<a id="ZH-CN_TOPIC_0000001577280961"></a>

**命令功能<a name="section135251624204320"></a>**

查询edge-manager版本。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/version**

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/version
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": "7.0.RC1"
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 1**  操作输出说明

|字段|类型|说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|字符串|当前edge-manager的版本号|

## 节点管理接口<a name="ZH-CN_TOPIC_0000001526881140"></a>

### 节点管理接口介绍<a id="ZH-CN_TOPIC_0000001577280885"></a>

加入同一个组的所有节点，可以对其进行批量的容器应用的管理，包括部署、更新和卸载容器应用等。节点和节点组以ID为唯一标识符，ID自动生成。节点ID、NodeName、UniqueName、SerialNumber等字段不能重复。

**条件约束<a name="section1127451245510"></a>**

- 支持的最大节点组数为1024。
- 支持的最大边缘节点数为1024。
- 单个节点组可包含边缘节点数最大为1024。
- 单个边缘节点可加入的节点组数最大为10。
- 对未纳管节点不能进行[查询未被MEF纳管的节点列表](#查询未被mindedge-framework纳管的节点列表)和[删除未被MEF纳管的节点](#删除未被mindedge-framework纳管的节点)以外的更新操作。

**节点状态<a name="section153111018349"></a>**

- ready：就绪。已对接K8s，节点正常。
- notReady：未就绪。已对接K8s，节点状态异常。
- unknown：未知。节点和K8s连接中断，状态未知。
- offline：掉线。K8s集群内部无该节点，但节点存在于MEF Center数据库中。
- abnormal：节点和MEF Center连接异常。

**纳管边缘节点流程介绍<a name="section433103413267"></a>**

纳管操作是将集群中的节点纳入MEF系统中进行管理。MEF Center对边缘节点的管理基于已成功纳管的节点，MEF Center对未纳管节点不进行任务操作。调用接口使用时，用户可先创建节点组，在纳管节点的同时将节点加入组中；也可先单独纳管节点，不加入任何节点组，后续进行容器应用相关操作前将该节点加入一个或多个组。流程示例如下。

1. 创建节点组或使用已有节点组

    创建节点组是为了将节点添加到节点组中进行纳管，对节点进行批量管理和操作。创建节点组接口参见[创建节点组](#创建节点组)。

    ```text
    https://{ip}:{port}/edgemanager/v1/nodegroup
    ```

2. （可选）查询未纳管节点

    查询未纳管节点是为了找到当前MEF Edge设备对应的节点ID。查询未纳管节点接口参见[查询未被MEF纳管的节点列表](#查询未被mindedge-framework纳管的节点列表)。

    ```text
    https://{ip}:{port}/edgemanager/v1/node/list/unmanaged?pageNum={pageNum}&pageSize={pageSize}&name={name}
    ```

3. 纳管节点
    - 纳管节点是为了将MEF Edge节点设备信息添加到节点数据库中，MEF Center对节点的所有操作都是基于已纳管的节点。纳管节点接口参见[纳管节点](#纳管节点)。

        ```text
        https://{ip}:{port}/edgemanager/v1/node/add
        ```

    - 纳管节点时若未设置节点组ID，在使用节点资源前，通过[向节点组添加节点](#向节点组添加节点)将节点添加到节点组中。

        ```text
        https://{ip}:{port}/edgemanager/v1/nodegroup/node
        ```

4. （可选）修改节点

    以节点ID作为唯一标识，修改节点可修改节点的名称和描述。修改节点接口参见[修改已纳管节点](#修改已纳管节点)。

    ```text
    https://{ip}:{port}/edgemanager/v1/node
    ```

5. （可选）删除已被纳管的节点或将节点移出节点组
    - 删除已被纳管的节点

        根据指定的节点ID数组批量删除节点。删除节点接口参见[删除已被MEF纳管的节点](#删除已被mindedge-framework纳管的节点)。

        ```text
        https://{ip}:{port}/edgemanager/v1/node/batch-delete
        ```

    - 将节点移出节点组
        - 通过将节点从节点组内删除，可批量删除多个节点，卸载并删除节点上的容器应用。接口参见[删除节点组内节点](#删除节点组内节点)。

            ```text
            https://{ip}:{port}/edgemanager/v1/nodegroup/node/batch-delete
            ```

        - 通过删除单个容器应用的Pod，可将节点移出节点组实现对应的容器应用卸载。接口参见[容器应用卸载](#容器应用卸载)。

            ```text
            https://{ip}:{port}/edgemanager/v1/nodegroup/pod/batch-delete
            ```

### 节点组管理<a name="ZH-CN_TOPIC_0000001526721216"></a>

#### 创建节点组<a id="创建节点组"></a>

**命令功能<a name="section6814152820473"></a>**

MEF Center收到节点组创建消息后，对消息中各个字段的有效性进行校验，校验通过之后，将节点组保存到数据库中，并在消息中返回节点组ID。

**命令格式<a name="section1523485417488"></a>**

操作类型：**POST**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup**

请求消息体：

```json
{
    "nodeGroupName": NodeGroupName,
    "description": NodeGroupDescription
}
```

**请求参数<a name="section629411392578"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|nodeGroupName|必选|节点组名称|字符串，取值长度为1~32个字符，支持大小写字母、数字、下划线(_)；只能以大小写字母开头，且不能以下划线结尾。边缘节点组名称是边缘节点组唯一标识，不能重复。|
|description|可选|节点组描述|字符串，长度范围为0~512个字符，包含非空白字符和空格。|

**使用样例<a name="section66472019595"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup
```

请求消息体：

```json
{
    "nodeGroupName": "node_group_name",
    "description": "node_group_description"
}
```

响应样例：

```json
{
   "status": "00000000",
   "msg": "success",
   "data": 1
}
```

响应状态码：200

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|数字|成功创建的节点组ID|

#### 查询节点组列表<a name="ZH-CN_TOPIC_0000001577441477"></a>

**命令功能<a name="section126491058142014"></a>**

查询节点组列表。URL参数用于对数据的分页查询，返回已创建的节点组，以及这些节点组包含的节点总数。

**命令格式<a name="section8841172362113"></a>**

操作类型：**GET**

**URL：https:**//_\{ip\}:\{port\}_/**edgemanager/v1/nodegroup/list?pageSize**=_\{pageSize\}_**&pageNum=**_\{pageNum\}_**&name=**_\{name\}_

**URL参数<a name="section6302105810236"></a>**

**表 1**  URL参数

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|pageSize|必选|页面大小|取值为1~100的整数。|
|pageNum|必选|取值页数，为序数词|取值最小为1，最大值为2^31-1。|
|name|可选|模糊查找的搜索关键词|字符串，长度为0~253位的字符串，不能包含空白字符。|

**使用样例<a name="section441715250263"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/nodegroup/list?pageSize=20&pageNum=1&name=group1
```

响应样例：

```json
{
   "status": "00000000",
   "msg": "success",
   "data": {
       "groups": [
           {
                "createdAt": "2022-12-27 03:29:33",
                "description": "node_group_description",
                "groupName": "node_group_name",
                "id": 1,
                "nodeCount": 1,
                "updatedAt": "2022-12-27 03:29:33"
           }
       ],
       "total": 1
   }
}
```

响应状态码：200

**输出说明<a name="section1855315253619"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|节点组信息|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|groups|数组|节点组详情|
|createdAt|字符串|节点组创建时间|
|description|字符串|节点组描述|
|groupName|字符串|节点组名称|
|id|数字|节点组ID|
|nodeCount|数字|该节点组的节点总数|
|updatedAt|字符串|节点组修改时间|
|total|数字|符合条件的节点组总数|

#### 查询节点组详情<a name="ZH-CN_TOPIC_0000001526881188"></a>

**命令功能<a name="section64293810487"></a>**

查询节点组详情。按照指定的节点组ID，返回该节点组的详细信息，以及所有属于该节点组的节点详情。

**命令格式<a name="section10138183434819"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup?id=**_\{__id__\}_

**URL参数<a name="section49163584912"></a>**

**表 1**  URL参数

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|id|必选|节点组ID|32位无符号数。取值最小为1，最大值为2^32-1。|

**使用样例<a name="section39743015111"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/nodegroup?id=1
```

响应样例：

```json
{
   "status": "00000000",
   "msg": "success",
   "data": {
       "createdAt": "2022-12-27 03:29:33",
       "updatedAt": "2022-12-27 03:29:33",
       "description": "node_group_description",
       "groupName": "node_group_name",
       "id": 1,
       "nodes": [
           {
                "createdAt": "2022-12-27 03:29:33",
                "description": "node_1_description",
                "id": 1,
                "ip": "xx.xx.xx.xx",
                "isManaged": true,
                "nodeName": "node-1",
                "nodeType": "",
                "serialNumber": "xxxxxxxxxxxx",
                "softwareInfo": "[{\"Name\":\"MEFEdge\",\"Version\":\"7.0.RC1\",\"InactiveVersion\":\"\"}]",
                "status": "ready",
                "uniqueName": "edge-66.67",
                "updatedAt": "2022-12-27 03:29:33",
           }
       ]
   }
}
```

响应状态码：200

**输出说明<a name="section051915331053"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|节点组详细信息|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|createdAt|字符串|节点组创建时间|
|updatedAt|字符串|节点组修改时间|
|description|字符串|节点组描述|
|groupName|字符串|节点组名称|
|id|数字|节点组ID|
|nodes|数组|节点组内节点详情|

**表 4**  nodes字段说明

|参数|类型|参数说明|
|--|--|--|
|createdAt|字符串|节点创建时间|
|description|字符串|节点描述|
|nodeName|字符串|节点名|
|id|数字|节点ID|
|ip|字符串|节点IP|
|isManaged|布尔值|节点是否被纳管|
|nodeType|字符串|节点类型|
|status|字符串|节点状态<li>ready：就绪</li><li>notReady：未就绪</li><li>unknown：未知</li><li>offline：掉线</li><li>abnormal：异常</li>|
|uniqueName|字符串|节点主机名|
|updatedAt|字符串|节点修改时间|
|serialNumber|字符串|节点序列号|
|softwareInfo|字符串|节点软件信息|

#### 修改节点组<a name="ZH-CN_TOPIC_0000001577401161"></a>

**命令功能<a name="section154651437121619"></a>**

修改节点组信息。按照指定的节点组ID，修改节点组参数，目前支持修改节点组名和节点组描述两个字段。

**命令格式<a name="section1914654181610"></a>**

操作类型：**PATCH**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup**

请求消息体：

```json
{
    "groupID": NodeGroupId,
    "nodeGroupName": NodeGroupName,
    "description": NodeGroupDescription
}
```

**请求参数<a name="section3775151511234"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|groupID|必选|节点组ID|32位无符号数。取值最小为1，最大值为2^32-1。|
|nodeGroupName|必选|节点组名称|字符串，取值长度为1~32个字符，支持大小写字母、数字、下划线（_）；只能以大小写字母开头，且不能以下划线结尾。边缘节点组名称是边缘节点组唯一标识，不能重复。|
|description|可选|节点组描述|字符串，取值长度范围为0~512个字符，包含非空白字符和空格。|

**使用样例<a name="section115421857182411"></a>**

请求样例：

```bash
PATCH https://10.10.10.10:30035/edgemanager/v1/nodegroup
```

请求消息体：

```json
{
    "groupID": 1,
    "nodeGroupName": "node_group_1",
    "description": "node_group_1_description"
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section921712284514"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|

#### 删除节点组<a name="ZH-CN_TOPIC_0000001527041116"></a>

**命令功能<a name="section151031891312"></a>**

可以批量删除节点组，根据指定的节点组ID数组进行删除；MEF只允许删除没有被部署的应用，当节点组部署有容器应用实例时，将拒绝删除。

**命令格式<a name="section161763119328"></a>**

操作类型：**POST**

**URL：https:**//_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/batch-delete**

请求消息体：

```json
{
    "groupIDs": [NodeGroupId]
}
```

**请求参数<a name="section189991912359"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|groupIDs|必选|删除节点组的ID数组|32位无符号数字数组。数组内元素个数最多为1024个；每个数字取值最小为1，最大值为2^32-1。|

**使用样例<a name="section152662567362"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/batch-delete
```

请求消息体：

```json
{
    "groupIDs": [1,2]
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section0500202382618"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功删除的节点组ID。|
|failedInfos|哈希表，key和value的类型都为字符串|key值为删除失败的节点组ID，value为此ID失败原因。|

#### 节点组数量统计<a name="ZH-CN_TOPIC_0000001577280981"></a>

**命令功能<a name="section12583119104618"></a>**

统计节点组数量。

**命令格式<a name="section17413205114713"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/stats**

**使用样例<a name="section1544122611485"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/nodegroup/stats
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": 1
}
```

响应状态码：200

**输出说明<a name="section3345210496"></a>**

**表 1**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|数字|节点组总数|

### 节点管理<a name="ZH-CN_TOPIC_0000001577601057"></a>

#### 查询未被MEF纳管的节点列表<a id="查询未被mindedge-framework纳管的节点列表"></a>

**命令功能<a name="section1477411417301"></a>**

查询未被MEF纳管的节点列表。URL参数用于对数据的分页查询，返回未纳管节点详情。

> [!NOTE] 说明   
> 安装了MEF Center软件包的节点不会出现在未被MEF纳管的节点中。

**命令格式<a name="section3953928183116"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/list/unmanaged?pageNum=**_\{pageNum\}_**&pageSize=**_\{pageSize\}_**&name=**_\{name\}_

**URL参数<a name="section19387122210348"></a>**

**表 1**  URL参数

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|pageSize|必选|页面大小|取值为1~100的整数。|
|pageNum|必选|取值页数，为序数词|取值最小为1，最大值为2^31-1。|
|name|可选|模糊查找的搜索关键词|字符串，长度为0~253位的字符串，不能包含空白字符。|

**使用样例<a name="section18245194013513"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/list/unmanaged?pageNum=1&pageSize=20&name=node-1
```

响应样例：

```json
{
   "status": "00000000",
   "msg": "success",
   "data": {
       "nodes": [
            {
                "createdAt": "2022-12-27 03:29:33",
                "description": "node_1_description",
                "id": 1,
                "ip": "xx.xx.xx.xx",
                "isManaged": false,
                "nodeName": "node-1",
                "nodeType": "",
                "serialNumber": "xxxxxxxxxxxxxxx",
                "softwareInfo": "",
                "status": "ready",
                "uniqueName": "edge-66.67",
                "updatedAt": "2022-12-27 03:29:33"
             }
       ],
       "total": 1
   }
}
```

响应状态码：200

**输出说明<a name="section051915331053"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|nodes|数组|未被纳管的节点详情|
|total|数字|查询到的节点总数|

**表 4**  nodes字段说明

|参数|类型|参数说明|
|--|--|--|
|createdAt|字符串|节点创建时间|
|description|字符串|节点描述|
|id|数字|节点ID（作为纳管节点时的标识）|
|ip|字符串|节点IP|
|isManaged|布尔值|节点是否被纳管|
|nodeType|字符串|节点类型|
|nodeName|字符串|节点名|
|serialNumber|字符串|节点序列号|
|softwareInfo|字符串|节点软件信息|
|status|字符串|节点状态<li>ready：就绪</li><li>notReady：未就绪</li><li>unknown：未知</li><li>offline：掉线</li><li>abnormal：异常</li>|
|uniqueName|字符串|节点主机名|
|updatedAt|字符串|节点修改时间|

#### 删除未被MEF纳管的节点<a id="删除未被mindedge-framework纳管的节点"></a>

**命令功能<a name="section95491426172615"></a>**

MEF Center能够将K8s集群内的节点加入数据库，使其成为未纳管节点。通过删除未纳管节点接口，可以将未纳管节点从集群中删除。

**命令格式<a name="section13923204216269"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/batch-delete/unmanaged**

请求消息体：

```json
{
    "nodeIDs": [NodeId]
}
```

**请求参数<a name="section1240011511303"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|nodeIDs|必选|节点ID数组|32位无符号数字数组。数组内元素个数最多为1024个；每个数字取值最小为1，最大值为2^32-1。|

**使用样例<a name="section15829115143510"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/node/batch-delete/unmanaged
```

请求消息体：

```json
{
     "nodeIDs": [1,2]
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section1293674119536"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功删除的节点ID。|
|failedInfos|哈希表，key和value的类型都为字符串|key值为失败的节点ID，value为此ID失败原因。|

#### 纳管节点<a id="纳管节点"></a>

**命令功能<a name="section95491426172615"></a>**

MEF Center能够将K8s集群内的节点加入数据库，使其成为未纳管节点。通过查询未纳管节点接口，可以获取这些节点的ID。利用ID参数，可以将未纳管节点加入MEF Center，使其成为已纳管节点。

**命令格式<a name="section13923204216269"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/add**

请求消息体：

```json
{
    "name": NodeName,
    "description": NodeDescription,
    "groupIDs": [GroupId],
    "nodeID": NodeId
}
```

**请求参数<a name="section1240011511303"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|name|必选|节点名称|字符串，取值长度范围为1~64，支持大写字母、小写字母、数字和其他字符（-_）；且不能以下划线或者中划线开头和结尾。|
|description|可选|节点描述|字符串，取值长度范围为0~512个字符；包含非空白字符和空格。|
|groupIDs|可选|节点加入的节点组ID数组|32位无符号数字数组。数组内元素个数最多为10个；每个数字取值最小为1，最大值为2^32-1。|
|nodeID|必选|节点ID|32位无符号数。取值最小为1，最大值为2^32-1。|

**使用样例<a name="section15829115143510"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/node/add
```

请求消息体：

```json
{
     "name": "node-1",
     "description": "node_1_description",
     "groupIDs": [1,2],
     "nodeID": 1
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section1293674119536"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功加入的节点组ID。|
|failedInfos|哈希表，key和value的类型都为字符串|key值为加入失败的节点组ID，value为此ID失败原因。|

#### 向节点组添加节点<a id="向节点组添加节点"></a>

**命令功能<a name="section19593191729"></a>**

向节点组添加节点。在节点添加到节点组的过程中，会检查当前节点剩余资源是否足够运行目标节点组内已部署的全部容器应用，如果节点剩余资源不满足要求，添加节点到节点组会失败。

> [!NOTE] 说明   
> MEF会根据容器应用的需求，对节点的CPU、内存和NPU三类资源进行检查，其他类型的容器资源需求需要用户自行保证。
> MEF的资源限制仅对节点状态为“ready“（就绪）的节点生效。

**命令格式<a name="section14423154212211"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/node**

请求消息体：

```json
{
    "groupID": NodeGroupId,
    "nodeIDs": [NodeId]
}
```

**请求参数<a name="section151467537510"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|groupID|必选|新增的节点组ID|32位无符号数。取值最小为1，最大值为2^32-1。|
|nodeIDs|必选|新增节点组的节点ID数组|32位无符号数字数组。数组内元素个数最多为1024个；每个数字取值最小为1，最大值为2^32-1。|

**使用样例<a name="section15101258574"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/node
```

请求消息体：

```json
{
    "groupID": 1,
    "nodeIDs": [1,2]
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

影响状态码：200

**输出说明<a name="section1293674119536"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功添加的节点ID|
|failedInfos|哈希表，key和value的类型都为字符串|key值为添加失败的节点组ID，value为此ID失败原因|

#### 删除节点组内节点<a id="删除节点组内节点"></a>

**命令功能<a name="section16519865104"></a>**

删除节点组内节点，可批量删除多个节点，将节点从节点组内驱逐，卸载并删除节点上的容器应用。

**命令格式<a name="section15356163871019"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/node/batch-delete**

请求消息体：

```json
{
    "groupID": NodeGroupId,
    "nodeIDs": [NodeId]
}
```

**请求参数<a name="section9187164141312"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|groupID|必选|删除节点的节点组的ID|32位无符号数。取值最小为1，最大值为2^32-1。|
|nodeIDs|必选|将要删除节点的ID数组|32位无符号数字数组。数组内元素个数最多为1024个；每个数字取值最小为1，最大值为2^32-1。|

**使用样例<a name="section15194112517155"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/node/batch-delete
```

请求消息体：

```json
{
    "groupID": 1,
    "nodeIDs": [1,2]
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section1293674119536"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功删除的节点ID|
|failedInfos|哈希表，key和value的类型都为字符串|key值为删除失败的节点ID，value为此ID失败原因|

#### 查询已被MEF纳管的节点列表<a id="ZH-CN_TOPIC_0000001577441385"></a>

**命令功能<a name="section135251624204320"></a>**

查询被MEF纳管的节点列表。URL参数用于对数据的分页查询，返回已纳管节点详情，以及这些节点加入的节点组。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/list/managed?pageNum=**_\{pageNum\}_**&pageSize=**_\{pageSize\}_**&name=**_\{name\}_

**URL参数<a name="section1774293413516"></a>**

**表 1**  URL参数

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|pageSize|必选|页面大小|取值为1~100的整数。|
|pageNum|必选|取值页数，为序数词|取值最小为1，最大值为2^31-1。|
|name|可选|模糊查找的搜索关键词|字符串，长度为0~253位的字符串，不能包含空白字符。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/list/managed?pageNum=1&pageSize=20&name=node-1
```

响应样例：

```json
{
   "status": "00000000",
   "msg": "success",
   "data": {
       "nodes": [
            {
                "createdAt": "2022-12-27 03:29:33",
                "description": "node_1_description",
                "id": 1,
                "ip": "xx.xx.xx.xx",
                "isManaged": true,
                "nodeGroup": "node_group_1",
                "nodeName": "node-1",
                "nodeType": "",
                "serialNumber": "xxxxxxxxxxxxxxx",
                "softwareInfo": "[{\"Name\":\"MEFEdge\",\"Version\":\"7.0.RC1\",\"InactiveVersion\":\"\"}]",
                "status": "ready",
                "uniqueName": "edge-66.67",
                "updatedAt": "2022-12-27 03:29:33"
             }
       ],
       "total": 1
   }
}
```

响应状态码：200

**输出说明<a name="section051915331053"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|nodes|数组|被纳管的节点详情|
|total|数字|查询到的节点总数|

**表 4**  nodes字段说明

|参数|类型|参数说明|
|--|--|--|
|createdAt|字符串|节点创建时间|
|description|字符串|节点描述|
|id|数字|节点ID|
|ip|字符串|节点IP|
|isManaged|布尔值|节点是否被纳管|
|nodeGroup|字符串|节点加入的节点组名称，用英文逗号进行分割|
|nodeName|字符串|节点名|
|nodeType|字符串|节点类型|
|serialNumber|字符串|节点序列号|
|softwareInfo|字符串|节点软件信息|
|status|字符串|节点状态<li>ready：就绪</li><li>notReady：未就绪</li><li>unknown：未知</li><li>offline：掉线</li><li>abnormal：异常</li>|
|uniqueName|字符串|节点主机名|
|updatedAt|字符串|节点修改时间|

#### 查询所有节点列表<a name="ZH-CN_TOPIC_0000001577600949"></a>

**命令功能<a name="section135251624204320"></a>**

查询所有节点列表。URL参数用于对数据的分页查询，返回数据库中的节点详情，以及这些节点加入的节点组。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/list?pageNum=**_\{pageNum\}_**&pageSize=**_\{pageSize\}_**&name=**_\{name\}_

**URL参数<a name="section1774293413516"></a>**

**表 1**  URL参数

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|pageSize|必选|页面大小|取值为1~100的整数。|
|pageNum|必选|取值页数，为序数词|取值最小为1，最大值为2^31-1。|
|name|可选|模糊查找的搜索关键词|字符串，长度为0~253位的字符串，不能包含空白字符。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/list?pageNum=1&pageSize=20&name=node-1
```

响应样例：

```json
{
   "status": "00000000",
   "msg": "success",
   "data": {
       "nodes": [
            {
                "createdAt": "2022-12-27 03:29:33",
                "description": "node_1_description",
                "id": 1,
                "ip": "xx.xx.xx.xx",
                "isManaged": true,
                "nodeGroup": "node_group_1",
                "nodeName": "node-1",
                "nodeType": "",
                "serialNumber": "xxxxxxxxxxxxx",
                "softwareInfo": "[{\"Name\":\"MEFEdge\",\"Version\":\"7.0.RC1\",\"InactiveVersion\":\"\"}]",
                "status": "ready",
                "uniqueName": "edge-66.67",
                "updatedAt": "2022-12-27 03:29:33"
             }
       ],
       "total": 1
   }
}
```

响应状态码：200

**输出说明<a name="section051915331053"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|nodes|数组|查询结果节点详情|
|total|数字|查询到的节点总数|

**表 4**  nodes字段说明

|参数|类型|参数说明|
|--|--|--|
|createdAt|字符串|节点创建时间|
|description|字符串|节点描述|
|id|数字|节点ID|
|ip|字符串|节点IP|
|isManaged|布尔值|节点是否被纳管|
|nodeGroup|字符串|节点加入的节点组名称，用英文逗号进行分割|
|nodeName|字符串|节点名|
|nodeType|字符串|节点类型|
|serialNumber|字符串|节点序列号|
|softwareInfo|字符串|节点软件信息|
|status|字符串|节点状态<li>ready：就绪</li><li>notReady：未就绪</li><li>unknown：未知</li><li>offline：掉线</li><li>abnormal：异常</li>|
|uniqueName|字符串|节点主机名|
|updatedAt|字符串|节点修改时间|

#### 查询节点详情<a id="查询节点详情"></a>

**命令功能<a name="section36445448496"></a>**

查询节点详细信息。该接口除数据库中节点信息外，还包含节点的CPU、NPU和内存资源数据。

**命令格式<a name="section832411165018"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/node?id=**_\{__id__\}_或**https://**_\{ip\}:\{port\}_**/edgemanager/v1/node?sn=**_\{__sn__\}_

**URL参数<a name="section9868223165119"></a>**

**表 1**  URL参数

<a name="table737675217261"></a>
<table><thead align="left"><tr id="row153763524260"><th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.5.1.1"><p id="p837615213264"><a name="p837615213264"></a><a name="p837615213264"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20.02%" id="mcps1.2.5.1.2"><p id="p15376205202612"><a name="p15376205202612"></a><a name="p15376205202612"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p193761052142613"><a name="p193761052142613"></a><a name="p193761052142613"></a>参数说明</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p3376195215262"><a name="p3376195215262"></a><a name="p3376195215262"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row18376195213268"><td class="cellrowborder" valign="top" width="19.98%" headers="mcps1.2.5.1.1 "><p id="p1837619526269"><a name="p1837619526269"></a><a name="p1837619526269"></a>id</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20.02%" headers="mcps1.2.5.1.2 "><p id="p137614528264"><a name="p137614528264"></a><a name="p137614528264"></a>必选，id或sn二选一</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1737615529262"><a name="p1737615529262"></a><a name="p1737615529262"></a>节点ID</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p151951628144614"><a name="p151951628144614"></a><a name="p151951628144614"></a>32位无符号数。取值最小为1，最大值为2^32-1。</p>
</td>
</tr>
<tr id="row167513014814"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p376163020482"><a name="p376163020482"></a><a name="p376163020482"></a>sn</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p16761930154818"><a name="p16761930154818"></a><a name="p16761930154818"></a>节点设备序列号</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1776530164819"><a name="p1776530164819"></a><a name="p1776530164819"></a>支持小写字母、大写字母，数字，下划线和连字符；不能以下划线或中划线开头和结尾；最大长度为64字节。</p>
</td>
</tr>
</tbody>
</table>

**使用样例<a name="section68445365579"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node?id=1
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
                 "cpu": 24,
                 "createdAt": "2022-12-27 03:29:33",
                 "description": "node_1_description",
                 "id": 1,
                 "ip": "xx.xx.xx.xx",
                 "isManaged": true,
                 "memory": 33408877724,
                 "nodeGroup": "node_group_1",
                 "nodeName": "node-1",
                 "nodeType": "",
                 "npu": 0,
                 "serialNumber": "xxxxxxxxxxxxxx",
                 "softwareInfo": "[{\"Name\":\"MEFEdge\",\"Version\":\"7.0.RC1\",\"InactiveVersion\":\"\"}]",
                 "status": "ready",
                 "uniqueName": "edge-66.67",
                 "updatedAt": "2022-12-27 03:29:33"
    }
}
```

响应状态码：200

**输出说明<a name="section19516562114"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|字段|类型|参数说明|
|--|--|--|
|cpu|数字|CPU数量|
|createdAt|字符串|创建时间|
|description|字符串|节点描述|
|id|数字|节点ID|
|ip|字符串|节点IP|
|isManaged|布尔值|节点是否被纳管|
|memory|数字|节点内存，单位为Byte|
|nodeGroup|字符串|节点加入的节点组名称，用英文逗号进行分割|
|nodeName|字符串|节点名|
|nodeType|字符串|节点类型|
|npu|数字|节点NPU数量|
|serialNumber|字符串|节点序列号|
|softwareInfo|字符串|节点软件信息|
|status|字符串|节点状态<li>ready：就绪</li><li>notReady：未就绪</li><li>unknown：未知</li><li>abnormal：异常</li><li>offline：离线</li>|
|uniqueName|字符串|节点主机名|
|updatedAt|字符串|修改时间|

#### 修改已纳管节点<a id="修改已纳管节点"></a>

**命令功能<a name="section8416161212520"></a>**

修改节点。按照指定的节点ID，修改节点参数，目前支持修改节点名和节点描述两个字段。

**命令格式<a name="section177338310511"></a>**

操作类型：**PATCH**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/node**

请求消息体：

```json
{
    "nodeID": NodeId,
    "nodeName": NodeName,
    "description": NodeDescription
}
```

**请求参数<a name="section136292931320"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|nodeID|必选|节点ID|32位无符号数。取值最小为1，最大值为2^32-1。|
|nodeName|必选|节点名称|字符串，取值长度为1~64字符，支持大写字母、小写字母、数字和其他字符（_-）；且不能以下划线或连字符开头和结尾。|
|description|可选|节点描述|字符串，长度范围为0~512个字符；包含非空白字符和空格。|

**使用样例<a name="section2392154315204"></a>**

请求样例：

```bash
PATCH https://10.10.10.10:30035/edgemanager/v1/node
```

请求消息体：

```json
{
    "nodeID": 1,
    "nodeName": "node-1",
    "description": "node_1_description"
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section10297132653614"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|

#### 删除已被MEF纳管的节点<a id="删除已被mindedge-framework纳管的节点"></a>

**命令功能<a name="section1893641916283"></a>**

可批量删除已被MEF纳管的节点，根据指定的节点ID数组进行删除。节点将会从K8s集群中删除，若希望再次纳管该节点，需要在MEF Edge重新进行网管对接。MEF只允许删除没有被部署的应用的节点，删除指定的节点前需要卸载部署在该节点上的所有应用，请参考[卸载容器应用](#卸载容器应用)。

**命令格式<a name="section26338342289"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/batch-delete**

请求消息体：

```json
{
  "nodeIDs": [nodeId]
}
```

**请求参数<a name="section2073163333113"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|nodeIDs|必选|删除节点ID数组|32位无符号数字数组。数组内元素个数最多为1024个；每个数字取值最小为1，最大值为2^32-1。|

**使用样例<a name="section370651413331"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/node/batch-delete
```

请求消息体：

```json
{
    "nodeIDs": [1,2]
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section1293674119536"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功删除的节点ID|
|failedInfos|哈希表，key和value的类型都为字符串|key值为删除失败的节点ID，value为此ID失败原因|

#### 节点状态统计<a name="ZH-CN_TOPIC_0000001527041068"></a>

**命令功能<a name="section15340202213373"></a>**

查询所有节点状态，并返回该状态的节点数。

**命令格式<a name="section2889313103819"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/stats**

**使用指南<a name="section19319163010396"></a>**

无。

**使用样例<a name="section17966161213413"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/stats
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "ready": 1,
        "notReady": 1,
        "unknown": 1,
        "offline": 1
    }
}
```

响应状态码：200

**输出说明<a name="section1293674119536"></a>**

**表 1**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 2**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|ready|数字|处于就绪状态的节点数|
|notReady|数字|处于未就绪状态的节点数|
|unknown|数字|处于未知状态的节点数|
|offline|数字|处于离线状态的节点数|
|abnormal|数字|处于异常状态的节点数|

#### 容器应用卸载<a id="容器应用卸载"></a>

**命令功能<a name="section141453116219"></a>**

容器应用卸载通过将节点移出对应节点组实现（即删除单个应用的Pod）。

**命令格式<a name="section146212029938"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/pod/batch-delete**

请求消息体：

```json
[
  {
     "nodeID": nodeId,
     "groupID":groupId
  }
 {
     "nodeID": nodeId,
     "groupID": groupId
  }
...
]
```

> [!NOTE] 说明      
> 支持通过列表的形式，一次卸载多个容器应用。列表的取值范围为1~1024，每个列表中的节点ID和节点组ID至少有一个取值必须与其他列表不同。

**请求参数<a name="section435110501651"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|nodeID|必选|节点ID|32位无符号数。取值最小为1，最大值为2^32-1。|
|groupID|必选|节点组ID|32位无符号数。取值最小为1，最大值为2^32-1。|

**使用样例<a name="section17189154898"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/pod/batch-delete
```

请求消息体：

```json
[
  {
   "nodeID": 1,
   "groupID":2
},
{
   "nodeID": 3,
   "groupID":4
}
]
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section1293674119536"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|操作结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功卸载的节点和节点组ID对|
|failedInfos|哈希表，key和value的类型都为字符串|key值为卸载失败的节点和节点组ID对，value为此ID对失败原因|

## 容器应用管理接口<a id="ZH-CN_TOPIC_0000001577401181"></a>

### 容器应用接口介绍<a name="ZH-CN_TOPIC_0000001577441409"></a>

容器应用管理作为MEF的基础特性，承担着对用户应用进行全生命周期管理的任务。用户应用以容器镜像的形式发布，MEF对用户的容器应用镜像进行管理，涉及容器应用的增、删、改、查，容器应用部署到节点组，容器从节点组卸载，以及容器从单个节点上卸载，用户通过调用相应接口来实现相应的功能。容器应用以节点组为单位操作，节点加入节点组时自动部署容器到该节点；节点退出节点组时自动卸载该节点上的容器。

> [!NOTE] 说明  
> MEF可以通过Docker公共镜像仓库、第三方镜像仓库、MEF Edge手动导入镜像三种方式使用容器应用的镜像。当使用镜像仓库时，用户需要确保安装MEF Edge设备和镜像仓库之间的网络连接，以及镜像仓库本身能够正常使用。如果用户需要使用第三方镜像仓库获取镜像，使用流程请参见[配置接口介绍](#配置接口介绍)章节进行操作。

**约束说明<a name="section1123013112347"></a>**

- MEF允许同时存在的最大容器应用数量为1000个。若用户将非MEF Center管理的容器应用部署到设备节点，可能导致容器应用因资源不足无法部署。
- MEF Edge节点最多支持部署20个容器应用，容器应用过多可能造成设备性能下降。
- 当用户并发调用部署容器应用相关接口（[部署容器应用](#部署容器应用)、[纳管节点](#纳管节点)、[向节点组添加节点](#向节点组添加节点)）时，可能因节点资源不足无法正常运行。
- MEF Edge系统预留内存资源为1024MB，CPU资源为1个CPU核。
    - 所有容器应用总的可用内存资源可参考公式：可用总内存资源 = 系统总内存资源 - 系统预留内存资源。
    - 所有容器应用总的可用CPU资源可参考公式：可用总CPU资源 = 系统总CPU资源 - 系统预留CPU资源。

- MEF Center最多允许在**单节点组**上部署20个容器应用和在**单节点**上部署20个容器应用。当节点组或者节点上部署的容器应用超过上限时，对应节点组的应用部署功能和新节点加入节点组功能会被限制。

**管理容器应用流程介绍<a name="section132131728522"></a>**

通过调用接口使用容器应用管理相关功能时，对容器应用的创建及部署可以分开操作。用户可以先创建需要的容器应用，后续再决定部署到哪些节点组上。管理容器应用流程示例如下。

1. 创建容器应用

    用户可以通过创建容器应用接口配置容器应用参数，调用成功后会返回创建成功的容器应用AppID。创建容器应用接口参见[创建容器应用](#创建容器应用)。

    ```text
    https://{ip}:{port}/edgemanager/v1/app
    ```

2. （可选）查询容器应用列表

    查询容器应用列表是为了获取到待部署容器应用的AppID。查询容器应用列表接口参见[查询容器应用列表](#查询容器应用列表)。

    ```text
    https://{ip}:{port}/edgemanager/v1/app/list?pageNum={value1}&pageSize={value2}&name={value3}
    ```

3. 部署容器应用

    部署容器应用接口参见[部署容器应用](#部署容器应用)。

    ```text
    https://{ip}:{port}/edgemanager/v1/app/deployment
    ```

4. （可选）查询已部署的容器应用

    用户可以通过查询已部署的容器应用接口获取指定AppID的容器应用运行情况，查询已部署的容器应用参见[查询已部署的容器应用列表](#查询已部署的容器应用列表)。

    ```text
    https://{ip}:{port}/edgemanager/v1/app/deployment?appID={id}
    ```

5. （可选）更新容器应用

    如果对应容器应用已部署，同时会更新对应已部署的容器应用。目前只支持修改容器镜像名称和容器镜像版本的更新。更新容器应用接口参见[更新容器应用](#更新容器应用)。

    ```text
    https://{ip}:{port}/edgemanager/v1/app
    ```

6. （可选）卸载容器应用

    卸载容器应用接口参见[卸载容器应用](#卸载容器应用)。

    ```text
    https://{ip}:{port}/edgemanager/v1/app/deployment/batch-delete
    ```

7. （可选）删除容器应用

    删除容器应用时只允许删除没有被部署的应用。如果对应容器应用已部署，需先卸载该容器应用。删除容器应用接口参见[删除容器应用](#删除容器应用)。

    ```text
    https://{ip}:{port}/edgemanager/v1/app/batch-delete
    ```

### 创建容器应用<a id="创建容器应用"></a>

**命令功能<a name="section135251624204320"></a>**

创建容器应用，消息参数为容器应用的配置信息，被创建的容器应用会保存在MEF Center的数据库中，只有被创建的容器应用才能进行容器应用部署卸载等操作。MEF Center收到容器应用创建消息后，对消息中各个字段的有效性进行校验，校验通过之后，将容器应用参数保存到数据库中，并在消息中返回唯一容器应用ID。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/app**

请求头：

```http
Content-Type: application/json
```

请求消息体：

```json
{
    "appName": AppName,
    "containers": [
        {
            "name": ContainerName,
            "cpuRequest": CpuRequest,
            "cpuLimit": CpuLimit,
            "memRequest": MemoryRequest,
            "memLimit": MemoryLimit,
            "npu": Npu,
            "image": ImageName,
            "imageVersion": ImageVersion,
            "env": [
                {
                    "name": EnvVarName,
                    "value": EnvVarValue
                }
            ],
            "userID": UserId,
            "groupID": GroupId,
            "command": [
                Command
            ],
            "args" : [
                Argument
            ],
            "containerPort" : [
                {
                    "name" : PortName,
                    "proto" : PortProto,
                    "containerPort" : ContainerPort,
                    "hostIP" : HostIP,
                    "hostPort" : HostPort
                }
            ],
            "hostPathVolumes":[
                {
                    "name": name1,
                    "hostPath": HostPath,
                    "mountPath": mountPath
                }
            ]
        }
    ],
    "description": Description
}
```

**请求参数<a id="请求参数"></a>**

**表 1**  创建容器应用参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|appName|必选|容器应用名称|字符串，支持1~32个字符，小写字母、数字和（-），开头和结尾只能是字母或数字。|
|description|可选|容器应用描述信息|字符串，支持0~512个字符；不支持除空格外的其他空白字符。|
|containers|必选|容器应用配置数组|对象数组，单个容器应用支持的容器配置数量为1~10。|

**表 2**  containers参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|name|必选|容器名称。|字符串，取值长度为1~32个字符；支持小写字母、数字和其他字符（-），开头和结尾只能是字母或者数字；应用内容器名称不能重复。|
|cpuRequest|必选|容器申请的CPU核个数。|数字，取值范围为0.01~1000，精确到小数点后两位。|
|cpuLimit|可选|容器最多使用的CPU核个数。|数字，取值范围为0.01~1000，精确到小数点后两位，且取值大于或等于cpuRequest。|
|memRequest|必选|容器申请的内存大小。|数字，取值范围为4~1024000，只能取整数，单位为MiB。|
|memLimit|可选|容器最多使用的内存大小。|数字，取值范围为4~1024000，只能取整数，单位为MiB，且取值大于或等于memRequest。|
|npu|可选|容器申请使用的NPU核个数。|数字，取值范围为0~32，只能取整数。|
|image|必选|使用的镜像名称，使用第三方镜像仓库时，全称需要包含镜像仓库服务器IP或域名、端口、项目和镜像名。例如fd.fusiondirector.huawei\.com:443/library/ubuntu；如果用户不指定镜像的主机名称和端口，容器应用会使用Docker公共仓库。|字符串，取值长度为1~256个字符；支持小写字母、大写字母、数字和其他字符（:-._/）。|
|imageVersion|必选|镜像版本。|字符串，取值长度1~32个字符，小写字母、大写字母、数字和其他字符（-._）。|
|env|可选|容器内配置的环境变量。|EnvVar对象数组，最大支持256组key~value。|
|userID|可选|容器运行指定的用户ID。<br>不配置此项参数时，会以镜像制作的用户运行，如果镜像制作时的用户不为数字ID或者数字ID为0时，容器应用会在部署后运行失败。当运行推理容器时，需要使用驱动设备，因此不能配置用户ID。|数字，取值范围为1~65535，不能配置为0，即不支持容器以root用户运行。部署推理容器时，指定用户ID为HwHiAiUser的用户ID（通常为1000）。|
|groupID|可选|容器运行指定的组ID。<br>不配置此项参数时，会以镜像制作的用户组运行，如果镜像制作时的用户不为数字组ID或者数字组ID为0时，容器应用会在部署后运行失败。当运行推理容器时，需要使用驱动设备，因此不能配置组ID。|数字，取值范围为1~65535，不能配置为0，即容器不支持以root组用户运行。部署推理容器时，指定组ID为HwHiAiUser的组ID（通常为1000）。|
|command|可选|容器启动时的执行命令。|字符串数组，命令列表最多只支持16个，每个命令长度为1~256个字符，支持小写字母、大小字母、数字、空格和其他字符（-/._），结尾只能是大小写字母或数字。|
|args|可选|容器启动时执行的命令参数。|字符串数组，参数列表最多只支持16个，每个命令长度为1~256个字符，支持小写字母、大小字母、数字、空格和其他字符（-/._=），结尾只能是大小写字母或数字。|
|containerPort|可选|容器配置的主机端口和容器内端口映射。|ContainerPort对象数组，最大支持16组端口。|
|hostPathVolumes|可选|容器主机路径挂载配置。<br>创建推理容器应用时需配置挂载路径，否则可能导致运行失败。|HostPathVolumes对象数组，最大支持256组。|

**表 3**  EnvVar参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|name|必选|环境变量名|字符串，取值为长度2~32个字符，支持大小写字母、数字和其他字符（-._）；只能以大小写字母开头，以大小写字母和数字结尾。|
|value|必选|环境变量值|字符串，取值长度为1~512个字符，支持大小写字母、数字其他字符（-._/:）和空格。|

**表 4**  containerPort参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|name|必选|端口映射名称|字符串，取值长度为1~32个字符，支持小写字母、数字、中划线（-）；只能以小写字母、数字开头和结尾。|
|proto|必选|端口映射指定的网络传输层协议|字符串，取值为TCP、UDP。|
|containerPort|必选|容器内端口|数字，取值范围为1~65535，且只能取整数。|
|hostIP|必选|端口映射绑定的主机IP地址|字符串，合法主机IP地址即可，只支持IPv4，不能配置为全0或者全255。|
|hostPort|必选|端口映射主机端口地址|数字，取值范围为1024~65535，且只能取整数。|

**表 5**  hostPathVolumes参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|name|必选|挂载卷名称|字符串，取值长度为1~32个字符；支持小写字母、数字和其他字符（-），开头和结尾只能是字母或者数字。同一个容器内部挂载卷名称不能重复。|
|hostPath|必选|容器挂载卷使用的主机路径。<br> > [!NOTE] 说明  <br>仅支持配置挂载右侧列出的文件或目录的主机路径。若用户参考[制作推理镜像](./common_operations.md#制作推理镜像)或《[Atlas 200I A2 加速模块 昇腾软件快速安装指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100423566/4a72915b)》制作容器镜像，对应默认的镜像内挂载路径请参见制作容器镜像->启动容器步骤。|仅支持配置挂载以下文件或目录的主机路径。<li>"/etc/sys_version.conf"</li><li>"/etc/hdcBasic.cfg"</li><li>"/usr/lib64/libaicpu_processer.so"</li><li>"/usr/lib64/libaicpu_prof.so"</li><li>"/usr/lib64/libaicpu_sharder.so"</li><li>"/usr/lib64/libadump.so"</li><li>"/usr/lib64/libtsd_eventclient.so"</li><li>"/usr/lib64/libaicpu_scheduler.so"</li><li>libcrypto.so.1.1</li><ul><li>Ubuntu宿主机操作系统下："/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1"</li><li>openEuler宿主机操作系统下："/usr/lib64/libcrypto.so.1.1.1m"</li></ul><li>libyaml-0.so.2<ul><li>Ubuntu宿主机操作系统下："/usr/lib/aarch64-linux-gnu/libyaml-0.so.2.0.6"</li><li>openEuler宿主机操作系统下："/usr/lib64/libyaml-0.so.2.0.9"</li></ul><li>"/usr/lib64/libdcmi.so"</li><li>"/usr/lib64/libmpi_dvpp_adapter.so"</li><li>"/usr/lib64/aicpu_kernels/"</li><li>"/usr/local/sbin/npu-smi"</li><li>"/usr/lib64/libstackcore.so"</li><li>"/usr/local/Ascend/driver/lib64"</li><li>"/var/slogd""/var/dmp_daemon"</li>|
|mountPath|必选|容器内挂载路径|以“/”开始的路径字符串，其后可以接大小写字母，数字和其他字符（_./-），不能包含“..”，路径总长度为2-512个字符，同一个容器内部容器挂载路径名称不能重复。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app
```

请求消息体：

```json
{
    "appName": "mef-apptest1",
    "containers": [
        {
            "name": "container1",
            "cpuRequest": 1,
            "cpuLimit": 1,
            "memRequest": 200,
            "memLimit": 200,
            "image": "ubuntu",
            "imageVersion": "22.04",
            "env": [
                {
                    "name": "lib",
                    "value": "/test"
                }
            ],
            "userID": 1001,
            "groupID": 1001,
            "command": [
                "/bin/bash","-c"
            ],
            "args" : [
                "sleep 30000"
            ],
            "containerPort" : [
                {
                    "name" : "test-port",
                    "proto" : "TCP",
                    "containerPort" : 1234,
                    "hostIP" : "xx.xx.xx.xx",
                    "hostPort" : 30023
                }
            ]
        }
    ],
    "description": "a test case for app-manager"
}
```

响应样例：

```json
{
    "status":"00000000",
    "msg":"success",
    "data":3
}
```

响应状态码：200

**输出说明<a name="section102201014616"></a>**

**表 6**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|数字|创建的容器应用ID|

### 查询容器应用列表<a id="查询容器应用列表"></a>

**命令功能<a name="section135251624204320"></a>**

查询容器应用列表，URL参数用于对数据的分页查询，返回已创建的容器应用配置详情，以及这些容器应用的部署的节点组信息。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/list?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&name=**_\{value3\}_

**URL参数<a name="section1774293413516"></a>**

**表 1**  URL参数

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|pageSize|必选|分页查询页面大小|取值为1~100的整数。|
|pageNum|必选|分页查询取值页数，为序数词|取值最小为1，最大值为2^31-1。|
|name|可选|模糊查找的搜索关键词，结果只会返回包含该字段的容器应用名称的容器应用。|字符串，长度为0~253位的字符串，不能包含空白字符。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/list?pageNum=1&pageSize=10&name=123
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "list apps Infos success",
    "data": {
        "appInfo": [
            {
                "appID": 1,
                "appName": "test0115",
                "containers": [
                    {
                        "args": null,
                        "command": null,
                        "containerPort": null,
                        "cpuLimit": 2.1,
                        "cpuRequest": 2.1,
                        "env": [
                            {
                                "name": "lib",
                                "value": "/test"
                            }
                        ],
                        "groupID": 1001,
                        "hostPathVolumes": null,
                        "image": "ubuntu",
                        "imageVersion": "18.04",
                        "memLimit": 200,
                        "memRequest": 200,
                        "npu": 2,
                        "name": "c1",
                        "userID": 1001
                    }
                ],
                "createdAt": "2023-09-12 17:26:04",
                "description": "app-description",
                "modifiedAt": "2023-09-12 17:26:04",
                "nodeGroupInfos": [
                    "nodeGroupID": 1,
                    "nodeGroupName": "test_group_1"
                ]
            }
        ],
        "deployed": 0,
        "total": 1,
        "unDeployed": 1
    }
}

```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

更多关于容器应用的字段说明，请参考创建容器应用章节的[请求参数](#请求参数)。

**表 2**  操作输出说明

|字段|类型|说明|
|--|--|--|
|status|字符串|状态码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|字段|类型|说明|
|--|--|--|
|appInfo|对象数组|容器应用信息数组|
|total|数字|按照模糊查询字段筛选后的结果总数|
|deployed|数字|已部署容器应用数量|
|unDeployed|数字|未部署的容器应用数量|

**表 4**  appInfo字段说明

|字段|类型|说明|
|--|--|--|
|appID|数字|容器应用ID|
|appName|字符串|容器应用名称|
|description|字符串|容器应用描述信息|
|createdAt|字符串|创建时间|
|modifiedAt|字符串|更新时间|
|nodeGroupInfos|nodeGroupInfo对象数组|节点组信息|
|containers|Container对象数组|容器配置数组|

**表 5**  nodeGroupInfo字段说明

|字段|类型|说明|
|--|--|--|
|nodeGroupID|数字|节点组ID|
|nodeGroupName|字符串|节点组名|

**表 6**  containers字段说明

|参数|类型|说明|
|--|--|--|
|name|字符串|容器名称|
|image|字符串|使用的镜像名称|
|imageVersion|字符串|镜像版本|
|cpuRequest|数字|需要的CPU数|
|cpuLimit|数字|最大使用CPU数|
|memRequest|数字|需要的内存大小|
|memLimit|数字|使用的最大内存大小|
|npu|数字|使用的NPU个数|
|command|字符串数组|容器执行命令|
|args|字符串数组|容器执行命令参数|
|env|EnvVar对象数组|环境变量|
|containerPort|ContainerPort对象数组|端口映射容器中端口|
|userID|数字|容器运行用户ID|
|groupID|数字|容器运行组ID|
|hostPathVolumes|HostPathVolumes对象数组|容器主机路径挂载信息|

### 查询容器应用详情<a id="查询容器应用详情"></a>

**命令功能<a name="section135251624204320"></a>**

查看容器应用详情，按照指定的容器应用ID，返回该容器应用的配置详情和部署的节点组信息。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app?appID=**_\{id\}_

**参数说明<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|appID|必选|容器应用ID|取值最小为1，最大值为2^32-1的整数。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app?appID=2
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "appID": 2,
        "appName": "test0115",
        "containers": [
            {
                "args": null,
                "command": [],
                "containerPort": null,
                "cpuLimit": 2.1,
                "cpuRequest": 2.1,
                "env": [
                    {
                        "name": "lib",
                        "value": "/test"
                    }
                ],
                "groupID": 1001,
                "hostPathVolumes": null,
                "image": "ubuntu",
                "imageVersion": "18.04",
                "memLimit": 200,
                "memRequest": 200,
                "npu": 2,
                "name": "c1",
                "userID": 1001
            }
        ],
        "createdAt": "2023-09-12 17:26:04",
        "description": "app-description",
        "modifiedAt": "2023-09-12 17:26:04",
        "nodeGroupInfos": []
    }
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

更多关于容器应用的字段说明，请参考创建容器应用章节的[请求参数](#请求参数)。

**表 2**  操作输出说明

|字段|类型|说明|
|--|--|--|
|status|字符串|状态码|
|msg|字符串|描述信息|
|data|对象|容器应用信息|

**表 3**  data字段说明

|字段|类型|说明|
|--|--|--|
|appID|数字|容器应用ID|
|appName|字符串|容器应用名称|
|containers|container数组|容器配置数组|
|createdAt|字符串|创建时间|
|description|字符串|容器应用描述信息|
|modifiedAt|字符串|更新时间|
|nodeGroupInfos|nodeGroupInfo对象数组|节点组信息|

**表 4**  containers字段说明

|参数|类型|说明|
|--|--|--|
|name|字符串|容器名称|
|image|字符串|使用的镜像名称|
|imageVersion|字符串|镜像版本|
|cpuRequest|数字|需要的CPU数|
|cpuLimit|数字|最大使用CPU数|
|memRequest|数字|需要的内存大小|
|memLimit|数字|使用的最大内存大小|
|npu|数字|使用的NPU个数|
|command|字符串数组|容器执行命令|
|args|字符串数组|容器执行命令参数|
|env|EnvVar对象数组|环境变量|
|containerPort|ContainerPort对象数组|端口映射容器中端口|
|userID|数字|容器运行用户ID|
|groupID|数字|容器运行组ID|
|hostPathVolumes|HostPathVolumes对象数组|容器主机路径挂载信息|

### 部署容器应用<a id="部署容器应用"></a>

**命令功能<a name="section135251624204320"></a>**

部署容器应用，批量接口，会将指定ID的容器应用部署到指定的ID数组的一个或者多个节点组上。部署应用时会根据边缘侧节点剩余可用资源进行限制，如果节点组内在线（即节点状态为“ready“）节点的资源不满足待部署容器应用的需求，部署容器应用到对应节点组会失败。

> [!NOTE] 说明  
>
>- MEF Center部署的容器应用成功时，容器应用在K8s中的daemonset资源会创建成功，MEF Edge容器应用实际运行需要通过查询应用实例进行确认。
>- MEF会根据容器应用的需求，对节点的CPU、内存和NPU三类资源进行检查，其他类型的容器资源需求需要用户自行保证。
>- MEF的资源限制仅对节点状态为“ready“（就绪）的节点生效。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment**

请求头：

```http
Content-Type: application/json
```

请求消息体：

```json
{
    "appID": AppId,
    "nodeGroupIds": [NodeGroupId]
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|类型|说明|取值要求|
|--|--|--|--|
|appID|必选|应用ID|数字，取值最小为1，最大值为2^32-1的整数，必须是存在的应用ID。|
|nodeGroupIds|必选|节点组ID列表|数组，必须是不能重复的节点组ID，且数组长度为[1,1024]。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app/deployment
```

请求消息体：

```json
{
    "appID": 1,
    "nodeGroupIds": [
        1,2
    ]
}
```

响应样例：

```json
{
    "status":"00000000",
    "msg":"success"
}
```

响应状态码：200

**输出说明<a name="section18838112010719"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功部署的节点组ID|
|failedInfos|哈希表，key和value的类型都为字符串|key值为部署失败的节点组ID，value为此ID失败原因|

### 查询已部署的容器应用列表<a id="查询已部署的容器应用列表"></a>

**命令功能<a name="section135251624204320"></a>**

查询已部署的容器应用信息列表，会根据指定的容器应用ID返回已部署的该容器应用的实例列表，包括这些实例的节点和节点组信息、运行状态以及容器状态。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment**?**appID**=_\{id\}_

**URL参数<a name="section1774293413516"></a>**

**表 1**  URL参数

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|appID|必选|容器应用ID|取值最小为1，最大值为2^32-1的整数。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/deployment?appID=1
```

响应样例：

```json
{
    "status":"00000000",
    "msg":"success",
    "data":{
        "appInstances": [
            "appID":1,
            "appName":"testapp",
            "appStatus":"pending",
            "containerInfo":[
                {
                    "image":"euler_image:1.0",
                    "name":"testcontainer",
                    "status":"unknown",
                    "restartCount":0
                }
            ],
            "createdAt":"2022-12-14 08:47:42",
            "nodeGroupInfo":{
                 "nodeGroupID":1,
                 "nodeGroupName":"group1"
             },
             "nodeId":2,
             "nodeName":"localhost.localdomain",
             "nodeStatus":"ready"
         },
        {
            "appID":2,
            "appName":"testapp2",
            "appStatus":"pending",
            "containerInfo":[
                {
                    "image":"ubuntu:18.04",
                    "name":"c1",
                    "status":"unknown",
                    "restartCount":0
                }
            ],
            "createdAt":"2022-12-14 08:48:49",
            "nodeGroupInfo":{
                 "nodeGroupID":1,
                 "nodeGroupName":"group1"
            },
            "nodeId":2,
            "nodeName":"localhost.localdomain",
            "nodeStatus":"ready"
        }
        ],
        "total": 2
    }
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

更多关于容器应用的字段说明，请参考创建容器应用章节的[请求参数](#请求参数)。

**表 2**  操作输出说明

|字段|类型|说明|
|--|--|--|
|status|字符串|状态码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|字段|类型|说明|
|--|--|--|
|appInstances|对象数组|已部署的容器应用列表|
|total|数字|查询结果总数|

**表 4**  appInstances字段说明

|字段|类型|说明|
|--|--|--|
|appID|数字|容器应用ID|
|appName|字符串|容器应用名称|
|appStatus|字符串|Pod应用运行状态<li>pending：等待处理</li><li>running：运行中</li><li>succeeded：成功</li><li>failed：失败</li><li>unknown：未知</li>|
|nodeGroupInfo|nodeGroupInfo对象|部署节点组信息|
|nodeID|数字|部署节点ID|
|nodeName|字符串|部署节点名称|
|nodeStatus|字符串|节点状态<li>ready：就绪</li><li>notReady：未就绪</li><li>offline：掉线</li><li>unknown：未知</li>|
|createdAt|字符串|创建时间|
|containerInfo|对象数组|容器应用Pod信息|

**表 5**  containerInfo字段说明

|字段|类型|说明|
|--|--|--|
|name|字符串|容器应用Pod下Container名称|
|image|字符串|Container的镜像名称|
|status|字符串|Pod下Container运行状态。<li>waiting：等待中</li><li>running：运行中</li><li>terminated：已终止</li><li>unknown：未知</li>|
|restartCount|数字|已部署容器应用对应container的重启次数|

**表 6**  nodeGroupInfo字段说明

|字段|类型|说明|
|--|--|--|
|nodeGroupID|数字|节点组ID|
|nodeGroupName|字符串|节点组名|

### 查询节点已部署的容器应用列表<a id="ZH-CN_TOPIC_0000001577441457"></a>

**命令功能<a name="section135251624204320"></a>**

按照节点ID查询指定节点上已部署容器应用列表，会根据指定的节点ID返回在该节点上运行的全部容器应用的实例列表，包括这些实例的节点和节点组信息、运行状态以及容器状态。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/node?**nodeID**=**_\{id\}_

**URL参数<a name="section1774293413516"></a>**

**表 1**  URL参数

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|nodeID|必选|节点ID|32位无符号数。取值最小为1，最大值为2^32-1。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/node?nodeID=2
```

响应样例：

```json
{
    "status":"00000000",
    "msg":"success",
    "data":{
        "appInstances": [
            "appID":1,
            "appName":"testapp",
            "appStatus":"pending",
            "containerInfo":[
                {
                    "image":"euler_image:1.0",
                    "name":"testcontainer",
                    "status":"unknown",
                    "restartCount":0
                }
            ],
            "createdAt":"2022-12-14 08:47:42",
            "nodeGroupInfo":{
                 "nodeGroupID":1,
                 "nodeGroupName":"group1"
             },
             "nodeId":2,
             "nodeName":"localhost.localdomain",
             "nodeStatus":"ready"
        },
        {
            "appID":2,
            "appName":"testapp2",
            "appStatus":"pending",
            "containerInfo":[
                {
                    "image":"ubuntu:18.04",
                    "name":"c1",
                    "status":"unknown"
                    "restartCount":0
                }
            ],
            "createdAt":"2022-12-14 08:48:49",
            "nodeGroupInfo":{
                 "nodeGroupID":1,
                 "nodeGroupName":"group1"
                 },
            "nodeId":2,
            "nodeName":"localhost.localdomain",
            "nodeStatus":"ready"
            }
        ],
        "total": 2
    }
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

更多关于容器应用的字段说明，请参考创建容器应用章节的[请求参数](#请求参数)。

**表 2**  操作输出说明

|字段|类型|说明|
|--|--|--|
|status|字符串|状态码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|字段|类型|说明|
|--|--|--|
|appInstances|对象数组|已部署的容器应用列表|
|total|数字|查询结果总数|

**表 4**  appInstances字段说明

|字段|类型|说明|
|--|--|--|
|appID|数字|容器应用ID|
|appName|字符串|容器应用名称|
|appStatus|字符串|Pod应用运行状态<li>pending：等待处理</li><li>running：运行中</li><li>succeeded：成功</li><li>failed：失败unknown：未知</li>|
|nodeGroupInfo|对象|部署节点组信息|
|nodeID|数字|部署节点ID|
|nodeName|字符串|部署节点名称|
|nodeStatus|字符串|节点状态ready：就绪notReady：未就绪offline：掉线unknown：未知|
|createdAt|字符串|创建时间|
|containerInfo|对象数组|容器应用Pod信息|

**表 5**  containerInfo字段说明

|字段|类型|说明|
|--|--|--|
|name|字符串|容器应用Pod下Container名称|
|image|字符串|Container的镜像名称|
|status|字符串|Pod下Container运行状态。<li>waiting：等待中</li><li>running：运行中</li><li>terminated：已终止</li><li>unknown：未知</li>|
|restartCount|数字|已部署容器应用对应container的重启次数|

**表 6**  nodeGroupInfo字段说明

|字段|类型|说明|
|--|--|--|
|nodeGroupID|数字|节点组ID|
|nodeGroupName|字符串|节点组名|

### 批量查询已部署的容器应用列表<a name="ZH-CN_TOPIC_0000001577401093"></a>

**命令功能<a name="section135251624204320"></a>**

查询已部署的容器应用信息列表，会根据指定的分页查询请求参数返回筛选后的容器应用的实例列表，包括这些实例的节点和节点组信息、运行状态以及容器状态。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment/list?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&name=**_\{value3\}_

**URL参数<a name="section1774293413516"></a>**

**表 1**  URL参数

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|pageSize|必选|分页查询页面大小|取值为1~100的整数。|
|pageNum|必选|分页查询取值页数，为序数词|取值最小为1，最大值为2^31-1。|
|name|可选|模糊查找的搜索关键词，结果只会返回包含该字段的容器应用名称的容器应用。|字符串，长度为0~253位的字符串，不能包含空白字符。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/deployment/list?pageNum=1&pageSize=10&name=
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "appInstances": [
            {
                "appID":1,
                "appName": "mef-apptest1",
                "appStatus": "running",
                "containerInfo": [
                    {
                        "image": "ubuntu:22.04",
                        "name": "c1",
                        "restartCount":0,
                        "status": "running"
                    }
                ],
                "createdAt": "2023-02-01 05:15:51",
                "nodeGroupInfo": {
                    "nodeGroupID": 1,
                    "nodeGroupName": "group1"
                },
                "nodeID": 1,
                "nodeName": "node221",
                "nodeStatus": "ready"
            }
        ],
        "total": 1
    }
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

更多关于容器应用的字段说明，请参考创建容器应用章节的[请求参数](#请求参数)。

**表 2**  操作输出说明

|字段|类型|说明|
|--|--|--|
|status|字符串|状态码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|字段|类型|说明|
|--|--|--|
|appInstances|对象数组|已部署的容器应用列表|
|total|数字|查询结果总数|

**表 4**  appInstances字段说明

|字段|类型|说明|
|--|--|--|
|appID|数字|容器应用ID|
|appName|字符串|容器应用名称|
|appStatus|字符串|Pod应用运行状态，包括以下类型：<li>pending：等待处理<li>running：运行中<li>succeeded：成功<li>failed：失败<li>unknown：未知|
|nodeGroupInfo|对象|部署节点组信息|
|nodeID|数字|部署节点ID|
|nodeName|字符串|部署节点名称|
|nodeStatus|字符串|节点状态，包括以下类型：<li>ready：就绪<li>notReady：未就绪<li>offline：掉线<li>unknown：未知|
|createdAt|字符串|创建时间|
|containerInfo|对象数组|容器应用Pod信息|

**表 5**  containerInfo字段说明

|字段|类型|说明|
|--|--|--|
|name|字符串|容器应用Pod下Container名称|
|image|字符串|Container的镜像名称|
|status|字符串|Pod下Container运行状态。<li>waiting：等待中<li>running：运行中<li>terminated：已终止<li>unknown：未知|
|restartCount|数字|已部署容器应用对应container的重启次数|

**表 6**  nodeGroupInfo字段说明

|字段|类型|说明|
|--|--|--|
|nodeGroupID|数字|节点组ID|
|nodeGroupName|字符串|节点组名|

### 更新容器应用<a id="更新容器应用"></a>

**命令功能<a name="section135251624204320"></a>**

更新容器应用，按照指定容器应用ID，更新MEF Center中保存的容器应用信息，如果对应容器应用已部署，同时会更新对应已部署的容器应用。目前只支持修改对容器镜像名称和容器镜像版本的更新，对其他字段的变更不会被使用。

> [!NOTE] 说明    
> MEF Center更新已部署的容器应用成功时是应用在K8s中的daemonset资源更新成功，MEF Edge运行容器实际更新需要通过查询应用实例进行确认。

**命令格式<a name="section6901955114320"></a>**

操作类型：**PATCH**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/app**

请求头：

```http
Content-Type: application/json
```

请求消息体：

```json
{
    "appID": AppId,
    "appName": AppName,
    "containers": [
       {
           "name": ContainerName,
           "cpuRequest": CpuRequest,
           "cpuLimit": CpuLimit,
           "memRequest": MemoryRequest,
           "memLimit": MemoryLimit,
           "image": ImageName,
           "imageVersion": ImageVersion,
           "env": [
               {
                   "name": EnvVarName,
                   "value": EnvVarValue
               }
           ],
           "userID": UserId,
           "groupID": GroupId,
           "command": [
               Command
           ],
           "args" : [
               Argument
           ],
           "containerPort" : [
               {
                   "name" : PortName,
                   "proto" : PortProto,
                   "containerPort" : ContainerPort,
                   "hostIP" : HostIP,
                   "hostPort" : HostPort
               }
           ]
        }
    ],
    "description": Description
}
```

**请求参数<a name="section1774293413516"></a>**

更多关于容器应用的字段说明，请参考创建容器应用章节的[请求参数](#请求参数)。

**表 1**  参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|appID|必选|容器应用ID|数字，取值最小为1，最大值为2^32-1的整数，必须是存在的应用ID。|
|appName|必选|容器应用名称|字符串，取值长度1~32个字符，小写字母、数字和“-”，开头结尾只能是字母数字。|
|description|可选|容器应用描述信息|字符串，取值长度0~512个字符。并且不支持除空格外的其他空白字符。|
|containers|必选|容器配置数组|对象数组，数组长度为1~10。|

**表 2**  containers参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|name|必选|容器名称。|字符串，取值长度为1~32个字符；支持小写字母、数字和其他字符（-），开头和结尾只能是字母或者数字；应用内容器名称不能重复。|
|cpuRequest|必选|容器申请的CPU核个数。|数字，取值范围为0.01~1000，精确到小数点后两位。|
|cpuLimit|可选|容器最多使用的CPU核个数。|数字，取值范围为0.01~1000，精确到小数点后两位，且取值大于或等于cpuRequest。|
|memRequest|必选|容器申请的内存大小。|数字，取值范围为4~1024000，只能取整数，单位为MiB。|
|memLimit|可选|容器最多使用的内存大小。|数字，取值范围为4~1024000，只能取整数，单位为MiB，且取值大于或等于memRequest。|
|npu|可选|容器申请使用的NPU核个数。|数字，取值范围为0~32，只能取整数。|
|image|必选|使用的镜像名称，使用第三方镜像仓库时，全称需要包含镜像仓库服务器IP或域名、端口、项目和镜像名。例如fd.fusiondirector.huawei\.com:443/library/ubuntu；如果用户不指定镜像的主机名称和端口，容器应用会使用Docker公共仓库。|字符串，取值长度为1~256个字符；支持小写字母、大写字母、数字和其他字符（:-._/）。|
|imageVersion|必选|镜像版本。|字符串，取值长度1~32个字符，小写字母、大写字母、数字和其他字符（-._）。|
|env|可选|容器内配置的环境变量。|EnvVar对象数组，最大支持256组key~value。|
|userID|可选|容器运行指定的用户ID。<br>不配置此项参数时，会以镜像制作的用户运行，如果镜像制作时的用户不为数字ID或者数字ID为0时，容器应用会在部署后运行失败。当运行推理容器时，需要使用驱动设备，因此不能配置用户ID。|数字，取值范围为1~65535，不能配置为0，即不支持容器以root用户运行。<br>部署推理容器时，指定用户ID为HwHiAiUser的用户ID（通常为1000）。|
|groupID|可选|容器运行指定的组ID。<br>不配置此项参数时，会以镜像制作的用户组运行，如果镜像制作时的用户不为数字组ID或者数字组ID为0时，容器应用会在部署后运行失败。当运行推理容器时，需要使用驱动设备，因此不能配置组ID。|数字，取值范围为1~65535，不能配置为0，即容器不支持以root组用户运行。<br>部署推理容器时，指定组ID为HwHiAiUser的组ID（通常为1000）。|
|command|可选|容器启动时的执行命令。|字符串数组，命令列表最多只支持16个，每个命令长度为1~256个字符，支持小写字母、大小字母、数字、空格和其他字符（-/._），结尾只能是大小写字母或数字。|
|args|可选|容器启动时执行的命令参数。|字符串数组，参数列表最多只支持16个，每个命令长度为1~256个字符，支持小写字母、大小字母、数字、空格和其他字符（-/._=），结尾只能是大小写字母或数字。|
|containerPort|可选|容器配置的主机端口和容器内端口映射。|ContainerPort对象数组，最大支持16组端口。|
|hostPathVolumes|可选|容器主机路径挂载配置。<br>创建推理容器应用时需配置挂载路径，否则可能导致运行失败。|HostPathVolumes对象数组，最大支持256组。|

**表 3**  containerPort参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|name|必选|端口映射名称|字符串，取值长度为1~32个字符，支持小写字母、数字、中划线（-）；只能以小写字母、数字开头和结尾。|
|proto|必选|端口映射指定的网络传输层协议|字符串，取值为TCP、UDP。|
|containerPort|必选|容器内端口|数字，取值范围为1~65535，且只能取整数。|
|hostIP|必选|端口映射绑定的主机IP地址|字符串，合法主机IP地址即可，只支持IPv4，不能配置为全0或者全255。|
|hostPort|必选|端口映射主机端口地址|数字，取值范围为1024~65535，且只能取整数。|

**表 4**  EnvVar参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|name|必选|环境变量名|字符串，取值为长度2~32个字符，支持大小写字母、数字和其他字符（-._）；只能以大小写字母开头，以大小写字母和数字结尾。|
|value|必选|环境变量值|字符串，取值长度为1~512个字符，支持大小写字母、数字其他字符（-._/:）和空格。|

**表 5**  hostPathVolumes参数说明

<table><thead align="left"><tr id="zh-cn_topic_0000001527041156_row12899173654011"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000001527041156_p17899836144010"><a name="zh-cn_topic_0000001527041156_p17899836144010"></a><a name="zh-cn_topic_0000001527041156_p17899836144010"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001527041156_p3900836124019"><a name="zh-cn_topic_0000001527041156_p3900836124019"></a><a name="zh-cn_topic_0000001527041156_p3900836124019"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000001527041156_p490033604011"><a name="zh-cn_topic_0000001527041156_p490033604011"></a><a name="zh-cn_topic_0000001527041156_p490033604011"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000001527041156_p5900143617405"><a name="zh-cn_topic_0000001527041156_p5900143617405"></a><a name="zh-cn_topic_0000001527041156_p5900143617405"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001527041156_row199001136144017"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001527041156_p1490053616402"><a name="zh-cn_topic_0000001527041156_p1490053616402"></a><a name="zh-cn_topic_0000001527041156_p1490053616402"></a>name</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001527041156_p189000364407"><a name="zh-cn_topic_0000001527041156_p189000364407"></a><a name="zh-cn_topic_0000001527041156_p189000364407"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001527041156_p18900436164014"><a name="zh-cn_topic_0000001527041156_p18900436164014"></a><a name="zh-cn_topic_0000001527041156_p18900436164014"></a>挂载卷名称</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001527041156_p49001362404"><a name="zh-cn_topic_0000001527041156_p49001362404"></a><a name="zh-cn_topic_0000001527041156_p49001362404"></a>字符串，取值长度为1~32个字符；支持小写字母、数字和其他字符（-），开头和结尾只能是字母或者数字。</p>
<p id="zh-cn_topic_0000001527041156_p14964185819578"><a name="zh-cn_topic_0000001527041156_p14964185819578"></a><a name="zh-cn_topic_0000001527041156_p14964185819578"></a>同一个容器内部挂载卷名称不能重复。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001527041156_row990083694012"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001527041156_p4900203644015"><a name="zh-cn_topic_0000001527041156_p4900203644015"></a><a name="zh-cn_topic_0000001527041156_p4900203644015"></a>hostPath</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001527041156_p490014369404"><a name="zh-cn_topic_0000001527041156_p490014369404"></a><a name="zh-cn_topic_0000001527041156_p490014369404"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001527041156_p169001736194016"><a name="zh-cn_topic_0000001527041156_p169001736194016"></a><a name="zh-cn_topic_0000001527041156_p169001736194016"></a>容器挂载卷使用的主机路径。</p>
<div class="note" id="zh-cn_topic_0000001527041156_note197621102286"><a name="zh-cn_topic_0000001527041156_note197621102286"></a><a name="zh-cn_topic_0000001527041156_note197621102286"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000001527041156_p97121126281"><a name="zh-cn_topic_0000001527041156_p97121126281"></a><a name="zh-cn_topic_0000001527041156_p97121126281"></a>仅支持配置挂载右侧列出的文件或目录的主机路径。若用户参考<a href="./common_operations.md#制作推理镜像">制作推理镜像</a>或<span id="zh-cn_topic_0000001527041156_ph1126911412338"><a name="zh-cn_topic_0000001527041156_ph1126911412338"></a><a name="zh-cn_topic_0000001527041156_ph1126911412338"></a>《<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100423566" target="_blank" rel="noopener noreferrer">Atlas 200I A2 加速模块 昇腾软件快速安装指南</a>》</span>制作容器镜像，对应默认的镜像内挂载路径请参见制作容器镜像-&gt;启动容器步骤。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001527041156_p1935845132715"><a name="zh-cn_topic_0000001527041156_p1935845132715"></a><a name="zh-cn_topic_0000001527041156_p1935845132715"></a>仅支持配置挂载以下文件或目录的主机路径。</p>
<a name="zh-cn_topic_0000001527041156_ul5305123819127"></a><a name="zh-cn_topic_0000001527041156_ul5305123819127"></a><ul id="zh-cn_topic_0000001527041156_ul5305123819127"><li>"/etc/sys_version.conf"</li><li>"/etc/hdcBasic.cfg"</li><li>"/usr/lib64/libaicpu_processer.so"</li><li>"/usr/lib64/libaicpu_prof.so"</li><li>"/usr/lib64/libaicpu_sharder.so"</li><li>"/usr/lib64/libadump.so"</li><li>"/usr/lib64/libtsd_eventclient.so"</li><li>"/usr/lib64/libaicpu_scheduler.so"</li><li>libcrypto.so.1.1<a name="zh-cn_topic_0000001527041156_ul015234614165"></a><a name="zh-cn_topic_0000001527041156_ul015234614165"></a><ul id="zh-cn_topic_0000001527041156_ul015234614165"><li><span id="zh-cn_topic_0000001527041156_ph1052114212228"><a name="zh-cn_topic_0000001527041156_ph1052114212228"></a><a name="zh-cn_topic_0000001527041156_ph1052114212228"></a>Ubuntu</span>宿主机操作系统下："/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1"</li><li><span id="zh-cn_topic_0000001527041156_ph9796114014252"><a name="zh-cn_topic_0000001527041156_ph9796114014252"></a><a name="zh-cn_topic_0000001527041156_ph9796114014252"></a>openEuler</span>宿主机操作系统下："/usr/lib64/libcrypto.so.1.1.1m"</li></ul>
</li><li>libyaml-0.so.2<a name="zh-cn_topic_0000001527041156_ul1253630154718"></a><a name="zh-cn_topic_0000001527041156_ul1253630154718"></a><ul id="zh-cn_topic_0000001527041156_ul1253630154718"><li><span id="zh-cn_topic_0000001527041156_ph125362014710"><a name="zh-cn_topic_0000001527041156_ph125362014710"></a><a name="zh-cn_topic_0000001527041156_ph125362014710"></a>Ubuntu</span>宿主机操作系统下："/usr/lib/aarch64-linux-gnu/libyaml-0.so.2.0.6"</li><li><span id="zh-cn_topic_0000001527041156_ph853614004715"><a name="zh-cn_topic_0000001527041156_ph853614004715"></a><a name="zh-cn_topic_0000001527041156_ph853614004715"></a>openEuler</span>宿主机操作系统下："/usr/lib64/libyaml-0.so.2.0.9"</li></ul>
</li><li>"/usr/lib64/libdcmi.so"</li><li>"/usr/lib64/libmpi_dvpp_adapter.so"</li><li>"/usr/lib64/aicpu_kernels/"</li><li>"/usr/local/sbin/npu-smi"</li><li>"/usr/lib64/libstackcore.so"</li><li>"/usr/local/Ascend/driver/lib64"</li><li>"/var/slogd"</li><li>"/var/dmp_daemon"</li></ul>
<p id="zh-cn_topic_0000001527041156_p15774163614317"><a name="zh-cn_topic_0000001527041156_p15774163614317"></a><a name="zh-cn_topic_0000001527041156_p15774163614317"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000001527041156_row11157325124311"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001527041156_p16157152514437"><a name="zh-cn_topic_0000001527041156_p16157152514437"></a><a name="zh-cn_topic_0000001527041156_p16157152514437"></a>mountPath</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001527041156_p915732564310"><a name="zh-cn_topic_0000001527041156_p915732564310"></a><a name="zh-cn_topic_0000001527041156_p915732564310"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001527041156_p121576255431"><a name="zh-cn_topic_0000001527041156_p121576255431"></a><a name="zh-cn_topic_0000001527041156_p121576255431"></a>容器内挂载路径</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001527041156_p51581525114312"><a name="zh-cn_topic_0000001527041156_p51581525114312"></a><a name="zh-cn_topic_0000001527041156_p51581525114312"></a>以“/”开始的路径字符串，其后可以接大小写字母，数字和其他字符（_./-），不能包含“..”，路径总长度为2-512个字符，同一个容器内部容器挂载路径名称不能重复。</p>
</td>
</tr>
</tbody>
</table>

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
PATCH https://10.10.10.10:30035/edgemanager/v1/app
```

请求消息体：

```json
{
    "appID": 3,
    "appName": "mef-apptest1",
    "containers": [
        {
            "name": "container1",
            "cpuRequest": 1,
            "cpuLimit": 1,
            "memRequest": 200,
            "memLimit": 200,
            "image": "ubuntu",
            "imageVersion": "18.04",
            "env": [
                {
                    "name": "lib",
                    "value": "/test"
                }
            ],
            "userId": 1001,
            "groupId": 1001,
            "command": [
                "/bin/bash","-c"
            ],
            "args" : [
                "sleep 30000"
            ],
            "containerPort" : [
                {
                    "name" : "test-port",
                    "proto" : "TCP",
                    "containerPort" : 1234,
                    "hostIP" : "xx.xx.xx.xx",
                    "hostPort" : 30023
                }
            ],
           "hostPathVolumes":[
           ]
        }
    ],
    "description": "a test case for app-manager"
}
```

响应样例：

```json
{
    "status":"00000000",
    "msg":"success"
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 6**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|

### 卸载容器应用<a id="卸载容器应用"></a>

**命令功能<a name="section135251624204320"></a>**

卸载容器应用，停止对应容器应用实例的运行状态。该接口为批量接口，会根据指定的容器ID将容器应用从指定的节点组ID列表中的一个或多个节点组上卸载。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment/batch-delete**

请求头：

```http
Content-Type: application/json
```

请求消息体：

```json
{
    "appID": AppId,
    "nodeGroupIds": [NodeGroupId]
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|类型|说明|取值要求|
|--|--|--|--|
|appID|必选|容器应用ID|数字，取值最小为1，最大值为2^32-1的整数，必须是存在的应用ID。|
|nodeGroupIds|必选|节点组ID列表|数组，必须是不能重复的节点组ID，且数组长度为[1,1024]。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app/deployment/batch-delete
```

请求消息体：

```json
{
    "appID": 1,
    "nodeGroupIds": [
      1, 2
    ]
}
```

响应样例：

```json
{
    "status":"00000000",
    "msg":"success"
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|成功卸载的节点组ID|
|failedInfos|哈希表，key和value的类型都为字符串|key值为卸载失败的节点组ID，value为此ID失败原因|

### 删除容器应用<a id="删除容器应用"></a>

**命令功能<a name="section135251624204320"></a>**

删除容器应用，批量接口，根据指定的容器应用ID数组进行删除，MEF只允许删除没有被部署的应用。在删除出现失败的情况下，返回字段会包括失败的ID和成功的ID数组。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/batch-delete**

请求头：

```http
Content-Type: application/json
```

请求消息体：

```json
{
    "appIDs": [AppId]
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|类型|说明|取值要求|
|--|--|--|--|
|appIDs|必选|容器应用ID数组|数组内元素个数最多为1024个。每个数字取值最小为1，最大值为2^32-1，必须是存在的应用ID。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app/batch-delete
```

请求消息体：

```json
{
    "appIDs": [
        1,2
    ]
}
```

响应样例：

```json
{
    "status":"00000000",
    "msg":"success"
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|批量操作结果。如果批量操作全部成功，不返回该字段。|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|数组|删除成功的容器应用ID。|
|failedInfos|哈希表，key和value的类型都为字符串|key值为删除失败的容器应用ID，value为此ID失败原因。|

## 日志收集接口<a id="ZH-CN_TOPIC_0000001640208710"></a>

### 日志收集接口介绍<a name="ZH-CN_TOPIC_0000001640049426"></a>

MEF支持收集并导出MEF Edge的日志，实现MEF Edge的日志排查，设备状态监测。

**约束说明<a name="section107641411309"></a>**

- 仅支持同时进行一个日志导出任务
- 每个日志导出任务最多支持100个节点
- 日志下载请求必须在2小时内完成
- 日志导出成功后一天内可以下载
- MEF Center重启后，所有任务会被删除，之前导出的日志不能再继续下载
- 成功开始新的日志收集任务后，之前任务收集的日志将会被删除
- 已经完成的日志收集任务最多存储2000个，多余的任务会被删除

**日志收集流程介绍<a name="section124811062447"></a>**

MEF Edge软件调用接口进行日志收集的流程示例如下。

1. 创建日志收集任务

    通过RESTful接口下发创建日志收集任务，创建任务成功可返回日志收集任务ID，接口请参见[创建日志收集任务](#创建日志收集任务)。

    ```text
    https://{ip}:{port}/edgemanager/v1/logmgmt/dump/task
    ```

2. （可选）查询日志收集任务进度

    通过日志收集任务ID查询日志收集任务进度，用户任务状态为succeed和partiallyFailed时，MEF Center成功收集到MEF Edge日志，接口请参见[查询日志收集任务进度](#查询日志收集任务进度)。

    ```text
    https://{ip}:{port}/edgemanager/v1/logmgmt/dump/task/?taskId={taskId}
    ```

3. （可选）下载日志收集文件

    下载日志收集文件，接口请参见[下载日志收集文件](#下载日志收集文件)。

    ```text
    https://{ip}:{port}/edgemanager/v1/logmgmt/dump/download/edgeNodes.tar.gz
    ```

### 创建日志收集任务<a id="创建日志收集任务"></a>

**命令功能<a name="section135251624204320"></a>**

创建日志收集任务。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/logmgmt/dump/task**

请求消息体：

```json
{
    "module": module,
    "edgeNodes": edgeNodes
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|module|必选|收集的日志类型|字符串，取值为edgeNode。|
|edgeNodes|必选|待收集的节点ID数组|数组，最多100个，最少1个，不可重复；数组内每个数字取值最小为1，最大值为2^32-1。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/logmgmt/dump/task
```

请求消息体：

```json
{
    "module": "edgeNode",
    "edgeNodes": [2]
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "taskId": "dumpMultiNodesLog.413f66d069888b135e976c57"
    }
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|创建日志收集任务结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|taskId|字符串|成功创建的任务ID|

### 查询日志收集任务进度<a id="查询日志收集任务进度"></a>

**命令功能<a name="section135251624204320"></a>**

查询日志收集任务进度。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/edgemanager/v1/logmgmt/dump/task?taskId=**_\{taskId\}_

请求消息体：无

**URL参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|taskId|必选|待查询的任务ID|字符串，以dumpMultiNodesLog开头，且dumpMultiNodesLog后面支持的长度为1~128，支持大写字母，小写字母、数字和其他字符（-_.），如dumpMultiNodesLog.eca14a007f930e5689652a30。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/logmgmt/dump/task?taskId=dumpMultiNodesLog.eca14a007f930e5689652a30
```

请求消息体：无

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "createdAt": "2023-08-23T17:06:56.28075167Z",
        "data": {
            "fileName": "edgeNodes.tar.gz"
        },
        "finishedAt": "2023-08-23T17:07:05.67786258Z",
        "progress": 100,
        "reason": "task succeeded",
        "startedAt": "2023-08-23T17:06:56.28095697Z",
        "status": "succeed",
        "taskId": "dumpMultiNodesLog.eca14a007f930e5689652a30"
    }
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|成功操作的返回值|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|createdAt|字符串|创建时间|
|data|对象|可选的结果|
|fileName|字符串|收集日志的文件名|
|finishedAt|字符串|结束时间|
|progress|数字|日志收集任务进度，最大取值100|
|reason|字符串|任务处于该状态的原因|
|startedAt|字符串|开始时间|
|status|字符串|任务状态，包括如下几种状态：<li>waiting：等待执行</li><li>processing：正在执行</li><li>aborting：正在终止</li><li>succeed：任务成功</li><li>failed：任务完全失败</li><li>partiallyFailed：任务部分失败</li>|
|taskId|字符串|待查询的任务ID|

### 下载日志收集文件<a id="下载日志收集文件"></a>

**命令功能<a name="section135251624204320"></a>**

下载日志收集文件。当用户任务状态为succeed和partiallyFailed时，用户可以下载日志文件。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_<b>/edgemanager/v1/logmgmt/dump/download/</b>edgeNodes.tar.gz

请求消息体：无

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/logmgmt/dump/download/edgeNodes.tar.gz
```

请求消息体：无

响应样例：直接下载文件

响应状态码：200

## 告警事件信息接口<a name="ZH-CN_TOPIC_0000001691744913"></a>

### 告警事件接口介绍<a name="ZH-CN_TOPIC_0000001691665661"></a>

MEF支持通过调用MEF Center提供的接口查询MEF Edge和MEF Center的告警或事件信息。

- 告警信息：从某个时间点开始出现的问题信息。
- 事件信息：在过去某个时间点发生的问题信息。MEF Center数据库针对单个MEF Edge设备，最大保存50个事件。

**查询告警事件信息流程介绍<a name="section124811062447"></a>**

MEF Center调用接口查询告警信息的流程示例如下；查询事件信息的步骤相同，只是调用的接口有所不同。

1. 查询告警列表

    通过RESTful接口下发查询告警或事件列表任务，接口请参见[查询告警列表](#查询告警列表)。

    ```text
    https://{ip}:{port}/alarmmanager/v1/alarms?pageNum={value1}&pageSize={value2}&ifCenter={value3}&sn={value4}&groupId={value5}
    ```

2. （可选）查询告警详情

    通过查询列表返回的MEF Center数据库中的告警标识，查询告警详细信息，接口请参见[查询告警详情](#查询告警详情)。

    ```text
    https://{ip}:{port}/alarmmanager/v1/alarm?id={value1}
    ```

### 查询告警列表<a id="查询告警列表"></a>

**命令功能<a name="section135251624204320"></a>**

查询系统告警数据，URL参数用于指定本次分页查询控制条件、查询的告警种类和查询操作类型，返回MEF Center数据库中已创建的告警信息。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/alarmmanager/v1/alarms?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&ifCenter=**_\{value3\}_**&sn=**_\{value4\}_**&groupId=**_\{value5\}_

请求消息体：无

**URL参数<a name="section1774293413516"></a>**

**表 1**  参数说明

<table><thead align="left"><tr id="row1643491410534"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p1943421410537"><a name="p1943421410537"></a><a name="p1943421410537"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p113476328527"><a name="p113476328527"></a><a name="p113476328527"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p343413143534"><a name="p343413143534"></a><a name="p343413143534"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p18434201414535"><a name="p18434201414535"></a><a name="p18434201414535"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row11236115220615"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p207789294419"><a name="p207789294419"></a><a name="p207789294419"></a>pageNum</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1477862910411"><a name="p1477862910411"></a><a name="p1477862910411"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1977814296415"><a name="p1977814296415"></a><a name="p1977814296415"></a>分页查询取值页数，为序数词</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p187782291042"><a name="p187782291042"></a><a name="p187782291042"></a>取值为1~2^31-1的整数。</p>
</td>
</tr>
<tr id="row19711641046"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1777710299410"><a name="p1777710299410"></a><a name="p1777710299410"></a>pageSize</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1177713299410"><a name="p1177713299410"></a><a name="p1177713299410"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p197771429244"><a name="p197771429244"></a><a name="p197771429244"></a>分页查询页面大小</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6777029447"><a name="p6777029447"></a><a name="p6777029447"></a>取值为1~100的整数。</p>
</td>
</tr>
<tr id="row10812924556"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p14778192919414"><a name="p14778192919414"></a><a name="p14778192919414"></a>ifCenter</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p0778182919415"><a name="p0778182919415"></a><a name="p0778182919415"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p147785291248"><a name="p147785291248"></a><a name="p147785291248"></a>标志查询节点类型是否为<span id="ph388619511264"><a name="ph388619511264"></a><a name="ph388619511264"></a>MEF Center</span>节点</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p17778152918411"><a name="p17778152918411"></a><a name="p17778152918411"></a>取值为true或false。</p>
<div class="note" id="note2862431999"><a name="note2862431999"></a><a name="note2862431999"></a><span class="notetitle"> > [!NOTE] 说明  </span><div class="notebody"><a name="ul172619261897"></a><a name="ul172619261897"></a><ul id="ul172619261897"><li>当ifCenter指定为<span class="parmvalue" id="parmvalue1231617358148"><a name="parmvalue1231617358148"></a><a name="parmvalue1231617358148"></a>“true”</span>时，将忽略groupId及sn。</li><li>当ifCenter指定为<span class="parmvalue" id="parmvalue1996315374144"><a name="parmvalue1996315374144"></a><a name="parmvalue1996315374144"></a>“false”</span>，且不提供sn和groupId时，分页查询所有<span id="ph16491162049"><a name="ph16491162049"></a><a name="ph16491162049"></a>MEF Edge</span>节点告警。</li></ul>
</div></div>
</td>
</tr>
<tr id="row1274362319416"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p37781229846"><a name="p37781229846"></a><a name="p37781229846"></a>sn</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1977812291141"><a name="p1977812291141"></a><a name="p1977812291141"></a>可选，二选一</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p177785291045"><a name="p177785291045"></a><a name="p177785291045"></a>指定<span id="ph174153115216"><a name="ph174153115216"></a><a name="ph174153115216"></a>MEF Edge</span>设备序列号可查询该节点上的告警信息</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p67781291347"><a name="p67781291347"></a><a name="p67781291347"></a>支持小写字母、大写字母，数字，下划线和连字符；不能以下划线或中划线开头和结尾；最大长度为64字节。</p>
</td>
</tr>
<tr id="row127648241644"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p477811292046"><a name="p477811292046"></a><a name="p477811292046"></a>groupId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p12778132911414"><a name="p12778132911414"></a><a name="p12778132911414"></a>指定groupId可查询该节点组中的所属节点上的告警信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1877852915417"><a name="p1877852915417"></a><a name="p1877852915417"></a>32位无符号数。取值最小为1，最大值为2^32-1。</p>
</td>
</tr>
</tbody>
</table>

> [!NOTE] 说明   
>
>- 当ifCenter、sn、groupId三个参数均为空时，按照pageNum和pageSize分页查询所有节点的告警。
>- 当仅有sn或groupId时，ifCenter默认为“false“。

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/alarms?pageNum=1&pageSize=100&ifCenter=false&sn=xxxxxxxxxxxxxx
```

请求消息体：无

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "records": [
            {
                "id": 1430946159,
                "alarmType": "alarm",
                "createAt": "2023-09-27T16:01:25Z",
                "ip": "xx.xx.xx.xx",
                "serialNumber": "xxxxxxxxxxxxxx",
                "resource": "ALARM DEFAULT RESOURCE",
                "severity": "MINOR"
            },
            {
                "id": 628196293,
                "alarmType": "alarm",
                "createAt": "2023-09-27T16:05:25Z",
                "ip": "xx.xx.xx.xx",
                "serialNumber": "xxxxxxxxxxxxxx",
                "resource": "ALARM DEFAULT RESOURCE",
                "severity": "MAJOR"
            },
            {
                "id": 2136868853,
                "alarmType": "alarm",
                "createAt": "2023-09-27T16:09:25Z",
                "ip": "xx.xx.xx.xx",
                "serialNumber": "xxxxxxxxxxxxxx",
                "resource": "ALARM DEFAULT RESOURCE",
                "severity": "MAJOR"
            }
        ],
        "total": 3
    }
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|records|数组|分页查询的对象数组|
|total|数字|查询结果总数|

**表 4**  records字段说明

|参数|类型|参数说明|
|--|--|--|
|id|数字|告警标识|
|alarmType|字符串|告警类型，取值为alarm|
|createAt|字符串|告警创建时间|
|ip|字符串|设备IP地址|
|serialNumber|字符串|MEF Edge告警为设备序列号，MEF Center告警为空字符串|
|resource|字符串|告警来源|
|severity|字符串|告警等级：<li>MINOR：一般告警</li><li>MAJOR：严重告警</li><li>CRITICAL：紧急告警</li>|

### 查询告警详情<a id="查询告警详情"></a>

**命令功能<a name="section135251624204320"></a>**

查看告警详情，按照指定的告警标识，返回该告警的详细信息。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/alarmmanager/v1/alarm?id=**<i>{value1}</i>

请求消息体：无

**URL参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|id|必选|告警标识|取值最小为1，最大值为2^32-1的整数。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/alarm?id=1430946159
```

请求消息体：无

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "alarmId": "0x00131001",
        "alarmName": "ALARM DEFAULT NAME",
        "alarmType": "alarm",
        "createAt": "2023-09-27T16:01:25Z",
        "detailedInformation": "ALARM DEFAULT INFO",
        "id": 1430946159,
        "impact": "ALARM DEFAULT Impact",
        "ip": "xx.xx.xx.xx",
        "serialNumber": "xxxxxxxxxxxxxx",
        "perceivedSeverity": "MAJOR",
        "reason": "ALARM DEFAULT Reason",
        "resource": "ALARM DEFAULT RESOURCE",
        "suggestion": "ALARM DEFAULT SUGGESTION"
    }
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|alarmId|字符串|告警ID|
|alarmName|字符串|告警名|
|alarmType|字符串|告警类型，取值为alarm|
|createAt|字符串|告警创建时间|
|detailedInformation|字符串|告警详细信息|
|id|数字|告警标识|
|impact|字符串|告警影响描述|
|ip|字符串|设备IP地址|
|serialNumber|字符串|MEF Edge告警为设备序列号，MEF Center告警为空字符串|
|perceivedSeverity|字符串|告警等级：<li>MINOR：一般告警<li>MAJOR：严重告警<li>CRITICAL：紧急告警|
|reason|字符串|告警产生原因|
|resource|字符串|告警来源|
|suggestion|字符串|告警处理建议|

### 查询事件列表<a name="ZH-CN_TOPIC_0000001691744917"></a>

**命令功能<a name="section135251624204320"></a>**

查询系统事件数据，URL参数用于指定本次分页查询控制条件、查询的事件种类和查询操作类型，返回MEF Center数据库中已创建的事件信息。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/alarmmanager/v1/events?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&ifCenter=**_\{value3\}_**&sn=**_\{value4\}_**&groupId=**_\{value5\}_

请求消息体：无

**URL参数<a name="section1774293413516"></a>**

**表 1**  参数说明

<table><thead align="left"><tr id="row1643491410534"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p1943421410537"><a name="p1943421410537"></a><a name="p1943421410537"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p113476328527"><a name="p113476328527"></a><a name="p113476328527"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p343413143534"><a name="p343413143534"></a><a name="p343413143534"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p18434201414535"><a name="p18434201414535"></a><a name="p18434201414535"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row11236115220615"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p207789294419"><a name="p207789294419"></a><a name="p207789294419"></a>pageNum</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1477862910411"><a name="p1477862910411"></a><a name="p1477862910411"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1977814296415"><a name="p1977814296415"></a><a name="p1977814296415"></a>分页查询取值页数，为序数词</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p187782291042"><a name="p187782291042"></a><a name="p187782291042"></a>取值为1~2^31-1的整数。</p>
</td>
</tr>
<tr id="row19711641046"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1777710299410"><a name="p1777710299410"></a><a name="p1777710299410"></a>pageSize</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1177713299410"><a name="p1177713299410"></a><a name="p1177713299410"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p197771429244"><a name="p197771429244"></a><a name="p197771429244"></a>分页查询页面大小</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6777029447"><a name="p6777029447"></a><a name="p6777029447"></a>取值为1~100的整数。</p>
</td>
</tr>
<tr id="row10812924556"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p14778192919414"><a name="p14778192919414"></a><a name="p14778192919414"></a>ifCenter</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p0778182919415"><a name="p0778182919415"></a><a name="p0778182919415"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p147785291248"><a name="p147785291248"></a><a name="p147785291248"></a>标志查询节点类型是否为<span id="ph388619511264"><a name="ph388619511264"></a><a name="ph388619511264"></a>MEF Center</span>节点</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p17778152918411"><a name="p17778152918411"></a><a name="p17778152918411"></a>取值为true或false。</p>
<div class="note" id="note2862431999"><a name="note2862431999"></a><a name="note2862431999"></a><span class="notetitle">  > [!NOTE] 说明  </span><div class="notebody"><a name="ul172619261897"></a><a name="ul172619261897"></a><ul id="ul172619261897"><li>当ifCenter指定为<span class="parmvalue" id="parmvalue1231617358148"><a name="parmvalue1231617358148"></a><a name="parmvalue1231617358148"></a>“true”</span>时，将忽略groupId及sn。</li><li>当ifCenter指定为<span class="parmvalue" id="parmvalue1996315374144"><a name="parmvalue1996315374144"></a><a name="parmvalue1996315374144"></a>“false”</span>，且不提供sn和groupId时，分页查询所有<span id="ph16491162049"><a name="ph16491162049"></a><a name="ph16491162049"></a>MEF Edge</span>节点告警。</li></ul>
</div></div>
</td>
</tr>
<tr id="row1274362319416"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p37781229846"><a name="p37781229846"></a><a name="p37781229846"></a>sn</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1977812291141"><a name="p1977812291141"></a><a name="p1977812291141"></a>可选，二选一</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p177785291045"><a name="p177785291045"></a><a name="p177785291045"></a>指定<span id="ph174153115216"><a name="ph174153115216"></a><a name="ph174153115216"></a>MEF Edge</span>设备序列号可查询该节点上的告警信息</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p67781291347"><a name="p67781291347"></a><a name="p67781291347"></a>支持小写字母、大写字母，数字，下划线和连字符；不能以下划线或中划线开头和结尾；最大长度为64字节。</p>
</td>
</tr>
<tr id="row127648241644"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p477811292046"><a name="p477811292046"></a><a name="p477811292046"></a>groupId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p12778132911414"><a name="p12778132911414"></a><a name="p12778132911414"></a>指定groupId可查询该节点组中的所属节点上的告警信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1877852915417"><a name="p1877852915417"></a><a name="p1877852915417"></a>32位无符号数。取值最小为1，最大值为2^32-1。</p>
</td>
</tr>
</tbody>
</table>

> [!NOTE] 说明   
>
>- 当ifCenter、sn、groupId三个参数均为空时，按照pageNum和pageSize分页查询所有节点的事件。
>- 当仅有sn或groupId时，ifCenter默认为“false“。

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/events?pageNum=1&pageSize=100&ifCenter=false&groupId=1
```

请求消息体：无

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "records": [
            {
                "id": 1430946159,
                "alarmType": "event",
                "createAt": "2023-09-27T16:01:25Z",
                "ip": "xx.xx.xx.xx",
                "serialNumber": "xxxxxxxxxxxxxx",
                "resource": "ALARM DEFAULT RESOURCE",
                "severity": "MAJOR"
            },
             {
                "id": 628196293,
                "alarmType": "event",
                "createAt": "2023-09-27T16:01:25Z",
                "ip": "xx.xx.xx.xx",
                "serialNumber": "xxxxxxxxxxxxxx",
                "resource": "ALARM DEFAULT RESOURCE",
                "severity": "MAJOR"
            }
        ],
        "total": 2
    }
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|records|数组|分页查询的对象数组|
|total|数字|查询结果总数|

**表 4**  records字段说明

|参数|类型|参数说明|
|--|--|--|
|id|数字|事件序号|
|alarmType|字符串|事件类型，取值为event|
|createAt|字符串|事件创建时间|
|ip|字符串|设备IP地址|
|serialNumber|字符串|MEF Edge事件为设备序列号，MEF Center事件为空字符串|
|resource|字符串|该事件来源|
|severity|字符串|事件等级：<li>MINOR：一般事件</li><li>MAJOR：严重事件</li><li>CRITICAL：紧急事件</li><li>OK：正常事件</li>|

### 查询事件详情<a name="ZH-CN_TOPIC_0000001643632794"></a>

**命令功能<a name="section135251624204320"></a>**

查看事件详情，按照指定的事件序号，返回该事件的详细信息。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL**：**https://**_\{ip\}:\{port\}_**/alarmmanager/v1/event?id=**<i>{value1}</i>

请求消息体：无

**URL参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|id|必选|事件序号|取值最小为1，最大值为2^32-1的整数。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/event?id=1430946159
```

请求消息体：无

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "alarmId": "40eafda7-bce6-4850-9e35-949fc81b50bf",
        "alarmName": "ALARM DEFAULT NAME",
        "alarmType": "event",
        "createAt": "2023-09-27T16:01:25Z",
        "detailedInformation": "ALARM DEFAULT INFO",
        "id": 1430946159,
        "ip": "xx.xx.xx.xx",
        "impact": "ALARM DEFAULT Impact",
        "serialNumber": "xxxxxxxxxxxxxx",
        "perceivedSeverity": "MINOR",
        "reason": "ALARM DEFAULT Reason",
        "resource": "ALARM DEFAULT RESOURCE",
        "suggestion": "ALARM DEFAULT SUGGESTION"
    }
}
```

响应状态码：200

**输出说明<a name="section99341441896"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|查询结果|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|alarmId|字符串|事件ID|
|alarmName|字符串|事件名|
|alarmType|字符串|事件类型，取值为event|
|createAt|字符串|事件创建时间|
|detailedInformation|字符串|事件详细信息|
|id|数字|事件序号|
|ip|字符串|设备IP地址|
|impact|字符串|事件影响描述|
|serialNumber|字符串|MEF Edge事件为设备序列号，MEF Center事件为空字符串|
|perceivedSeverity|字符串|事件等级：<li>OK：正常事件<li>MINOR：一般事件<li>MAJOR：严重事件<li>CRITICAL：紧急事件|
|reason|字符串|事件产生原因|
|resource|字符串|事件来源|
|suggestion|字符串|事件处理建议|

## 配置接口<a id="ZH-CN_TOPIC_0000001526721288"></a>

### 配置接口介绍<a id="配置接口介绍"></a>

配置接口用于支持用户配置第三方软件仓库和镜像仓库的根证书，以及配置镜像下载信息。用户使用第三方镜像仓库需要先调用[导入根证书](#导入根证书)接口导入镜像仓库证书，再通过调用[配置镜像下载信息](#配置镜像下载信息)接口配置镜像下载信息。

### 导入根证书<a id="导入根证书"></a>

**命令功能<a name="section135251624204320"></a>**

用于支持第三方软件仓库和镜像仓库，导入对应的根证书。使用容器应用前需要导入软件仓库和镜像仓库的根证书。

> [!NOTE] 说明     
>
>- 重复调用该接口会更新根证书。
>- 根证书的有效期建议大于[证书告警的检测周期](./common_operations.md#mef-center配置和查询证书过期告警)（默认值为7）。
>- 重复导入根证书会备份之前导入的前一份证书。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/certmanager/v1/certificates/import**

请求消息体：

```json
{
    "certName": certname,
    "cert": cert
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|certName|字符串|导入证书的用途|取值为以下参数之一。<li>software：软件仓库根证书<li>image：镜像仓库根证书|
|cert|字符串|base64编码格式的PEM根证书|<li>必须是base64编码后的根证书。<li>证书需要是PEM格式。<li>根CA证书中签名正确。<li>根CA证书处于有效期内。<li>证书需为X.509 V3数字证书，根CA证书的“基本限制”扩展域须标明为“CA”，“密钥用法”扩展域中须包含“证书签名”。</li><li>密钥要求密钥算法为RSA算法，长度不小于3072；或ECDSA算法，长度不小于256。摘要算法需为SHA256、SHA384、SHA512。</li>|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/certmanager/v1/certificates/import
```

请求消息体：

```json
{
    "certName": "software",
    "cert": "xxxxxxxxxxxxxxxxxxx..."
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "import certificate success"
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|

### 配置镜像下载信息<a id="配置镜像下载信息"></a>

**命令功能<a name="section135251624204320"></a>**

用于配置第三方镜像仓库地址和账号密码，仓库服务器地址支持域名或者IP地址。重复调用接口时，会更新已有的镜像下载信息配置。

> [!NOTICE] 须知  
>
>- 镜像下载信息将会被MEF Center传递到K8s和KubeEdge中，后者会将数据保存在K8s中和MEF Edge设备的edgecore数据库中。用户可根据需要，通过定制K8s和KubeEdge的方式，对镜像仓库的账号密码进行安全加固。
>- 建议用户使用可信的第三方镜像仓库。如果配置了不安全的第三方镜像仓库地址，可能会存在不安全的传输过程。

> [!NOTE] 说明   
> 当使用域名时，需要在边缘设备MEF Edge的主机目录“/etc/hosts“中，配置域名和IP地址的映射关系，具体操作请参考[配置本地域名映射](./common_operations.md#配置本地域名映射)。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/image/config**

请求消息体：

```json
{
    "domain": domain
    "ip": ip,
    "port": port,
    "account": account,
    "password": password
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值|
|--|--|--|--|
|domain|可选|镜像仓库服务器的域名|字符串，合法的域名。支持长度3~63位，大小写字母、数字和符号（.-）的组合，且只能以大小写字母、数字开头和结尾，不能为全数字。取值不能为localhost。<br>domain和IP参数至少需要传入一个。同时提供两个参数时，以域名为准。|
|ip|可选|镜像仓库服务器的IP地址|字符串，合法的IPv4地址，不能为全零或者全255，不能为回环地址127.0.0.1，且不能为MEF Edge设备主机地址。domain和IP参数至少需要传入一个。同时提供两个参数时，以域名为准。|
|port|必选|镜像仓库对外提供服务的端口号|数字，只能为1-65535之间的整数。|
|account|必选|镜像下载账号|字符串，长度最大256个字符。支持大小写字母、数字、下划线（_）、短横线（-），下划线和短横线不能在开头结尾。|
|password|必选|镜像下载密码|字节数组，数组长度为[8，20]。镜像仓库密码不支持英文冒号。<br> > [!NOTE] 说明 <br>密码复杂度建议满足如下要求，若设置的密码不符合以下规则，可能存在安全风险。<li>密码长度至少需要8个字符。<li>密码必须包含如下至少两种字符的组合：<ul><li>至少一个小写字母。<li>至少一个大写字母。<li>至少一个数字。<li>至少一个特殊字符：`~!@#$%^&*()-_=+\|[{}];:'",<.>/?和空格</ul><li>密码不能和账号一样。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/image/config
```

请求消息体：

```json
{
    "domain": "xxx.huawei.com",
    "ip": "10.10.10.10",
    "port": 6443,
    "account": "ImageRepository",
    "password": [72, 117, 97, 119, 101, 105, 49, 50, 35, 36]
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|

### 删除根证书<a name="ZH-CN_TOPIC_0000001526881244"></a>

**命令功能<a name="section135251624204320"></a>**

用于删除第三方的软件仓库和镜像仓库已导入的根证书。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/certmanager/v1/certificates/delete-cert**

请求消息体：

```json
{
    "type": type
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  请求参数

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|type|必选|证书的类型|取值为software、image之一。<li>software：软件仓库根证书</li><li>image：镜像仓库根证书</li>|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/certmanager/v1/certificates/delete-cert
```

请求消息体：

```json
{
    "type": "software"
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "delete ca file success"
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|

### 导出根证书<a id="ZH-CN_TOPIC_0000001526721348"></a>

**命令功能<a name="section135251624204320"></a>**

用于导出MEF Center的根证书。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_<b>/certmanager/v1/export?certName=</b>hub\_svr

**URL参数<a name="section1774293413516"></a>**

**表 1**  URL参数

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|certName|必选|导出根证书目标类型|目前只支持hub_svr类型导出。hub_svr：MEF Center根证书|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/certmanager/v1/export?certName=hub_svr
```

成功响应样例：输出root.crt文件

```text
-----BEGIN CERTIFICATE-----
xxxxxxxxxxxxxxxxxxx...
-----END CERTIFICATE-----
```

### 获取云边认证token<a id="ZH-CN_TOPIC_0000001566531326"></a>

**命令功能<a name="section135251624204320"></a>**

用于获取MEF Center和MEF Edge进行云边认证的token。

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_**/edgemanager/v1/token**

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/token
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "export token success",
    "data": "xxxxxxxxxxxxxxxxxxxx..."
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 1**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|字符串|云边认证token，有效期为7天，到期后自动过期|

### 获取集成方证书信息<a id="ZH-CN_TOPIC_0000001621432513"></a>

**命令格式<a name="section6901955114320"></a>**

操作类型：**GET**

**URL：https://**_\{ip\}:\{port\}_**/certmanager/v1/certificates/info?certName=**_\{__certName__\}_

**请求参数<a name="section1774293413516"></a>**

**表 1**  获取集成方证书信息参数说明

|参数|是否必选|参数说明|取值|
|--|--|--|--|
|certName|必选|证书名|字符串，当前只支持north。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/certmanager/v1/certificates/info?certName=north
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": [
    {
        "FingerPrint": "xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx",
        "FingerPrintAlgorithm": "sha256",
        "Issuer": "xxxxxxxxxxxx",
        "SerialNumber": "xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx",
        "Subject": "xxxxxxxxxxxx",
        "Validity": 
        {
              "NotAfter": "2033-03-31 08:41:58",
              "NotBefore": "2023-04-03 08:41:58"
         }
     }
     ]
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|集成方证书信息|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|FingerPrint|字符串|证书指纹|
|FingerPrintAlgorithm|字符串|证书指纹算法|
|Issuer|字符串|证书签发者|
|SerialNumber|字符串|证书序列号|
|Subject|字符串|证书持有者|
|Validity|对象|证书有效性，包含NotBefore和NotAfter|
|NotAfter|字符串|有效期截止时间|
|NotBefore|字符串|有效期开始时间|

### 导入吊销列表<a id="ZH-CN_TOPIC_0000001621540625"></a>

**命令功能<a name="section135251624204320"></a>**

用于导入与MEF Center对接的业务平台、软件仓库、镜像仓库等根证书的对应吊销列表链，可取消被吊销的对接第三方平台证书的RESTful请求访问权限。重复调用此接口会更新导入吊销列表。导入对接业务平台吊销列表后，需要用户手动重启MEF Center后生效。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**URL：https://**_\{ip\}:\{port\}_**/certmanager/v1/crl/import**

请求消息体：

```json
{
    "crlName": crlName,
    "crl": crl
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|crlName|字符串|导入吊销列表的用途|<li>north：集成方证书吊销列表<li>software：软件仓库根证书吊销列表<li>image：镜像仓库证书吊销列表|
|crl|字符串|base64编码的PEM格式的证书吊销列表链|<li>必须是base64编码后的CRL。<li>CRL需要是PEM格式。<li>CRL必须和crlName对应的证书链拥有相同的级数。<li>CRL证书处于有效期内。<li>CRL必须包含导入的crlName对应的证书链的每一级证书所签出的吊销列表。|

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/certmanager/v1/crl/import
```

请求消息体：

```json
{
    "crlName": software,
    "crl": crl
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success"
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|

## 升级接口<a id="ZH-CN_TOPIC_0000001527041092"></a>

### 软件升级接口介绍<a name="ZH-CN_TOPIC_0000001577601037"></a>

MEF支持5.0.RC2及以上版本通过MEF Center软件升级接口进行MEF Edge的在线升级、同版本升级和版本回退。

**升级前准备<a name="section84371423133717"></a>**

1. 准备待升级的软件。在线升级需要通过可正常使用的第三方软件仓库准备待升级软件，用户需要确保MEF Edge设备和软件仓库之间的网络连接，并保证下发软件下载消息时指定的https网址可获取到待升级的软件。软件仓库准备及对接具体请参见章节[对接软件仓库](./usage.md#对接软件仓库)和[对接镜像仓库](./usage.md#对接镜像仓库)。
2. （可选）查询节点详情。软件升级时，请求中需涉及到边缘设备的设备序列号，若用户不清楚设备序列号，可通过RESTful接口查询节点详情，查看对应节点的设备序列号。具体请参见[查询节点详情](#查询节点详情)章节。

**升级流程介绍<a name="section131291530113720"></a>**

MEF Edge软件调用接口升级的流程示例如下。

1. 下发软件下载消息

    通过RESTful接口下发软件下载消息，触发边缘设备下载待升级软件。下发软件下载消息接口请参见[下载软件](#下载软件)。

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/download
    ```

2. （可选）查询软件下载进度

    下发软件下载消息时，可通过RESTful接口查询软件下载的进度。查询软件下载进度接口请参见[查询软件下载进度](#查询软件下载进度)。

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/download-progress?serialNumber={value}
    ```

3. （可选）查询软件信息

    触发软件下载后，可通过RESTful接口查询当前MEF Edge软件信息，包括软件当前版本和待升级版本。查询软件信息接口请参见[查询软件信息](#查询软件信息)。

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/version-info?serialNumber={value}
    ```

4. 下发软件升级消息

    通过RESTful接口下发软件升级消息，触发边缘设备升级MEF Edge软件。下发软件升级消息接口请参见[升级软件](#升级软件)。

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/upgrade
    ```

### 下载软件<a id="下载软件"></a>

**命令功能<a name="section135251624204320"></a>**

下发下载MEF Edge软件消息进行软件下载，配置第三方软件仓库地址和账号密码，通过第三方软件仓库下载已准备好的待升级软件。

**命令格式<a name="section6901955114320"></a>**

操作类型：**POST**

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/download**

请求消息体：

```json
{
    "serialNumbers": ["2102312NSF10K8000130"],
    "softwareName": "MEFEdge",
    "downloadInfo": {
        "package": "GET https://xxx.tar.gz",
        "userName": "FileTransferAccount",
        "password": [xx,yy,zz,ww]
    }
}
```

**请求参数<a name="section1774293413516"></a>**

**表 1**  参数说明

<table><thead align="left"><tr id="row153763524260"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p837615213264"><a name="p837615213264"></a><a name="p837615213264"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p15376205202612"><a name="p15376205202612"></a><a name="p15376205202612"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p193761052142613"><a name="p193761052142613"></a><a name="p193761052142613"></a>参数说明</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p3376195215262"><a name="p3376195215262"></a><a name="p3376195215262"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row137821209392"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p184721335322"><a name="p184721335322"></a><a name="p184721335322"></a>serialNumbers</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p17831206397"><a name="p17831206397"></a><a name="p17831206397"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1568916172332"><a name="p1568916172332"></a><a name="p1568916172332"></a>设备序列号</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1088711633416"><a name="p1088711633416"></a><a name="p1088711633416"></a>数组，支持小写字母、大写字母，数字，下划线和连字符。开头和结尾不能包含下划线和中划线，最大长度为64字节。数组长度为[1,2048]。</p>
</td>
</tr>
<tr id="row183821834123910"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p144727383214"><a name="p144727383214"></a><a name="p144727383214"></a>softwareName</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1538243403917"><a name="p1538243403917"></a><a name="p1538243403917"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p8689161717331"><a name="p8689161717331"></a><a name="p8689161717331"></a>待下载的软件名</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p48161222151911"><a name="p48161222151911"></a><a name="p48161222151911"></a>必须为<span class="parmvalue" id="parmvalue361983132513"><a name="parmvalue361983132513"></a><a name="parmvalue361983132513"></a>“MEFEdge”</span>。</p>
</td>
</tr>
<tr id="row945705083116"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1845819504311"><a name="p1845819504311"></a><a name="p1845819504311"></a>downloadInfo</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p745865043119"><a name="p745865043119"></a><a name="p745865043119"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1745875043114"><a name="p1745875043114"></a><a name="p1745875043114"></a>下载信息</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p194581550153117"><a name="p194581550153117"></a><a name="p194581550153117"></a>软件仓库的文件下载信息。</p>
</td>
</tr>
<tr id="row4160114413115"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p164722031320"><a name="p164722031320"></a><a name="p164722031320"></a>package</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p161611944183113"><a name="p161611944183113"></a><a name="p161611944183113"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p16689141763319"><a name="p16689141763319"></a><a name="p16689141763319"></a>软件文件包</p>
<p id="p1068981783318"><a name="p1068981783318"></a><a name="p1068981783318"></a></p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p159771941173311"><a name="p159771941173311"></a><a name="p159771941173311"></a>字符串，URL最大长度为512字节，不能包含"\n!\\|;$&lt;&gt;@` "等特殊字符。其中协议必须为https，请求方式为GET。</p>
<div class="note" id="note134513320143"><a name="note134513320143"></a><a name="note134513320143"></a><span class="notetitle">  > [!NOTE] 说明  </span><div class="notebody"><a name="ul17843314116"></a><a name="ul17843314116"></a><ul id="ul17843314116"><li>若包含域名，需要为合法的域名，格式符合正则模式^[a-zA-Z0-9][a-zA-Z0-9.-]{1,61}[a-zA-Z0-9]$，取值不能为全数字，且不能为localhost。</li><li>若包含IP，需要为合法的IPv4地址，不能为回环地址（127.0.0.1）或<span id="ph1564518021910"><a name="ph1564518021910"></a><a name="ph1564518021910"></a>MEF Edge</span>节点IP。</li></ul>
</div></div>
</td>
</tr>
<tr id="row32121849203113"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p34721736323"><a name="p34721736323"></a><a name="p34721736323"></a>userName</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1121374911312"><a name="p1121374911312"></a><a name="p1121374911312"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1068981714338"><a name="p1068981714338"></a><a name="p1068981714338"></a>下载账号</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p10213154963115"><a name="p10213154963115"></a><a name="p10213154963115"></a>字符串，长度为6~32；只能由小写字母、大写字母、数字组成。</p>
</td>
</tr>
<tr id="row2851651163112"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1047212318323"><a name="p1047212318323"></a><a name="p1047212318323"></a>password</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p138565123119"><a name="p138565123119"></a><a name="p138565123119"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1168921714333"><a name="p1168921714333"></a><a name="p1168921714333"></a>下载密码</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p208535118313"><a name="p208535118313"></a><a name="p208535118313"></a>byte数组，下发信息中必须包含下载密码。数组长度为[8,20]。</p>
<div class="note" id="note107555422113"><a name="note107555422113"></a><a name="note107555422113"></a><span class="notetitle"> > [!NOTE] 说明    </span><div class="notebody"><div class="p" id="p82854163268"><a name="p82854163268"></a><a name="p82854163268"></a>密码复杂度建议满足如下要求，若设置的密码不符合以下规则，可能存在安全风险。<a name="ul753618188215"></a><a name="ul753618188215"></a><ul id="ul753618188215"><li>密码长度至少需要8个字符。</li><li>密码必须包含如下至少两种字符的组合：<a name="ul682642514227"></a><a name="ul682642514227"></a><ul id="ul682642514227"><li>至少一个小写字母。</li><li>至少一个大写字母。</li><li>至少一个数字。</li><li>至少一个特殊字符：`~!@#$%^&amp;*()-_=+\|[{}];:'",&lt;.&gt;/?和空格。</li></ul>
</li><li>密码不能和账号一样。</li></ul>
</div>
</div></div>
</td>
</tr>
</tbody>
</table>

**使用样例<a name="section9299576112"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/software/edge/download
```

请求体：

```json
{
    "serialNumbers": ["xxxxxxxxx"],
    "softwareName": "MEFEdge",
     "downloadInfo": {
        "package": "GET https://xxx.tar.gz",
        "userName": "FileTransferAccount",
        "password": [xx,yy,zz,ww]
    }
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "failedInfos": {},
        "successIDs": [
            "xxxxxxxxx"
        ]
    }
}
```

响应状态码：200

**输出说明<a name="section127921251728"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|-|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|字符串列表|成功的序列号列表|
|failedInfos|哈希表，key和value的类型都为字符串|key值为失败的序列号，value为此序列号失败原因|

### 查询软件下载进度<a id="查询软件下载进度"></a>

**命令功能<a name="section19727974820"></a>**

查询下载进度。

**命令格式<a name="section101971341114914"></a>**

操作类型：**GET**

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/download-progress?serialNumber=**_\{value\}_

**URL参数<a name="section12700104818484"></a>**

**表 1**  URL参数

|参数|类型|参数说明|取值要求|
|--|--|--|--|
|serialNumber|必选|设备序列号|字符串，支持大小写字母、下划线，中划线，开头和结尾不能包含下划线和中划线，且以大小字母数字开头和结尾，最大长度为64字节。|

**使用样例<a name="section3878166125016"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/software/edge/download-progress?serialNumber=2102312NSF10K8000130
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "msg": "",
        "progress": 50,
        "res": "success"
    }
}
```

响应状态码：200

**输出说明<a name="section1390714110514"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|-|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|msg|字符串|具体状态信息|
|progress|uint64|下载进度值|
|res|字符串|下载结果，取值为success或failed|

### 查询软件信息<a id="查询软件信息"></a>

**命令功能<a name="section82653133719"></a>**

查询软件信息。

**命令格式<a name="section16835327372"></a>**

操作类型：**GET**

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/version-info?serialNumber=**_\{value\}_

**URL参数<a name="section1130318261113"></a>**

**表 1**  URL参数

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|serialNumber|必选|设备序列号|字符串，只能支持小写字母、大写字母、数字，下划线和中划线。开头和结尾不能包含下划线和中划线，最大长度为64字节。|

**使用样例<a name="section569514203919"></a>**

请求样例：

```bash
GET https://10.10.10.10:30035/edgemanager/v1/software/edge/version-info?serialNumber=2102312NSF10K8000130
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": [
        {
            "InactiveVersion": "",
            "Name": "MEFEdge",
            "Version": "x.x.xxx"
        }
    ]
}
```

响应状态码：200

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|-|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|InactiveVersion|字符串|待生效版本信息|
|Name|字符串|待生效软件名|
|Version|字符串|当前运行的软件版本|

### 升级软件<a id="升级软件"></a>

**命令功能<a name="section1612804915619"></a>**

下发升级MEF Edge软件消息进行软件升级。

**命令格式<a name="section121284496614"></a>**

操作类型：POST

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/upgrade**

请求消息体：

```json
{
    "SerialNumbers": ["xxxxxxxxxxxxx"],
    "softwareName": "MEFEdge"
}
```

**请求参数<a name="section263114675317"></a>**

**表 1**  参数说明

|参数|是否必选|参数说明|取值要求|
|--|--|--|--|
|SerialNumbers|必选|设备序列号|数组，支持小写字母、大写字母、数字，下划线和中划线。开头和结尾不能包含下划线和中划线，最大长度为64字节。数组长度为[1,2048]。|
|softwareName|必选|待下载的软件名|取值为MEFEdge。|

**使用样例<a name="section9550421315"></a>**

请求样例：

```bash
POST https://10.10.10.10:30035/edgemanager/v1/software/edge/upgrade
```

请求消息体：

```json
{
    "SerialNumbers": ["xxxxxxxxxxxxx"],
    "softwareName": "MEFEdge"
}
```

响应样例：

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "failedInfos": {},
        "successIDs": [
            "xxxxxxxxxxxxxx"
        ]
    }
}
```

响应状态码：200

**输出说明<a name="section851619548567"></a>**

**表 2**  操作输出说明

|参数|类型|参数说明|
|--|--|--|
|status|字符串|错误码|
|msg|字符串|描述信息|
|data|对象|-|

**表 3**  data字段说明

|参数|类型|参数说明|
|--|--|--|
|successIDs|字符串列表|成功的序列号列表|
|failedInfos|哈希表，key和value的类型都为字符串|key值为失败的序列号，value为此序列号失败原因|

## 错误码说明<a name="ZH-CN_TOPIC_0000001526721244"></a>

### 错误码取值范围说明<a name="ZH-CN_TOPIC_0000001582102534"></a>

**表 1**  错误码取值范围说明

|错误码区间|说明|
|--|--|
|0000 0000|公共部分|
|4000 0000|edge-manager模块<li>4001 0000：节点管理接口</li><li>4002 0000：容器应用管理接口</li>|
|5000 0000|alarm-manager模块|
|6000 0000|cert-manager模块|

### 公共部分<a name="ZH-CN_TOPIC_0000001577600997"></a>

**表 1**  公共部分错误码说明

|错误码|说明|
|--|--|
|00000000|请求成功|
|00001001|解析请求体失败|
|00001002|获取请求接口结果失败|
|00001003|向模块发起同步消息失败|
|00001004|请求路由转发失败|
|00001005|参数校验失败|
|00001006|请求结构体转化失败|
|00001007|请求体类型判断失败|
|00001008|创建请求消息失败|

### 节点接口<a name="ZH-CN_TOPIC_0000001527201156"></a>

**表 1**  节点管理错误码说明

|错误码|说明|
|--|--|
|00000000|请求成功|
|40011000|节点管理规格校验失败|
|40012000|数据已存在|
|40012001|创建节点组接口失败|
|40012002|获取节点组列表失败|
|40012003|获取节点组详情失败|
|40012004|修改节点组信息失败|
|40012005|获取节点组数量失败|
|40012006|删除节点组失败|
|40012007|获取节点详情失败|
|40012008|修改节点信息失败|
|40012009|根据状态获取节点数量失败|
|40012010|获取已纳管节点列表失败|
|40012011|获取未纳管节点列表失败|
|40012012|向节点组添加节点错误|
|40012013|纳管节点出错|
|40012014|删除节点失败|
|40012015|从节点组中删除节点失败|
|40012016|向节点发送信息失败|
|40012017|获取token失败|
|40012018|查询软件信息时获取节点上软件版本号失败|

### 容器应用接口<a name="ZH-CN_TOPIC_0000001577401145"></a>

**表 1**  应用管理错误码说明

|错误码|说明|
|--|--|
|00000000|请求成功|
|40021000|应用管理规格校验失败|
|40021001|参数转化失败|
|40021002|解析容器参数失败|
|40022000|数据已存在|
|40022001|数据不存在|
|40022002|创建应用失败|
|40022003|查询应用详情失败|
|40022004|查询应用列表失败|
|40022005|部署应用失败|
|40022006|卸载应用失败|
|40022007|更新应用失败|
|40022008|删除应用失败|
|40022009|根据应用查询容器实例失败|
|40022010|根据节点查询容器实例失败|
|40022011|查询所有实例列表失败|
|40022012|查询节点组下的实例数量失败|

### 日志收集接口<a name="ZH-CN_TOPIC_0000001640828986"></a>

**表 1**  日志收集接口错误码说明

|错误码|说明|
|--|--|
|40052002|日志导出业务异常|
|40051002|日志导出获取节点信息异常|

### 告警信息接口<a name="ZH-CN_TOPIC_0000001643483134"></a>

**表 1**  告警接口错误码说明

|错误码|说明|
|--|--|
|00000000|请求成功|
|50011001|查询MEF Center节点告警或事件失败|
|50011002|查询MEF Edge节点告警或事件列表失败|
|50011003|按节点组查询告警或事件失败|
|50011004|查询告警或事件失败|
|50011005|查询节点组详情失败|
|50011006|解析节点组详情数据失败|
|50011007|获取告警或事件详情失败|

### 配置接口<a name="ZH-CN_TOPIC_0000001527041140"></a>

**表 1**  配置接口错误码说明

|错误码|说明|
|--|--|
|00000000|请求成功|
|60001001|查询指定类别的根证书失败|
|60001002|签发根证书失败|
|60001003|校验根证书内容失败|
|60001004|保存根证书失败|
|60001005|删除根证书失败|
|60001006|分发根证书到边侧失败|
|60001007|获取云侧secret失败|
|60001008|配置镜像下载信息失败|
|60001009|导出根证书失败|
|60001010|获取证书信息失败|
|60001011|导入吊销列表失败|
|60001012|alarm-manager获取导入证书的信息失败|
|60002001|获取云边认证token失败|
