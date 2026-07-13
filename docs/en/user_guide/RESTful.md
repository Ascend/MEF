# API Reference<a name="ZH-CN_TOPIC_0000001674256246"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:20:03.188Z pushedAt=2026-06-09T01:46:27.916Z -->

Users can use MEF-related functions by calling the RESTful APIs provided by MEF Center.

## Introduction<a name="ZH-CN_TOPIC_0000001527201136"></a>

This document describes the RESTful APIs of MEF Center in detail, guiding users to use MEF functions by calling RESTful APIs.

Redfish is a management standard based on HTTPS services, using RESTful APIs to manage devices. Each HTTPS operation submits or returns a resource in the form of UTF-8 encoded JSON. Just as a web application returns HTML to a browser, a RESTful API returns data to the client in JSON format through the same transport mechanism (HTTPS). Parameter names passed in JSON format in this document are case-insensitive.

> [!NOTE]
> All RESTful APIs provided by MEF only support serial calls.

**Response Status Code<a name="zh-cn_topic_0000001082477678_zh-cn_topic_0178823233_section37953311"></a>**

**Table 1** Status codes

|Status Code|Description|
|--|--|
|200|Request succeeded.|
|400|Bad Request. An error occurred on the client side and an error message is returned.|
|401|Unauthorized. The request requires user authentication.|
|404|Not Found. The requested resource does not exist.|
|405|Method Not Allowed. The request method is invalid, for example, using an incorrect method type for the request.|
|423|Locked. The current resource is locked.|
|429|Too Many Requests. Too many requests have been sent, exceeding the frequency limit.|
|499|Client Closed Request. The client disconnected.|
|503|Service Unavailable. The server is currently unable to process the request due to maintenance or overload.|

**Base URL Description<a name="section20778152719278"></a>**

All API calls are based on the unified Base URL format `https://{ip}:{port}`. For specific API paths, see the detailed description of each API.

**Table 2** Base URL composition description

|Name|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|ip|Yes|IP address for logging in to the device|IPv4 address.|
|port|Yes|Port number for calling a specific module API|-|

**Request Header Parameters<a name="section13745311172616"></a>**

The request header parameter involved in this document is described as follows.

**Table 3** Request headers

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|Content-Type|Yes|Media type|String, the value is `application`or `json`, carried in the request header.|

## Feature Description<a name="ZH-CN_TOPIC_0000001526721304"></a>

The current MEF Center includes the nginx-manager module, edge-manager module, cert-manager module, and alarm-manager module. _API Reference_ introduces the APIs according to the functions provided by each module.

- The nginx-manager module provides the APIG service, implementing functions such as receiving external access, rate limiting for northbound APIs, and forwarding.
- The edge-manager module includes node management APIs and containerized application management APIs, enabling management of edge nodes and the containerized applications running on them.
  - Supports creating, querying, modifying, and deleting node groups.
  - Supports managing, adding, deleting, and querying nodes.
  - Supports creating, querying, updating, deleting, and deploying containerized applications.

- The cert-manager module includes configuration APIs, which support importing, querying, and deleting root certificates, as well as issuing certificates.
- The alarm-manager module supports querying alarm or event information.

## Component Version APIs<a name="ZH-CN_TOPIC_0000001526881264"></a>

### Querying the edge-manager Version<a id="ZH-CN_TOPIC_0000001577280961"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the edge-manager version.

**Syntax<a name="section6901955114320"></a>**

Operation type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/version**

**Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/version
```

Response example:

```json
{
    "status": "00000000",
    "msg": "success",
    "data": "7.0.RC1"
}
```

Response status code: 200

**Output Description<a name="section127921251728"></a>**

**Table 1**  Operation output description

|Field|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|String|Current version number of edge-manager|

## Node Management APIs<a name="ZH-CN_TOPIC_0000001526881140"></a>

### Overview<a id="ZH-CN_TOPIC_0000001577280885"></a>

For nodes that join the same group, containerized applications can be managed in batches, including deploying, updating, and uninstalling containerized applications. Nodes and node groups use IDs as unique identifiers, which are automatically generated. Fields such as `Node ID`, `NodeName`, `UniqueName`, and `SerialNumber` must not be duplicated.

**Constraints<a name="section1127451245510"></a>**

- The maximum number of supported node groups is 1024.
- The maximum number of supported edge nodes is 1024.
- A single node group can contain a maximum of 1024 edge nodes.
- A single edge node can join a maximum of 10 node groups.
- For unmanaged nodes, update operations other than [Querying the List of Nodes Not Managed by MEF](#querying-the-list-of-nodes-not-managed-by-mef) and [Deleting Nodes Not Managed by MEF](#deleting-nodes-not-managed-by-mef) are not allowed.

**Node Status<a name="section153111018349"></a>**

- `ready`: The node has been connected to K8s, and is ready. The node is normal.
- `notReady`: The node has been connected to K8s, but is not ready. The node status is abnormal.
- `unknown`: The connection between the node and K8s is interrupted. The node status is unknown.
- `offline`: The node is not found within the K8s cluster, but is found in the MEF Center database. The node is offline.
- `abnormal`: The connection between the node and MEF Center is abnormal.

**Edge Node Management Process<a name="section433103413267"></a>**

The management operation includes onboarding nodes in the cluster under the management of the MEF system. MEF Center manages edge nodes that have been successfully onboarded. MEF Center does not perform task operations on unmanaged nodes. When using the APIs, users can first create a node group and add nodes to the group while managing them. Alternatively, they can manage nodes individually first without adding them to any node group, and then add the node to one or more groups before performing subsequent containerized application operations. The process example is as follows.

1. Create a node group or use an existing node group.

    Creating a node group is to add nodes to a node group for management, enabling batch management and operations on nodes. For details on APIs for creating node groups, see [Creating a Node Group](#creating-a-node-group).

    ```text
    https://{ip}:{port}/edgemanager/v1/nodegroup
    ```

2. (Optional) Query unmanaged nodes.

    Querying unmanaged nodes is to find the node ID corresponding to the current MEF Edge device. For details on APIs for querying unmanaged nodes, see [Querying the List of Nodes Not Managed by MEF](#querying-the-list-of-nodes-not-managed-by-mef).

    ```text
    https://{ip}:{port}/edgemanager/v1/node/list/unmanaged?pageNum={pageNum}&pageSize={pageSize}&name={name}
    ```

3. Manage a node.
    - Managing a node is to add the MEF Edge node device information to the node database. MEF Center performs operations only on onboarded nodes. For details on APIs for onboarding nodes, see [Node Management](#node-management).

        ```text
        https://{ip}:{port}/edgemanager/v1/node/add
        ```

    - If the node group ID is not set when onboarding a node, add the node to a node group by following the instructions in [Adding a Node to a Node Group](#adding-a-node-to-a-node-group) before using node resources.

        ```text
        https://{ip}:{port}/edgemanager/v1/nodegroup/node
        ```

4. (Optional) Modify a node.

    The node ID serves as the unique identifier. You can modify the node by changing its name and description. For details on APIs for modifying nodes, see [Modifying Managed Nodes](#modifying-managed-nodes).

    ```text
    https://{ip}:{port}/edgemanager/v1/node
    ```

5. (Optional) Delete a managed node or remove a node from a node group.
    - Deleting a managed node:

        Batch delete nodes based on a specified array of node IDs. For details on the node deletion APIs, see [Deleting Nodes Managed by MEF](#deleting-nodes-managed-by-mef).

        ```text
        https://{ip}:{port}/edgemanager/v1/node/batch-delete
        ```

    - Removing a node from a node group:
        - By deleting nodes from a node group, you can batch delete multiple nodes and uninstall and delete containerized applications on the nodes. For details about the APIs, see [Deleting a Nodes from a Node Group](#deleting-a-node-from-a-node-group).

            ```text
            https://{ip}:{port}/edgemanager/v1/nodegroup/node/batch-delete
            ```

        - By deleting the Pod of a single containerized application, you can remove a node from a node group to uninstall the corresponding containerized application. For details about the APIs, see [Uninstalling a Containerized Application](#uninstalling-a-container-application).

            ```text
            https://{ip}:{port}/edgemanager/v1/nodegroup/pod/batch-delete
            ```

### Node Group Management<a name="ZH-CN_TOPIC_0000001526721216"></a>

#### Creating a Node Group<a id="creating-a-node-group"></a>

**Command Function<a name="section6814152820473"></a>**

After receiving the node group creation message, MEF Center validates the validity of each field in the message. Upon successful validation, it saves the node group to the database and returns the node group ID in the response.

**Syntax<a name="section1523485417488"></a>**

Operation Type: **POST**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup**

Request Body:

```json
{
    "nodeGroupName": NodeGroupName,
    "description": NodeGroupDescription
}
```

**Request Parameters<a name="section629411392578"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|nodeGroupName|Yes|Node Group Name|String, 1 to 32 characters in length. Supports uppercase and lowercase letters, digits, and underscores (_). Must start with a letter and cannot end with an underscore. The edge node group name is the unique identifier of the edge node group and cannot be duplicated.|
|description|No|Node Group Description|String, 0 to 512 characters in length, containing non-whitespace characters and spaces.|

**Example<a name="section66472019595"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup
```

Request Body:

```json
{
    "nodeGroupName": "node_group_name",
    "description": "node_group_description"
}
```

Response Example:

```json
{
   "status": "00000000",
   "msg": "success",
   "data": 1
}
```

Response Status Code: 200

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Integer|Successfully created Node Group ID|

#### Querying a Node Group List<a name="ZH-CN_TOPIC_0000001577441477"></a>

**Command Function<a name="section126491058142014"></a>**

Queries the node group list. URL parameters are used for paginated queries of data, returning the created node groups and the total number of nodes contained in these node groups.

**Syntax<a name="section8841172362113"></a>**

Operation Type: **GET**

**URL：https:**//_\{ip\}:\{port\}_/**edgemanager/v1/nodegroup/list?pageSize**=_\{pageSize\}_**&pageNum=**_\{pageNum\}_**&name=**_\{name\}_

**URL Parameters<a name="section6302105810236"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|pageSize|Yes|Page Size|An integer ranging from 1 to 100.|
|pageNum|Yes|Page Number (Ordinal)|The minimum value is 1, and the maximum value is 2^31-1.|
|name|Optional|Fuzzy Search Keyword|A string of 0 to 253 characters, which cannot contain whitespace characters.|

**Usage Example<a name="section441715250263"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/nodegroup/list?pageSize=20&pageNum=1&name=group1
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section1855315253619"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Node Group Information|

**Table 3**  Data field description

|Parameter|Type|Description|
|--|--|--|
|groups|Array|Node group details|
|createdAt|String|Node Group Creation Time|
|description|String|Node Group Description|
|groupName|String|Node Group Name|
|id|Number|Node Group ID|
|nodeCount|Number|Total number of nodes in this node group|
|updatedAt|String|Node Group Modification Time|
|total|Number|Total number of node groups matching the criteria|

#### Querying Node Group Details<a name="ZH-CN_TOPIC_0000001526881188"></a>

**Command Function<a name="section64293810487"></a>**

Queries node group details. It can return the detailed information of the specified node group and the details of all nodes belonging to that node group based on the specified Node Group ID.

**Syntax<a name="section10138183434819"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup?id=**_\{**id**\}_

**URL parameter<a name="section49163584912"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|id|Yes|Node Group ID|32-bit unsigned number. Minimum value is 1, maximum value is 2^32-1.|

**Usage Example<a name="section39743015111"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/nodegroup?id=1
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section051915331053"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|Object|Node Group Information|

**Table 3**  Data field description

| Parameter | Type | Description |
|--|--|--|
| createdAt | String | Node Group Creation Time |
| updatedAt | String | Node Group Modification Time |
| description | String | Node Group Description |
| groupName | String | Node Group Name |
| id | Number | Node Group ID |
| nodes | Array | Details of nodes in the node group |

**Table 4**  Node field description

| Parameter | Type | Description |
|--|--|--|
| createdAt | String | Node Creation Time |
| description | String | Node description |
| nodeName | String | Node name |
| id | Number | Node ID |
| ip | String | Node IP |
| isManaged | Boolean | Node Managed Status |
| nodeType | String | Node type |
| status | String | Node status<li>ready: Ready</li><li>notReady: Not ready</li><li>unknown: Unknown</li><li>offline: Offline</li><li>abnormal: Abnormal</li> |
| uniqueName | String | Node Hostname |
| updatedAt | String | Node Modification Time |
| serialNumber | String | Node Serial Number |
| softwareInfo | String | Node Software Information |

#### Modifying a Node Group<a name="ZH-CN_TOPIC_0000001577401161"></a>

**Command Function<a name="section154651437121619"></a>**

Modifies node group information. It can modify node group parameters based on the specified node group ID. Currently, the node group name and node group description fields are supported for modification.

**Syntax<a name="section1914654181610"></a>**

Operation Type: **PATCH**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup**

Request Body:

```json
{
    "groupID": NodeGroupId,
    "nodeGroupName": NodeGroupName,
    "description": NodeGroupDescription
}
```

**Request Parameters<a name="section3775151511234"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|groupID|Yes|Node Group ID|32-bit unsigned integer. Minimum value: 1, maximum value: 2^32-1.|
|nodeGroupName|Yes|Node Group Name|String, 1 to 32 characters in length. Supports uppercase and lowercase letters, digits, and underscores (_). Must start with a letter and cannot end with an underscore. The edge node group name is a unique identifier for the edge node group and cannot be duplicated.|
|description|No|Node Group Description|String, 0 to 512 characters in length, including non-whitespace characters and spaces.|

**Usage Example<a name="section115421857182411"></a>**

Request example:

```bash
PATCH https://10.10.10.10:30035/edgemanager/v1/nodegroup
```

Request Body:

```json
{
    "groupID": 1,
    "nodeGroupName": "node_group_1",
    "description": "node_group_1_description"
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section921712284514"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|

#### Deleting a Node Group<a name="ZH-CN_TOPIC_0000001527041116"></a>

**Command Function<a name="section151031891312"></a>**

Node groups can be deleted in batches based on a specified array of Node Group IDs. MEF only allows the deletion of node groups that have no deployed applications. If a node group has deployed containerized application instances, the deletion will be rejected.

**Syntax<a name="section161763119328"></a>**

Operation Type: **POST**

**URL: https:**//_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/batch-delete**

Request Body:

```json
{
    "groupIDs": [NodeGroupId]
}
```

**Request Parameters<a name="section189991912359"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|groupIDs|Yes|Array of IDs of the node groups to delete|Array of 32-bit unsigned numbers. The array can contain a maximum of 1024 elements; each number ranges from 1 to 2^32-1.|

**Example<a name="section152662567362"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/batch-delete
```

Request body:

```json
{
    "groupIDs": [1,2]
}
```

Response example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section0500202382618"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|Deleted Node Group ID.|
|failedInfos|Hash table, both key and value are strings|The key is the ID of the node group that failed to be deleted, and the value is the reason for the failure of this ID.|

#### Counting up Node Groups<a name="ZH-CN_TOPIC_0000001577280981"></a>

**Command Function<a name="section12583119104618"></a>**

Counts the number of node groups.

**Syntax<a name="section17413205114713"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/stats**

**Example<a name="section1544122611485"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/nodegroup/stats
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success",
    "data": 1
}
```

Response Status Code: 200

**Output Description<a name="section3345210496"></a>**

**Table 1** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Number|Total number of node groups|

### Node Management<a name="ZH-CN_TOPIC_0000001577601057"></a>

#### Querying the List of Nodes Not Managed by MEF<a id="querying-the-list-of-nodes-not-managed-by-mef"></a>

**Command Function<a name="section1477411417301"></a>**

Queries the list of nodes not managed by MEF. URL parameters are used for paginated data queries, returning details of unmanaged nodes.

> [!NOTE]
> Nodes with the MEF Center software package installed will not appear among the nodes not managed by MEF.

**Syntax<a name="section3953928183116"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/list/unmanaged?pageNum=**_\{pageNum\}_**&pageSize=**_\{pageSize\}_**&name=**_\{name\}_

**URL Parameters<a name="section19387122210348"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|pageSize|Yes|Page Size|An integer from 1 to 100.|
|pageNum|Yes|Page Number (Ordinal)|Minimum value is 1, maximum value is 2^31-1.|
|name|Optional|Fuzzy Search Keyword|A string of 0 to 253 characters, cannot contain whitespace characters.|

**Usage Example<a name="section18245194013513"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/list/unmanaged?pageNum=1&pageSize=20&name=node-1
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section051915331053"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|Object|Query result|

**Table 3** data field description

|Parameter|Type|Description|
|--|--|--|
|nodes|Array|Details of unmanaged nodes|
|total|Number|Total Nodes Queried|

**Table 4** nodes field description

|Parameter|Type|Description|
|--|--|--|
|createdAt|String|Node Creation Time|
|description|String|Node Description|
|id|Integer|Node ID (Identifier when used as a managed node)|
|ip|String|Node IP|
|isManaged|Boolean|Node Managed Status|
|nodeType|String|Node Type|
|nodeName|String|Node Name|
|serialNumber|String|Node Serial Number|
|softwareInfo|String|Node Software Information|
|status|String|Node Status<li>ready: Ready</li><li>notReady: Not Ready</li><li>unknown: Unknown</li><li>offline: Offline</li><li>abnormal: Abnormal</li>|
|uniqueName|String|Node Hostname|
|updatedAt|String|Node Modification Time|

#### Deleting Nodes Not Managed by MEF<a id="deleting-nodes-not-managed-by-mef"></a>

**Command Function<a name="section95491426172615"></a>**

MEF Center can add nodes within a K8s cluster to the database, making them unmanaged nodes. Unmanaged nodes can be deleted from the cluster through the delete unmanaged node API.

**Syntax<a name="section13923204216269"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/batch-delete/unmanaged**

Request Body:

```json
{
    "nodeIDs": [NodeId]
}
```

**Request Parameters<a name="section1240011511303"></a>**

**Table 1**  Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|nodeIDs|Yes|Node ID array|32-bit unsigned integer array. The array can contain a maximum of 1024 elements; each value ranges from 1 to 2^32-1.|

**Usage Example<a name="section15829115143510"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/node/batch-delete/unmanaged
```

Request Body:

```json
{
     "nodeIDs": [1,2]
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section1293674119536"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|Deleted Node ID.|
|failedInfos|Hash table, both key and value are strings|The key is the ID of the failed node, and the value is the failure reason for that ID.|

#### Managing Nodes<a id="managing-nodes"></a>

**Command Function<a name="section95491426172615"></a>**

MEF Center can add nodes from the K8s cluster to the database, making them unmanaged nodes. By querying the unmanaged node API, you can obtain the IDs of these nodes. Using the ID parameter, you can add unmanaged nodes to MEF Center, making them managed nodes.

**Syntax<a name="section13923204216269"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/add**

Request Body:

```json
{
    "name": NodeName,
    "description": NodeDescription,
    "groupIDs": [GroupId],
    "nodeID": NodeId
}
```

**Request Parameters<a name="section1240011511303"></a>**

**Table 1**  Parameter description

| Parameter | Mandatory (Yes/No)| Description | Value Requirement |
|--|--|--|--|
| name | Yes | Node Name | String, length range 1~64, supports uppercase letters, lowercase letters, digits, and other characters (-_); must not start or end with an underscore or hyphen. |
| description | No | Node Description | String, length range 0~512 characters; contains non-whitespace characters and spaces. |
| groupIDs | No | Array of Node Group IDs to join | Array of 32-bit unsigned integers. Maximum 10 elements; each value ranges from 1 to 2^32-1. |
| nodeID | Yes | Node ID | 32-bit unsigned integer. Value ranges from 1 to 2^32-1. |

**Usage Example<a name="section15829115143510"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/node/add
```

Request Body:

```json
{
     "name": "node-1",
     "description": "node_1_description",
     "groupIDs": [1,2],
     "nodeID": 1
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section1293674119536"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|IDs of the node groups that were successfully joined.|
|failedInfos|Hash table, both key and value types are string|The key is the ID of the node group that failed to join, and the value is the reason for the failure of this ID.|

#### Adding Nodes to a Node Group<a id="adding-a-node-to-a-node-group"></a>

**Command Function<a name="section19593191729"></a>**

Adds nodes to a node group. During the process of adding a node to a node group, a check is performed to determine whether the remaining resources on the current node are sufficient to run all containerized applications already deployed in the target node group. If the remaining resources on the node do not meet the requirements, adding the node to the node group will fail.

> [!NOTE]
> MEF checks three types of node resources based on the requirements of containerized applications: CPU, memory, and NPU. Users must ensure other types of container resource requirements themselves.
> MEF's resource limits only take effect for nodes whose status is "ready".

**Syntax<a name="section14423154212211"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/node**

Request Body:

```json
{
    "groupID": NodeGroupId,
    "nodeIDs": [NodeId]
}
```

**Request Parameters<a name="section151467537510"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|groupID|Yes|ID of the node group to be added|32-bit unsigned integer. Minimum value: 1, maximum value: 2^32-1.|
|nodeIDs|Yes|Array of node IDs to be added to the node group|Array of 32-bit unsigned integers. Maximum number of elements in the array: 1024; minimum value for each integer: 1, maximum value: 2^32-1.|

**Example<a name="section15101258574"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/node
```

Request body:

```json
{
    "groupID": 1,
    "nodeIDs": [1,2]
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section1293674119536"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|IDs of nodes successfully added|
|failedInfos|Hash table, both key and value types are strings|The key is the ID of the node group that failed to be added, and the value is the reason for the failure of this ID|

#### Deleting Nodes from a Node Group<a id="deleting-a-node-from-a-node-group"></a>

**Command Function<a name="section16519865104"></a>**

Deletes nodes from a node group. It supports batch deletion of multiple nodes, evicting nodes from the node group, and uninstalling and deleting containerized applications on the nodes.

**Syntax<a name="section15356163871019"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/node/batch-delete**

Request Body:

```json
{
    "groupID": NodeGroupId,
    "nodeIDs": [NodeId]
}
```

**Request Parameters<a name="section9187164141312"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|groupID|Yes|ID of the node group from which nodes are to be deleted|32-bit unsigned number. The minimum value is 1, and the maximum value is 2^32-1.|
|nodeIDs|Yes|Array of IDs of nodes to be deleted|Array of 32-bit unsigned numbers. The maximum number of elements in the array is 1024; the minimum value of each number is 1, and the maximum value is 2^32-1.|

**Example<a name="section15194112517155"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/node/batch-delete
```

Request message body:

```json
{
    "groupID": 1,
    "nodeIDs": [1,2]
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section1293674119536"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. This field is not returned if all batch operations are successful.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|Deleted Node ID|
|failedInfos|Hash table, both key and value are strings|The key is the ID of the node that failed to be deleted, and the value is the reason for the failure of this ID|

#### Querying the List of Nodes Managed by MEF<a id="ZH-CN_TOPIC_0000001577441385"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the list of nodes managed by MEF. URL parameters are used for paginated queries of data, returning details of managed nodes and the node groups to which these nodes belong.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/list/managed?pageNum=**_\{pageNum\}_**&pageSize=**_\{pageSize\}_**&name=**_\{name\}_

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|pageSize|Yes|Page Size|An integer ranging from 1 to 100.|
|pageNum|Yes|Page Number (Ordinal)|The minimum value is 1, and the maximum value is 2^31-1.|
|name|No|Fuzzy Search Keyword|A string with a length of 0 to 253 characters, and cannot contain whitespace characters.|

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/list/managed?pageNum=1&pageSize=20&name=node-1
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section051915331053"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** data field description

|Parameter|Type|Description|
|--|--|--|
|nodes|Array|Details of managed nodes|
|total|Number|Total Nodes Queried|

**Table 4** Description of nodes fields

|Parameter|Type|Description|
|--|--|--|
|createdAt|String|Node Creation Time|
|description|String|Node description|
|id|Number|Node ID|
|ip|String|Node IP|
|isManaged|Boolean|Node Managed Status|
|nodeGroup|String|Name of the node group the node joins, separated by commas|
|nodeName|String|Node name|
|nodeType|String|Node type|
|serialNumber|String|Node Serial Number|
|softwareInfo|String|Node Software Information|
|status|String|Node status<li>ready: Ready</li><li>notReady: Not ready</li><li>unknown: Unknown</li><li>offline: Offline</li><li>abnormal: Abnormal</li>|
|uniqueName|String|Node Hostname|
|updatedAt|String|Node Modification Time|

#### Querying the List of ALL Nodes<a name="ZH-CN_TOPIC_0000001577600949"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the list of all nodes. URL parameters are used for paginated queries, returning node details in the database and the node groups these nodes join.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/list?pageNum=**_\{pageNum\}_**&pageSize=**_\{pageSize\}_**&name=**_\{name\}_

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|pageSize|Yes|Page Size|An integer ranging from 1 to 100.|
|pageNum|Yes|Page Number (Ordinal)|The minimum value is 1, and the maximum value is 2^31-1.|
|name|Optional|Fuzzy Search Keyword|A string with a length of 0 to 253 characters, which cannot contain whitespace characters.|

**Usage Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/list?pageNum=1&pageSize=20&name=node-1
```

Response example:

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

Response Status Code: 200

**Output Description<a name="section051915331053"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** data field description

|Parameter|Type|Description|
|--|--|--|
|nodes|Array|Query result node details|
|total|Number|Total Nodes Queried|

**Table 4** Description of nodes fields

|Parameter|Type|Description|
|--|--|--|
|createdAt|String|Node Creation Time|
|description|String|Node description|
|id|Number|Node ID|
|ip|String|Node IP|
|isManaged|Boolean|Node Managed Status|
|nodeGroup|String|Name of the node group the node joins, separated by commas|
|nodeName|String|Node name|
|nodeType|String|Node type|
|serialNumber|String|Node Serial Number|
|softwareInfo|String|Node Software Information|
|status|String|Node status<li>ready: Ready</li><li>notReady: Not Ready</li><li>unknown: Unknown</li><li>offline: Offline</li><li>abnormal: Abnormal</li>|
|uniqueName|String|Node Hostname|
|updatedAt|String|Node Modification Time|

#### Querying Node Details<a id="querying-node-details"></a>

**Command Function<a name="section36445448496"></a>**

Queries detailed node information. In addition to the node information in the database, this API also includes the node's CPU, NPU, and memory resource data.

**Syntax<a name="section832411165018"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/node?id=**_\{**id**\}_ or **https://**_\{ip\}:\{port\}_**/edgemanager/v1/node?sn=**_\{**sn**\}_

**URL Parameters<a name="section9868223165119"></a>**

**Table 1** URL Parameters

<a name="table737675217261"></a>
<table><thead align="left"><tr id="row153763524260"><th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.5.1.1"><p id="p837615213264"><a name="p837615213264"></a><a name="p837615213264"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="20.02%" id="mcps1.2.5.1.2"><p id="p15376205202612"><a name="p15376205202612"></a><a name="p15376205202612"></a>Mandatory (Yes/No)</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p193761052142613"><a name="p193761052142613"></a><a name="p193761052142613"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p3376195215262"><a name="p3376195215262"></a><a name="p3376195215262"></a>Value Requirement</p>
</th>
</tr>
</thead>
<tbody><tr id="row18376195213268"><td class="cellrowborder" valign="top" width="19.98%" headers="mcps1.2.5.1.1 "><p id="p1837619526269"><a name="p1837619526269"></a><a name="p1837619526269"></a>id</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20.02%" headers="mcps1.2.5.1.2 "><p id="p137614528264"><a name="p137614528264"></a><a name="p137614528264"></a>Yes. Either <code>id</code> or <code>sn</code> must be selected.</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1737615529262"><a name="p1737615529262"></a><a name="p1737615529262"></a>Node ID</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p151951628144614"><a name="p151951628144614"></a><a name="p151951628144614"></a>32-bit unsigned integer. Minimum value is 1, maximum value is 2^32-1.</p>
</td>
</tr>
<tr id="row167513014814"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p376163020482"><a name="p376163020482"></a><a name="p376163020482"></a>sn</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p16761930154818"><a name="p16761930154818"></a><a name="p16761930154818"></a>Node Serial Number</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1776530164819"><a name="p1776530164819"></a><a name="p1776530164819"></a>Supports lowercase letters, uppercase letters, digits, underscores, and hyphens. Cannot start or end with an underscore or hyphen. Maximum length is 64 bytes.</p>
</td>
</tr>
</tbody>
</table>

**Example<a name="section68445365579"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node?id=1
```

Response example:

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

Response Status Code: 200

**Output Description<a name="section19516562114"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** Data Field Description

|Field|Type|Description|
|--|--|--|
|cpu|Number|Number of CPUs|
|createdAt|String|Node Creation Time|
|description|String|Node description|
|id|Number|Node ID|
|ip|String|Node IP|
|isManaged|Boolean|Node Managed Status|
|memory|Number|Node memory, in bytes|
|nodeGroup|String|Name of the node group the node belongs to, separated by commas|
|nodeName|String|Node name|
|nodeType|String|Node type|
|npu|Number|Number of node NPUs|
|serialNumber|String|Node Serial Number|
|softwareInfo|String|Node Software Information|
|status|String|Node status<li>ready: Ready</li><li>notReady: Not Ready</li><li>unknown: Unknown</li><li>abnormal: Abnormal</li><li>offline: Offline</li>|
|uniqueName|String|Node Hostname|
|updatedAt|String|Node Modification Time|

#### Modifying Managed Nodes<a id="modifying-managed-nodes"></a>

**Command Function<a name="section8416161212520"></a>**

Modifies a node. Modify node parameters based on the specified node ID. Currently, only the node name and node description fields are supported.

**Syntax<a name="section177338310511"></a>**

Operation Type: **PATCH**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/node**

Request Body:

```json
{
    "nodeID": NodeId,
    "nodeName": NodeName,
    "description": NodeDescription
}
```

**Request Parameters<a name="section136292931320"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|nodeID|Yes|Node ID|32-bit unsigned integer. Minimum value is 1, maximum value is 2^32-1.|
|nodeName|Yes|Node Name|String, 1 to 64 characters in length. Supports uppercase letters, lowercase letters, digits, and other characters (_-); must not start or end with an underscore or hyphen.|
|description|No|Node Description|String, 0 to 512 characters in length; contains non-whitespace characters and spaces.|

**Usage Example<a name="section2392154315204"></a>**

Request Example:

```bash
PATCH https://10.10.10.10:30035/edgemanager/v1/node
```

Request Body:

```json
{
    "nodeID": 1,
    "nodeName": "node-1",
    "description": "node_1_description"
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section10297132653614"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|

#### Deleting Nodes Managed by MEF<a id="deleting-nodes-managed-by-mef"></a>

**Command Function<a name="section1893641916283"></a>**

Nodes that have been managed by MEF can be deleted in batches based on a specified array of node IDs. The nodes will be removed from the K8s cluster. If you wish to manage the nodes again, you need to re-establish network management integration in MEF Edge. MEF only allows deletion of nodes that have no deployed applications. Before deleting the specified nodes, you must uninstall all applications deployed on those nodes. For details, see [Uninstalling Containerized Applications](#uninstalling-containerized-applications).

**Syntax<a name="section26338342289"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/batch-delete**

Request Body:

```json
{
  "nodeIDs": [nodeId]
}
```

**Request Parameters<a name="section2073163333113"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|nodeIDs|Yes|Array of node IDs to delete|Array of 32-bit unsigned numbers. The maximum number of elements in the array is 1024; each number ranges from 1 to 2^32-1.|

**Usage Example<a name="section370651413331"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/node/batch-delete
```

Request Body:

```json
{
    "nodeIDs": [1,2]
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section1293674119536"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3**  Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|Deleted Node ID|
|failedInfos|Hash table, both key and value types are String|The key is the ID of the node that failed to be deleted, and the value is the reason for the failure of this ID|

#### Collecting Node Status Information<a name="ZH-CN_TOPIC_0000001527041068"></a>

**Command Function<a name="section15340202213373"></a>**

Queries the status of all nodes and returns the number of nodes in each status.

**Syntax<a name="section2889313103819"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/node/stats**

**Usage Guide<a name="section19319163010396"></a>**

None.

**Usage Example<a name="section17966161213413"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/node/stats
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section1293674119536"></a>**

**Table 1**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Query result|

**Table 2**  data Field Description

|Parameter|Type|Description|
|--|--|--|
|ready|Number|Number of nodes in ready state|
|notReady|Number|Number of nodes in not ready state|
|unknown|Number|Number of nodes in unknown state|
|offline|Number|Number of nodes in offline state|
|abnormal|Number|Number of nodes in abnormal state|

#### Uninstalling Containerized Applications<a id="uninstalling-containerized-applications"></a>

**Command Function<a name="section141453116219"></a>**

Uninstalls a containerized application by removing a node from the corresponding node group (i.e., deleting the Pod of a single application).

**Syntax<a name="section146212029938"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/nodegroup/pod/batch-delete**

Request Body:

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

> [!NOTE]
> Multiple containerized applications can be uninstalled at once in a list format. The list size ranges from 1 to 1024. For each entry in the list, at least one of the Node ID and Node Group ID must differ from other entries.

**Request Parameters<a name="section435110501651"></a>**

**Table 1** Parameter description

| Parameter | Mandatory (Yes/No) | Description | Value Requirement |
|--|--|--|--|
| nodeID | Yes | Node ID | 32-bit unsigned integer. Minimum value is 1, maximum value is 2^32-1. |
| groupID | Yes | Node Group ID | 32-bit unsigned integer. Minimum value is 1, maximum value is 2^32-1. |

**Example<a name="section17189154898"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/nodegroup/pod/batch-delete
```

Request Body:

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

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section1293674119536"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Operation result|

**Table 3**  Data Field Description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|ID pairs of nodes and node groups successfully uninstalled|
|failedInfos|Hash table, both key and value types are strings|The key is the ID pair of the node and node group that failed to uninstall, and the value is the reason for the failure of this ID pair|

## Containerized Application Management APIs<a id="ZH-CN_TOPIC_0000001577401181"></a>

### Overview<a name="ZH-CN_TOPIC_0000001577441409"></a>

As a fundamental feature of MEF, containerized application management handles the full lifecycle management of user applications. User applications are published as container images. MEF manages user containerized application images, covering the creation, deletion, modification, and querying of containerized applications, deployment of containerized applications to node groups, uninstallation of containers from node groups, and uninstallation of containers from individual nodes. Users implement the corresponding functions by calling the relevant APIs. Containerized applications are operated on a per-node-group basis: when a node joins a node group, containers are automatically deployed to that node; when a node leaves a node group, containers on that node are automatically uninstalled.

> [!NOTE]
>
> MEF can use containerized application images in three ways: through the Docker public image repository, third-party image repositories, or by manually importing images to MEF Edge. When using an image repository, users must ensure the network connection between the MEF Edge device and the image repository is functional, and that the image repository itself is working properly. If users need to obtain images from a third-party image repository, for the usage process, see [Overview](#configuration-api-introduction).

**Constraints<a name="section1123013112347"></a>**

- MEF allows a maximum of 1000 concurrent containerized applications. If a user deploys containerized applications not managed by MEF Center to device nodes, it may cause containerized applications to fail to deploy due to insufficient resources.
- An MEF Edge node supports deploying a maximum of 20 containerized applications. Excessive containerized applications may cause device performance degradation.
- When users concurrently call APIs related to deploying containerized applications ([Deploying Containerized Applications](#deploying-container-applications), [Node Management](#node-management), [Adding a Node to a Node Group](#adding-a-node-to-a-node-group)), they may fail to run properly due to insufficient node resources.
- MEF Edge reserves 1024 MB of memory and 1 CPU core.
  - The total available memory resources for all containerized applications can be calculated using the formula: Total Available Memory Resources = Total System Memory Resources - System Reserved Memory Resources.
  - The total available CPU resources for all containerized applications can be calculated using the formula: Total Available CPU Resources = Total System CPU Resources - System Reserved CPU Resources.

- MEF Center allows a maximum of 20 containerized applications to be deployed on a **single node group** and 20 containerized applications on a **single node**. When the number of containerized applications deployed on a node group or node exceeds the upper limit, the application deployment function for the corresponding node group and the function for adding new nodes to the node group will be restricted.

**Containerized Application Management Process Overview<a name="section132131728522"></a>**

When using containerized application management functions through API calls, the creation and deployment of containerized applications can be performed separately. Users can first create the required containerized applications and later decide which node groups to deploy them to. An example of the containerized application management process is as follows.

1. Create a Containerized Application

    Users can configure containerized application parameters through the Create Containerized Application API. Upon a successful call, the AppID of the created containerized application is returned. For details on the Create Containerized Application API, see [Creating a Containerized Application](#creating-a-container-application).

    ```text
    https://{ip}:{port}/edgemanager/v1/app
    ```

2. (Optional) Query the Containerized Application List

    Querying the containerized application list is to obtain the AppID of the containerized application to be deployed. For details on the Query Containerized Application List API, see [Querying the Containerized Application List](#querying-the-container-application-list).

    ```text
    https://{ip}:{port}/edgemanager/v1/app/list?pageNum={value1}&pageSize={value2}&name={value3}
    ```

3. Deploy a Containerized Application

    For the API to deploy a containerized application, see [Deploying Containerized Applications](#deploying-container-applications).

    ```text
    https://{ip}:{port}/edgemanager/v1/app/deployment
    ```

4. (Optional) Query Deployed Containerized Applications

    Users can query the running status of a containerized application with a specified AppID through the query deployed containerized applications API. For details on querying deployed containerized applications, see [Querying the List of Deployed Containerized Applications](#querying-the-list-of-deployed-container-applications).

    ```text
    https://{ip}:{port}/edgemanager/v1/app/deployment?appID={id}
    ```

5. (Optional) Update a Containerized Application

    If the corresponding containerized application has been deployed, the deployed containerized application will also be updated. Currently, only modifications to the container image name and container image version are supported. For the API to update a containerized application, see [Updating a Containerized Application](#updating-a-container-application).

    ```text
    https://{ip}:{port}/edgemanager/v1/app
    ```

6. (Optional) Uninstall a containerized application

    For the containerized application uninstallation API, see [Uninstalling a Containerized Application](#uninstalling-a-container-application).

    ```text
    https://{ip}:{port}/edgemanager/v1/app/deployment/batch-delete
    ```

7. (Optional) Delete a containerized application

    When deleting a containerized application, only applications that have not been deployed can be deleted. If the corresponding containerized application has been deployed, you must uninstall it first. For the containerized application deletion API, see [Deleting a Containerized Application](#deleting-a-container-application).

    ```text
    https://{ip}:{port}/edgemanager/v1/app/batch-delete
    ```

### Creating a Containerized Application<a id="creating-a-container-application"></a>

**Command Function<a name="section135251624204320"></a>**

Creates a containerized application. The message parameters are the configuration information of the containerized application. The created containerized application is saved in the MEF Center database. Only created containerized applications can be deployed, uninstalled, and undergo other operations. After MEF Center receives the containerized application creation message, it validates the validity of each field in the message. After the validation passes, it saves the containerized application parameters to the database and returns the unique Containerized Application ID in the message.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/app**

Request Header:

```http
Content-Type: application/json
```

Request Body:

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

**Request Parameters<a id="request-parameters"></a>**

**Table 1**  Parameters for creating a containerized application

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|appName|Yes|Containerized Application Name|String, supporting 1 to 32 characters, lowercase letters, digits, and hyphens (-). Must start and end with a letter or digit.|
|description|No|Containerized Application Description|String, supporting 0 to 512 characters; whitespace characters other than spaces are not supported.|
|containers|Yes|Container Configuration Array|Array of Container objects. A single containerized application supports 1 to 10 container configurations.|

**Table 2** Container parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|name|Yes|Container Name.|String, 1 to 32 characters in length; supports lowercase letters, digits, and hyphens (-). Must start and end with a letter or digit. Containerized application names must be unique.|
|cpuRequest|Yes|Number of CPU cores requested by the container.|Numeric, ranging from 0.01 to 1000, accurate to two decimal places.|
|cpuLimit|Optional|Maximum number of CPU cores the container can use.|Numeric, ranging from 0.01 to 1000, accurate to two decimal places, and must be greater than or equal to cpuRequest.|
|memRequest|Yes|Container Memory Request.|Numeric, ranging from 4 to 1024000, integer only, in MiB.|
|memLimit|Optional|Maximum memory size the container can use.|Numeric, ranging from 4 to 1024000, integer only, in MiB, and must be greater than or equal to memRequest.|
|npu|Optional|Number of NPU cores requested by the container.|Numeric, ranging from 0 to 32, integer only.|
|image|Yes|Image Name used. When using a third-party image registry, the full name must include the image registry server IP or domain name, port, project, and image name. For example, fd.fusiondirector.huawei\.com:443/library/ubuntu. If the user does not specify the image hostname and port, the containerized application will use the Docker public registry.|String, 1 to 256 characters in length; supports lowercase letters, uppercase letters, digits, and other characters (:-._/).|
|imageVersion|Yes|Image Version.|String, 1 to 32 characters in length; supports lowercase letters, uppercase letters, digits, and other characters (-._).|
|env|No|Environment variables configured inside the container.|Array of EnvVar objects, supporting a maximum of 256 key-value pairs.|
|userID|No|Specifies the user ID for container execution.<br>If this parameter is not configured, the container will run as the user specified during image creation. If the user during image creation is not a numeric ID or the numeric ID is 0, the containerized application will fail to run after deployment. When running an inference container, driver devices are required, so a user ID cannot be configured.|Numeric, ranging from 1 to 65535. Cannot be set to 0, meaning the container does not support running as the root user. When deploying an inference container, specify the user ID of HwHiAiUser (usually 1000).|
|groupID|No|Specifies the group ID for container execution.<br>If this parameter is not configured, the container will run as the user group specified during image creation. If the user group during image creation is not a numeric group ID or the numeric group ID is 0, the containerized application will fail to run after deployment. When running an inference container, driver devices are required, so a group ID cannot be configured.|Numeric, ranging from 1 to 65535. Cannot be set to 0, meaning the container does not support running as the root group. When deploying an inference container, specify the group ID of HwHiAiUser (usually 1000).|
|command|No|Container Command executed at startup.|Array of strings. Supports a maximum of 16 commands, each 1 to 256 characters in length. Supports lowercase letters, uppercase letters, digits, spaces, and other characters (-/._). Must end with a letter or digit.|
|args|No|Container Args executed at startup.|Array of strings. Supports a maximum of 16 arguments, each 1 to 256 characters in length. Supports lowercase letters, uppercase letters, digits, spaces, and other characters (-/._=). Must end with a letter or digit.|
|containerPort|No|Host port to Container Port mapping configuration.|Array of ContainerPort objects, supporting a maximum of 16 port mappings.|
|hostPathVolumes|No|Host Path Mount Information for the container.<br>When creating an inference containerized application, you must configure the mount path; otherwise, the application may fail to run.|Array of HostPathVolume objects, supporting a maximum of 256 entries.|

**Table 3** EnvVar parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|name|Yes|Environment Variable Name|String, 2 to 32 characters in length. Supports uppercase and lowercase letters, digits, and other characters (-._). Must start with an uppercase or lowercase letter and end with an uppercase or lowercase letter or digit.|
|value|Yes|Environment Variable Value|String, 1 to 512 characters in length. Supports uppercase and lowercase letters, digits, other characters (-._/:), and spaces.|

**Table 4** ContainerPort parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|name|Yes|Port Mapping Name|String, 1 to 32 characters in length. Supports lowercase letters, digits, and hyphens (-). Must start and end with a lowercase letter or digit.|
|proto|Yes|Network Transport Layer Protocol Specified for Port Mapping|String, value is TCP or UDP.|
|containerPort|Yes|Container Port|Integer, value range: 1–65535.|
|hostIP|Yes|Host IP Address Bound to Port Mapping|String, must be a valid host IP address. Only IPv4 is supported. Cannot be configured as all zeros or all 255s.|
|hostPort|Yes|Host Port Address|Integer, value range: 1024–65535.|

**Table 5** HostPathVolume parameters

| Parameter | Mandatory (Yes/No) | Description | Value Requirement |
|--|--|--|--|
| name | Yes | Mount volume name | String, 1 to 32 characters in length; supports lowercase letters, digits, and hyphens (-). Must start and end with a letter or digit. The mount volume name must be unique within the same container. |
| hostPath | Yes | Host path used by the container mount volume. <div class="note"><span class="notetitle">[!NOTE] NOTE</span><div class="notebody">Only host paths for the files or directories listed in the right column are supported for mount configuration. If you create a container image by referring to [Creating an Inference Image](./common_operations.md#creating-an-inference-image) or the [Atlas 200I A2 Accelerator Module Ascend Software Quick Installation Guide](https://support.huawei.com/enterprise/zh/doc/EDOC1100423566/4a72915b), for the corresponding default in-image mount paths, see Creating a Container Image -> Starting the Container step.</div></div> | Only host paths for the following files or directories are supported for mount configuration. <li>"/etc/sys_version.conf"</li><li>"/etc/hdcBasic.cfg"</li><li>"/usr/lib64/libaicpu_processer.so"</li><li>"/usr/lib64/libaicpu_prof.so"</li><li>"/usr/lib64/libaicpu_sharder.so"</li><li>"/usr/lib64/libadump.so"</li><li>"/usr/lib64/libtsd_eventclient.so"</li><li>"/usr/lib64/libaicpu_scheduler.so"</li><li>libcrypto.so.1.1</li><ul><li>On Ubuntu host OS: "/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1"</li><li>On openEuler host OS: "/usr/lib64/libcrypto.so.1.1.1m"</li></ul><li>/usr/lib64/libcrypto.so.3<ul><li>On Ubuntu host OS: "/usr/lib/aarch64-linux-gnu/libcrypto.so.3.0.12"</li><li>On openEuler host OS: "/usr/lib64/libcrypto.so.3.0.12"</li></ul></li><li>libyaml-0.so.2<ul><li>On Ubuntu host OS: "/usr/lib/aarch64-linux-gnu/libyaml-0.so.2.0.6"</li><li>On openEuler host OS: "/usr/lib64/libyaml-0.so.2.0.9"</li></ul></li><li>"/usr/lib64/libdcmi.so"</li><li>"/usr/lib64/libmpi_dvpp_adapter.so"</li><li>"/usr/lib64/libunified_timer.so"</li><li>"/usr/lib64/libmmpa.so"</li><li>"/usr/lib64/aicpu_kernels/"</li><li>"/usr/local/sbin/npu-smi"</li><li>"/usr/lib64/libstackcore.so"</li><li>"/usr/local/Ascend/driver/lib64"</li><li>"/var/slogd"</li><li>"/var/dmp_daemon"</li>|
| mountPath | Yes | Container mount path | A path string starting with "/", followed by uppercase and lowercase letters, digits, and other characters (_./-). Cannot contain "..". Total path length is 2 to 512 characters. The container mount path name must be unique within the same container. |

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app
```

Request Body:

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

Response Example:

```json
{
    "status":"00000000",
    "msg":"success",
    "data":3
}
```

Response Status Code: 200

**Output Description<a name="section102201014616"></a>**

**Table 6** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Number|Created Containerized Application ID|

### Querying the Containerized Application List<a id="querying-the-container-application-list"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the containerized application list. URL parameters are used for paginated data queries. Returns the configuration details of created containerized applications and the deployment node group information for these containerized applications.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/list?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&name=**_\{value3\}_

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|pageSize|Yes|Page Size|An integer from 1 to 100.|
|pageNum|Yes|Page Number (Ordinal)|Minimum value is 1, maximum value is 2^31-1.|
|name|No|Fuzzy Search Keyword. Results will only return containerized applications whose Containerized Application Name contains this field.|A string of 0 to 253 characters, must not contain whitespace characters.|

**Usage Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/list?pageNum=1&pageSize=10&name=123
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

For more field descriptions of containerized applications, see [Request Parameters](#request-parameters) in section "Creating a Containerized Application".

**Table 2**  Operation output description

|Field|Type|Description|
|--|--|--|
|status|String|Status code|
|msg|String|Description|
|data|Object|Query result|

**Table 3**  Data field description

|Field|Type|Description|
|--|--|--|
|appInfo|Array of AppInfo objects|Array of containerized application information|
|total|Number|Total number of results filtered by the fuzzy search field|
|deployed|Number|Number of deployed containerized applications|
|unDeployed|Number|Number of undeployed containerized applications|

**Table 4**  Description of the AppInfo field

|Field|Type|Description|
|--|--|--|
|appID|Number|Containerized Application ID|
|appName|String|Containerized Application Name|
|description|String|Containerized Application Description|
|createdAt|String|Creation time|
|modifiedAt|String|Update time|
|nodeGroupInfos|Array of nodeGroupInfo objects|Node Group Information|
|containers|Array of Container objects|Container Configuration Array|

**Table 5**  nodeGroupInfo field description

|Field|Type|Description|
|--|--|--|
|nodeGroupID|Number|Node Group ID|
|nodeGroupName|String|Node Group Name|

**Table 6**  Container field description

|Parameter|Type|Description|
|--|--|--|
|name|String|Container Name.|
|image|String|Image Name|
|imageVersion|String|Image Version.|
|cpuRequest|Number|CPU Request|
|cpuLimit|Number|CPU Limit|
|memRequest|Number|Memory Request|
|memLimit|Number|Memory Limit|
|npu|Number|NPU Count|
|command|String array|Container Command|
|args|String array|Container Args|
|env|EnvVar object array|Environment variables|
|containerPort|ContainerPort object array|Container Port|
|userID|Number|Container User ID|
|groupID|Number|Container Group ID|
|hostPathVolumes|HostPathVolume object array|Host Path Mount Information|

### Querying Containerized Application Details<a id="querying-container-application-details"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the details of a containerized application. Based on the specified containerized application ID, it returns the configuration details and the deployed node group information of the containerized application.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app?appID=**_\{id\}_

**Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|appID|Mandatory|Containerized Application ID|An integer with a minimum value of 1 and a maximum value of 2^32-1.|

**Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app?appID=2
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

For more field descriptions of containerized applications, see [Request Parameters](#request-parameters) in the Create Containerized Application chapter.

**Table 2** Operation output description

|Field|Type|Description|
|--|--|--|
|status|String|Status code|
|msg|String|Description information|
|data|Object|Containerized application information|

**Table 3**  Data field description

|Field|Type|Description|
|--|--|--|
|appID|Number|Containerized Application ID|
|appName|String|Containerized Application Name|
|containers|Array of Container objects|Container Configuration Array|
|createdAt|String|Creation time|
|description|String|Containerized Application Description|
|modifiedAt|String|Update time|
|nodeGroupInfos|Array of nodeGroupInfo objects|Node Group Information|

**Table 4**  Description of the Container field

|Parameter|Type|Description|
|--|--|--|
|name|String|Container Name.|
|image|String|Image Name|
|imageVersion|String|Image Version.|
|cpuRequest|Number|CPU Request|
|cpuLimit|Number|CPU Limit|
|memRequest|Number|Memory Request|
|memLimit|Number|Memory Limit|
|npu|Number|NPU Count|
|command|Array of strings|Container Command|
|args|Array of strings|Container Args|
|env|Array of EnvVar objects|Environment variables|
|containerPort|Array of ContainerPort objects|Container Port|
|userID|Number|Container User ID|
|groupID|Number|Container Group ID|
|hostPathVolumes|Array of HostPathVolume objects|Host Path Mount Information|

### Deploying Containerized Applications<a id="deploying-container-applications"></a>

**Command Function<a name="section135251624204320"></a>**

Deploys a containerized application. This is a batch API that deploys the containerized application with the specified ID to one or more node groups in the specified ID array. Deployment is subject to the remaining available resources on edge nodes. If the resources of online nodes (i.e., nodes with a status of "ready") in a node group do not meet the requirements of the containerized application to be deployed, the deployment of the containerized application to that node group will fail.

> [!NOTE]
>
>- When a containerized application is successfully deployed by MEF Center, the daemonset resource for the containerized application in K8s will be created successfully. The actual running status of the MEF Edge containerized application must be confirmed by querying the application instance.
>- MEF checks three types of node resources based on the containerized application requirements: CPU, memory, and NPU. Users are responsible for ensuring other types of container resource requirements.
>- MEF's resource limits only apply to nodes with a status of "ready".

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment**

Request Header:

```http
Content-Type: application/json
```

Request Body:

```json
{
    "appID": AppId,
    "nodeGroupIds": [NodeGroupId]
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|appID|Yes|Application ID|Integer, minimum value 1, maximum value 2^32-1. Must be an existing application ID.|
|nodeGroupIds|Yes|Node Group ID List|Array, must contain unique node group IDs, with an array length of [1, 1024].|

**Usage Example<a name="section9299576112"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app/deployment
```

Request message body:

```json
{
    "appID": 1,
    "nodeGroupIds": [
        1,2
    ]
}
```

Response Example:

```json
{
    "status":"00000000",
    "msg":"success"
}
```

Response Status Code: 200

**Output Description<a name="section18838112010719"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|IDs of node groups successfully deployed|
|failedInfos|Hash table, both key and value types are strings|The key is the ID of the node group that failed to deploy, and the value is the reason for the failure of this ID|

### Querying the List of Deployed Containerized Applications<a id="querying-the-list-of-deployed-container-applications"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the list of deployed containerized applications. Based on the specified containerized application ID, it returns the list of deployed instances of that containerized application, including the node and node group information, running status, and container status of these instances.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment**?**appID**=_\{id\}_

**URL Parameters<a name="section1774293413516"></a>**

**Table 1**  URL Parameters

| Parameter | Mandatory (Yes/No) | Description | Value Requirement |
|--|--|--|--|
| appID | Yes | Containerized Application ID | An integer with a minimum value of 1 and a maximum value of 2^32-1. |

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/deployment?appID=1
```

Response Example:

```json
{
    "status":"00000000",
    "msg":"success",
    "data":{
        "appInstances": [
        {
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

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

For more field descriptions of containerized applications, see [Request Parameters](#request-parameters) in the Create Containerized Application chapter.

**Table 2** Operation output description

|Field|Type|Description|
|--|--|--|
|status|String|Status code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** Data field description

|Field|Type|Description|
|--|--|--|
|appInstances|Array of AppInstance objects|Deployed Containerized Application List|
|total|Number|Total Results|

**Table 4**  AppInstance field description

|Field|Type|Description|
|--|--|--|
|appID|Number|Containerized Application ID|
|appName|String|Containerized Application Name|
|appStatus|String|Pod application running status<li>pending: Pending</li><li>running: Running</li><li>succeeded: Succeeded</li><li>failed: Failed</li><li>unknown: Unknown</li>|
|nodeGroupInfo|nodeGroupInfo object|Deployment Node Group Information|
|nodeID|Number|Deployment Node ID|
|nodeName|String|Deployment Node Name|
|nodeStatus|String|Node status<li>ready: Ready</li><li>notReady: NotReady</li><li>offline: Offline</li><li>unknown: Unknown</li>|
|createdAt|String|Creation time|
|containerInfo|Array of ContainerInfo objects|Containerized Application Pod Information|

**Table 5**  ContainerInfo field description

|Field|Type|Description|
|--|--|--|
|name|String|Container name under the containerized application Pod|
|image|String|Image name of the Container|
|status|String|Running status of the Container under the Pod.<li>waiting: Waiting</li><li>running: Running</li><li>terminated: Terminated</li><li>unknown: Unknown</li>|
|restartCount|Number|Restart count of the corresponding container in the deployed containerized application|

**Table 6**  nodeGroupInfo field description

|Field|Type|Description|
|--|--|--|
|nodeGroupID|Number|Node Group ID|
|nodeGroupName|String|Node Group Name|

### Querying the List of Containerized Applications Deployed on a Node<a id="ZH-CN_TOPIC_0000001577441457"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the list of containerized applications deployed on a specified node by node ID. It returns a list of all containerized application instances running on that node, including the node and node group information, running status, and container status of these instances.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/node?**nodeID**=**_\{id\}_

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|nodeID|Yes|Node ID|32-bit unsigned integer. Minimum value: 1, maximum value: 2^32-1.|

**Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/node?nodeID=2
```

Response Example:

```json
{
    "status":"00000000",
    "msg":"success",
    "data":{
        "appInstances": [
        {
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

Status Code: 200

**Output Description<a name="section127921251728"></a>**

For more field descriptions of containerized applications, see [Request Parameters](#request-parameters) in the Create Containerized Application chapter.

**Table 2** Operation output description

|Field|Type|Description|
|--|--|--|
|status|String|Status code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** Data field description

|Field|Type|Description|
|--|--|--|
|appInstances|Array of AppInstance objects|Deployed Containerized Application List|
|total|Number|Total Results|

**Table 4**  AppInstance field description

|Field|Type|Description|
|--|--|--|
|appID|Number|Containerized Application ID|
|appName|String|Containerized Application Name|
|appStatus|String|Pod application running status<li>pending: Pending</li><li>running: Running</li><li>succeeded: Succeeded</li><li>failed: Failed</li><li>unknown: Unknown</li>|
|nodeGroupInfo|Object|Deployment Node Group Information|
|nodeID|Number|Deployment Node ID|
|nodeName|String|Deployment Node Name|
|nodeStatus|String|Node status<li>ready: Ready</li><li>notReady: Not Ready</li><li>offline: Offline</li><li>unknown: Unknown</li>|
|createdAt|String|Creation time|
|containerInfo|Array of ContainerInfo objects|Containerized Application Pod Information|

**Table 5**  ContainerInfo field description

|Field|Type|Description|
|--|--|--|
|name|String|Container name under the containerized application Pod|
|image|String|Image name of the Container|
|status|String|Running status of the Container under the Pod.<li>waiting: Waiting</li><li>running: Running</li><li>terminated: Terminated</li><li>unknown: Unknown</li>|
|restartCount|Number|Restart count of the corresponding container in the deployed containerized application|

**Table 6** nodeGroupInfo field description

|Field|Type|Description|
|--|--|--|
|nodeGroupID|Number|Node Group ID|
|nodeGroupName|String|Node Group Name|

### Querying the List of Deployed Containerized Applications in Batches<a name="ZH-CN_TOPIC_0000001577401093"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the list of deployed containerized applications. Based on the specified pagination query request parameters, it returns a filtered list of containerized application instances, including the node and node group information, running status, and container status of these instances.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment/list?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&name=**_\{value3\}_

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** URL Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|pageSize|Yes|Page Size|An integer ranging from 1 to 100.|
|pageNum|Yes|Page Number (Ordinal)|The minimum value is 1, and the maximum value is 2^31-1.|
|name|No|Fuzzy Search Keyword. Results will only return containerized applications whose Containerized Application Name contains this field.|A string with a length of 0 to 253 characters, and cannot contain whitespace characters.|

**Usage Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/app/deployment/list?pageNum=1&pageSize=10&name=
```

Response example:

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

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

For more field descriptions of containerized applications, see [Request Parameters](#request-parameters) in the Create Containerized Application chapter.

**Table 2** Operation output description

|Field|Type|Description|
|--|--|--|
|status|String|Status code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** Data field description

|Field|Type|Description|
|--|--|--|
|appInstances|Array of AppInstance objects|Deployed Containerized Application List|
|total|Number|Total Results|

**Table 4**  AppInstance field description

|Field|Type|Description|
|--|--|--|
|appID|Number|Containerized Application ID|
|appName|String|Containerized Application Name|
|appStatus|String|Pod application running status, including the following types: <li>pending: Pending</li><li>running: Running</li><li>succeeded: Succeeded</li><li>failed: Failed</li><li>unknown: Unknown</li>|
|nodeGroupInfo|Object|Deployment Node Group Information|
|nodeID|Number|Deployment Node ID|
|nodeName|String|Deployment Node Name|
|nodeStatus|String|Node status, including the following types: <li>ready: Ready</li><li>notReady: Not Ready</li><li>offline: Offline</li><li>unknown: Unknown</li>|
|createdAt|String|Creation time|
|containerInfo|Array of ContainerInfo objects|Containerized Application Pod Information|

**Table 5**  ContainerInfo field description

|Field|Type|Description|
|--|--|--|
|name|String|Container name under the containerized application Pod|
|image|String|Image name of the Container|
|status|String|Container running status under the Pod. <li>waiting: Waiting</li><li>running: Running</li><li>terminated: Terminated</li><li>unknown: Unknown</li>|
|restartCount|Number|Number of restarts for the corresponding container of the deployed containerized application|

**Table 6**  nodeGroupInfo field description

|Field|Type|Description|
|--|--|--|
|nodeGroupID|Number|Node Group ID|
|nodeGroupName|String|Node group name|

### Updating a Containerized Application<a id="updating-a-container-application"></a>

**Command Function<a name="section135251624204320"></a>**

Updates a containerized application. It can update the containerized application information saved in MEF Center based on the specified containerized application ID. If the corresponding containerized application has been deployed, the deployed containerized application will also be updated. Currently, only modifications to the container image name and container image version are supported; changes to other fields will not be used.

> [!NOTE]
> When MEF Center successfully updates a deployed containerized application, it means the daemonset resource in K8s has been updated successfully. The actual update of the running container on MEF Edge needs to be confirmed by querying the application instance.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **PATCH**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/app**

Request Header:

```http
Content-Type: application/json
```

Request body:

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

**Request Parameters<a name="section1774293413516"></a>**

For more field descriptions of containerized applications, see [Request Parameters](#request-parameters) in the Creating a Containerized Application section.

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|appID|Yes|Containerized Application ID|Integer ranging from 1 to 2^32-1. Must be an existing application ID.|
|appName|Yes|Containerized Application Name|String of 1 to 32 characters, consisting of lowercase letters, digits, and hyphens (-). Must start and end with a letter or digit.|
|description|No|Containerized Application Description|String of 0 to 512 characters. Whitespace characters other than spaces are not supported.|
|containers|Yes|Container Configuration Array|Array of Container objects, with a length of 1 to 10.|

**Table 2** Container parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|name|Yes|Container Name.|String, 1 to 32 characters in length; supports lowercase letters, digits, and other characters (-), and must start and end with a letter or digit; containerized application names must be unique.|
|cpuRequest|Yes|Number of CPU cores requested by the container.|Number, value range: 0.01 to 1000, accurate to two decimal places.|
|cpuLimit|No|Maximum number of CPU cores the container can use.|Number, value range: 0.01 to 1000, accurate to two decimal places, and must be greater than or equal to cpuRequest.|
|memRequest|Yes|Container Memory Request.|Number, value range: 4 to 1024000, integer only, in MiB.|
|memLimit|No|Maximum memory the container can use.|Number, value range: 4 to 1024000, integer only, in MiB, and must be greater than or equal to memRequest.|
|npu|No|Number of NPU cores requested by the container.|Number, value range: 0 to 32, integer only.|
|image|Yes|Image Name used. When using a third-party image registry, the full name must include the image registry server IP or domain name, port, project, and image name. For example, fd.fusiondirector.huawei\.com:443/library/ubuntu; if the user does not specify the image hostname and port, the containerized application will use the Docker public registry.|String, 1 to 256 characters in length; supports lowercase letters, uppercase letters, digits, and other characters (:-._/).|
|imageVersion|Yes|Image Version.|String, 1 to 32 characters in length, supports lowercase letters, uppercase letters, digits, and other characters (-._).|
|env|No|Environment variables configured in the container.|Array of EnvVar objects, supports a maximum of 256 key-value pairs.|
|userID|No|Container User ID for running the container.<br>If this parameter is not configured, the container runs as the user specified during image creation. If the user during image creation is not a numeric ID or the numeric ID is 0, the containerized application will fail to run after deployment. When running an inference container, driver devices are required, so the user ID cannot be configured.|Number, value range: 1 to 65535, cannot be configured as 0, meaning running the container as the root user is not supported.<br>When deploying an inference container, specify the user ID of the HwHiAiUser (usually 1000).|
|groupID|No|Container Group ID for running the container.<br>If this parameter is not configured, the container runs as the user group specified during image creation. If the user during image creation is not a numeric group ID or the numeric group ID is 0, the containerized application will fail to run after deployment. When running an inference container, driver devices are required, so the group ID cannot be configured.|Number, value range: 1 to 65535, cannot be configured as 0, meaning running the container as the root group is not supported.<br>When deploying an inference container, specify the group ID of the HwHiAiUser (usually 1000).|
|command|No|Container Command executed at container startup.|Array of strings, supports a maximum of 16 commands, each 1 to 256 characters in length, supports lowercase letters, uppercase letters, digits, spaces, and other characters (-/._), and must end with a letter or digit.|
|args|No|Container Args for the command executed at container startup.|Array of strings, supports a maximum of 16 arguments, each 1 to 256 characters in length, supports lowercase letters, uppercase letters, digits, spaces, and other characters (-/._=), and must end with a letter or digit.|
|containerPort|No|Host port and container port mapping configured for the container.|Array of ContainerPort objects, supports a maximum of 16 port groups.|
|hostPathVolumes|No|Host Path Mount Information for the container.<br>When creating an inference containerized application, mount paths must be configured; otherwise, the application may fail to run.|Array of HostPathVolume objects, supports a maximum of 256 groups.|

**Table 3** ContainerPort parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|name|Yes|Port Mapping Name|String, 1 to 32 characters in length. Supports lowercase letters, digits, and hyphens (-). Must start and end with a lowercase letter or digit.|
|proto|Yes|Network transport layer protocol specified by the port mapping|String, value must be TCP or UDP.|
|containerPort|Yes|Container Port|Integer, value range: 1 to 65535.|
|hostIP|Yes|Host IP address bound to the port mapping|String, must be a valid host IP address. Only IPv4 is supported. Cannot be all zeros or all 255s.|
|hostPort|Yes|Host Port Address|Integer, value range: 1024 to 65535.|

**Table 4** EnvVar parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|name|Yes|Environment Variable Name|String, 2 to 32 characters in length. Supports uppercase and lowercase letters, digits, and other characters (-._). Must start with an uppercase or lowercase letter and end with an uppercase or lowercase letter or digit.|
|value|Yes|Environment Variable Value|String, 1 to 512 characters in length. Supports uppercase and lowercase letters, digits, other characters (-._/:), and spaces.|

**Table 5** HostPathVolume parameter description

<table><thead align="left"><tr id="zh-cn_topic_0000001527041156_row12899173654011"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000001527041156_p17899836144010"><a name="zh-cn_topic_0000001527041156_p17899836144010"></a><a name="zh-cn_topic_0000001527041156_p17899836144010"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001527041156_p3900836124019"><a name="zh-cn_topic_0000001527041156_p3900836124019"></a><a name="zh-cn_topic_0000001527041156_p3900836124019"></a>Mandatory (Yes/No)</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000001527041156_p490033604011"><a name="zh-cn_topic_0000001527041156_p490033604011"></a><a name="zh-cn_topic_0000001527041156_p490033604011"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000001527041156_p5900143617405"><a name="zh-cn_topic_0000001527041156_p5900143617405"></a><a name="zh-cn_topic_0000001527041156_p5900143617405"></a>Value Requirement</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001527041156_row199001136144017"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001527041156_p1490053616402"><a name="zh-cn_topic_0000001527041156_p1490053616402"></a><a name="zh-cn_topic_0000001527041156_p1490053616402"></a>name</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001527041156_p189000364407"><a name="zh-cn_topic_0000001527041156_p189000364407"></a><a name="zh-cn_topic_0000001527041156_p189000364407"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001527041156_p18900436164014"><a name="zh-cn_topic_0000001527041156_p18900436164014"></a><a name="zh-cn_topic_0000001527041156_p18900436164014"></a>Mount volume name</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001527041156_p49001362404"><a name="zh-cn_topic_0000001527041156_p49001362404"></a><a name="zh-cn_topic_0000001527041156_p49001362404"></a>String, 1 to 32 characters in length; supports lowercase letters, digits, and hyphens (-). Must start and end with a letter or digit.</p>
<p id="zh-cn_topic_0000001527041156_p14964185819578"><a name="zh-cn_topic_0000001527041156_p14964185819578"></a><a name="zh-cn_topic_0000001527041156_p14964185819578"></a>Mount volume names must be unique within the same container.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001527041156_row990083694012"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001527041156_p4900203644015"><a name="zh-cn_topic_0000001527041156_p4900203644015"></a><a name="zh-cn_topic_0000001527041156_p4900203644015"></a>hostPath</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001527041156_p490014369404"><a name="zh-cn_topic_0000001527041156_p490014369404"></a><a name="zh-cn_topic_0000001527041156_p490014369404"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001527041156_p169001736194016"><a name="zh-cn_topic_0000001527041156_p169001736194016"></a><a name="zh-cn_topic_0000001527041156_p169001736194016"></a>Host path used by the container mount volume.</p>
<div class="note" id="zh-cn_topic_0000001527041156_note197621102286"><a name="zh-cn_topic_0000001527041156_note197621102286"></a><a name="zh-cn_topic_0000001527041156_note197621102286"></a><span class="notetitle"> NOTE: </span><div class="notebody"><p id="zh-cn_topic_0000001527041156_p97121126281"><a name="zh-cn_topic_0000001527041156_p97121126281"></a><a name="zh-cn_topic_0000001527041156_p97121126281"></a>Only host paths for the files or directories listed on the right can be configured for mounting. If users create a container image by referring to <a href="./common_operations.md#creating-an-inference-image">Creating an Inference Image</a> or <span id="zh-cn_topic_0000001527041156_ph1126911412338"><a name="zh-cn_topic_0000001527041156_ph1126911412338"></a><a name="zh-cn_topic_0000001527041156_ph1126911412338"></a><cite><a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100423566" target="_blank" rel="noopener noreferrer">Atlas 200I A2 Accelerator Module Ascend Software Quick Installation Guide</a></cite></span>, for the corresponding default mount paths inside the image, see Creating a Container Image -> Starting the Container step.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001527041156_p1935845132715"><a name="zh-cn_topic_0000001527041156_p1935845132715"></a><a name="zh-cn_topic_0000001527041156_p1935845132715"></a>Only host paths for the following files or directories can be configured for mounting.</p>
<a name="zh-cn_topic_0000001527041156_ul5305123819127"></a><a name="zh-cn_topic_0000001527041156_ul5305123819127"></a><ul id="zh-cn_topic_0000001527041156_ul5305123819127"><li>"/etc/sys_version.conf"</li><li>"/etc/hdcBasic.cfg"</li><li>"/usr/lib64/libaicpu_processer.so"</li><li>"/usr/lib64/libaicpu_prof.so"</li><li>"/usr/lib64/libaicpu_sharder.so"</li><li>"/usr/lib64/libadump.so"</li><li>"/usr/lib64/libtsd_eventclient.so"</li><li>"/usr/lib64/libaicpu_scheduler.so"</li><li>libcrypto.so.1.1<ul><li>On Ubuntu host OS: "/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1"</li><li>On openEuler host OS: "/usr/lib64/libcrypto.so.1.1.1m"</li></ul></li><li>/usr/lib64/libcrypto.so.3<ul><li>On Ubuntu host OS: "/usr/lib/aarch64-linux-gnu/libcrypto.so.3.0.12"</li><li>On openEuler host OS: "/usr/lib64/libcrypto.so.3.0.12"</li></ul></li><li>libyaml-0.so.2<ul><li>On Ubuntu host OS: "/usr/lib/aarch64-linux-gnu/libyaml-0.so.2.0.6"</li><li>On openEuler host OS: "/usr/lib64/libyaml-0.so.2.0.9"</li></ul></li><li>"/usr/lib64/libdcmi.so"</li><li>"/usr/lib64/libmpi_dvpp_adapter.so"</li><li>"/usr/lib64/libunified_timer.so"</li><li>"/usr/lib64/libmmpa.so"</li><li>"/usr/lib64/aicpu_kernels/"</li><li>"/usr/local/sbin/npu-smi"</li><li>"/usr/lib64/libstackcore.so"</li><li>"/usr/local/Ascend/driver/lib64"</li><li>"/var/slogd"</li><li>"/var/dmp_daemon"</li></ul>
<p id="zh-cn_topic_0000001527041156_p15774163614317"><a name="zh-cn_topic_0000001527041156_p15774163614317"></a><a name="zh-cn_topic_0000001527041156_p15774163614317"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000001527041156_row11157325124311"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001527041156_p16157152514437"><a name="zh-cn_topic_0000001527041156_p16157152514437"></a><a name="zh-cn_topic_0000001527041156_p16157152514437"></a>mountPath</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001527041156_p915732564310"><a name="zh-cn_topic_0000001527041156_p915732564310"></a><a name="zh-cn_topic_0000001527041156_p915732564310"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001527041156_p121576255431"><a name="zh-cn_topic_0000001527041156_p121576255431"></a><a name="zh-cn_topic_0000001527041156_p121576255431"></a>Mount path inside the container</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001527041156_p51581525114312"><a name="zh-cn_topic_0000001527041156_p51581525114312"></a><a name="zh-cn_topic_0000001527041156_p51581525114312"></a>A path string starting with "/", followed by uppercase and lowercase letters, digits, and other characters (_./-). Cannot contain "..". Total path length is 2 to 512 characters. Mount path names must be unique within the same container.</p>
</td>
</tr>
</tbody>
</table>

**Usage Example<a name="section9299576112"></a>**

Request example:

```bash
PATCH https://10.10.10.10:30035/edgemanager/v1/app
```

Request body:

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
            ],
           "hostPathVolumes":[
           ]
        }
    ],
    "description": "a test case for app-manager"
}
```

Response example:

```json
{
    "status":"00000000",
    "msg":"success"
}
```

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 6**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|

### Uninstalling a Containerized Application<a id="uninstalling-a-container-application"></a>

**Command Function<a name="section135251624204320"></a>**

Uninstalls a containerized application and stops the running state of the corresponding containerized application instance. This is a batch API that uninstalls the containerized application from one or more node groups in the specified node group ID list based on the specified container ID.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/deployment/batch-delete**

Request Header:

```http
Content-Type: application/json
```

Request Body:

```json
{
    "appID": AppId,
    "nodeGroupIds": [NodeGroupId]
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|appID|Yes|Containerized Application ID|Integer, minimum value 1, maximum value 2^32-1. Must be an existing application ID.|
|nodeGroupIds|Yes|Node Group ID List|Array, must contain unique node group IDs, with array length in [1,1024].|

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app/deployment/batch-delete
```

Request Body:

```json
{
    "appID": 1,
    "nodeGroupIds": [
      1, 2
    ]
}
```

Response Example:

```json
{
    "status":"00000000",
    "msg":"success"
}
```

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|Node Group IDs successfully uninstalled|
|failedInfos|Hash table, both key and value types are strings|The key is the Node Group ID that failed to uninstall, and the value is the reason for the failure of this ID|

### Deleting a Containerized Application<a id="deleting-a-container-application"></a>

**Command Function<a name="section135251624204320"></a>**

Deletes containerized applications. This is a batch API that deletes applications based on a specified array of Containerized Application IDs. MEF only allows deletion of applications that have not been deployed. If a deletion failure occurs, the response fields will include an array of failed IDs and an array of successful IDs.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/app/batch-delete**

Request Header:

```http
Content-Type: application/json
```

Request Body:

```json
{
    "appIDs": [AppId]
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Type|Description|Value Requirement|
|--|--|--|--|
|appIDs|Mandatory|Array of containerized application IDs|The array can contain a maximum of 1024 elements. Each number must be at least 1 and at most 2^32-1, and must be an existing application ID.|

**Example<a name="section9299576112"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/app/batch-delete
```

Request body:

```json
{
    "appIDs": [
        1,2
    ]
}
```

Response example:

```json
{
    "status":"00000000",
    "msg":"success"
}
```

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Batch operation result. If all batch operations are successful, this field is not returned.|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|Array|IDs of containerized applications successfully deleted.|
|failedInfos|Hash table, both key and value types are strings|The key is the ID of the containerized application that failed to be deleted, and the value is the reason for the failure of this ID.|

## Log Collection APIs<a id="ZH-CN_TOPIC_0000001640208710"></a>

### Overview<a name="ZH-CN_TOPIC_0000001640049426"></a>

MEF supports collecting and exporting MEF Edge logs to facilitate MEF Edge log troubleshooting and device status monitoring.

**Constraints<a name="section107641411309"></a>**

- Only one log export task is supported at a time
- Each log export task supports a maximum of 100 nodes
- The log download request must be completed within 2 hours
- Logs can be downloaded within one day after a successful export
- After MEF Center restarts, all tasks are deleted, and previously exported logs can no longer be downloaded
- After a new log collection task is successfully started, logs collected by previous tasks will be deleted.
- A maximum of 2000 completed log collection tasks can be stored. Excess tasks will be deleted.

**Log Collection Process Introduction<a name="section124811062447"></a>**

The following is an example of the process for MEF Edge software to call APIs for log collection.

1. Create a log collection task

    Create a log collection task through the RESTful API. A successful creation returns the log collection task ID. For details, see [Creating a Log Collection Task](#creating-a-log-collection-task).

    ```text
    https://{ip}:{port}/edgemanager/v1/logmgmt/dump/task
    ```

2. (Optional) Query the progress of the log collection task

    Query the progress of the log collection task using the log collection task ID. When the user task status is "succeed" or "partiallyFailed", MEF Center has successfully collected logs from MEF Edge. For details, see [Querying the Log Collection Task Progress](#querying-the-log-collection-task-progress).

    ```text
    https://{ip}:{port}/edgemanager/v1/logmgmt/dump/task/?taskId={taskId}
    ```

3. (Optional) Download the log collection file

    Download the log collection file. For details, see [Downloading a Log Collection File](#downloading-a-log-collection-file).

    ```text
    https://{ip}:{port}/edgemanager/v1/logmgmt/dump/download/edgeNodes.tar.gz
    ```

### Creating a Log Collection Task<a id="Creating a Log Collection Task"></a>

**Command Function<a name="section135251624204320"></a>**

Creates a log collection task.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/logmgmt/dump/task**

Request Body:

```json
{
    "module": module,
    "edgeNodes": edgeNodes
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|module|Yes|Type of logs to collect|String, value: edgeNode.|
|edgeNodes|Yes|Array of node IDs to collect|Array, 1 to 100 unique items; each item is an integer ranging from 1 to 2^32-1.|

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/logmgmt/dump/task
```

Request Body:

```json
{
    "module": "edgeNode",
    "edgeNodes": [2]
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success",
    "data": {
        "taskId": "dumpMultiNodesLog.413f66d069888b135e976c57"
    }
}
```

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Result of creating the log collection task|

**Table 3**  Data Field Description

|Parameter|Type|Description|
|--|--|--|
|taskId|String|ID of the successfully created task|

### Querying the Log Collection Task Progress<a id="querying-the-log-collection-task-progress"></a>

**Command Function<a name="section135251624204320"></a>**

Queries the progress of a log collection task.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/edgemanager/v1/logmgmt/dump/task?taskId=**_\{taskId\}_

Request Body: None

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|taskId|Yes|Task ID to Query|String, starting with dumpMultiNodesLog, with a supported length of 1 to 128 characters after dumpMultiNodesLog. Supports uppercase letters, lowercase letters, numbers, and other characters (-_.), for example, dumpMultiNodesLog.eca14a007f930e5689652a30.|

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/logmgmt/dump/task?taskId=dumpMultiNodesLog.eca14a007f930e5689652a30
```

Request Body: None

Response Example:

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

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Return value of a successful operation|

**Table 3** data Field Description

|Parameter|Type|Description|
|--|--|--|
|createdAt|String|Creation time|
|data|Object|Optional result|
|fileName|String|Filename of the collected logs|
|finishedAt|String|End time|
|progress|Number|Progress of the log collection task, with a maximum value of 100|
|reason|String|Reason why the task is in this state|
|startedAt|String|Start time|
|status|String|Task status, including the following states: <li>waiting: Waiting to execute</li><li>processing: Executing</li><li>aborting: Terminating</li><li>succeed: Task succeeded</li><li>failed: Task completely failed</li><li>partiallyFailed: Task partially failed</li>|
|taskId|String|Task ID to query|

### Downloading a Log Collection File<a id="downloading-a-log-collection-file"></a>

**Command Function<a name="section135251624204320"></a>**

Downloads log collection files. Users can download log files when the task status is succeed or partiallyFailed.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_<b>/edgemanager/v1/logmgmt/dump/download/</b>edgeNodes.tar.gz

Request Body: None

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/logmgmt/dump/download/edgeNodes.tar.gz
```

Request Body: None

Response Example: Direct File Download

Response Status Code: 200

## Alarm Event APIs<a name="ZH-CN_TOPIC_0000001691744913"></a>

### Overview<a name="ZH-CN_TOPIC_0000001691665661"></a>

MEF supports querying alarm or event information of MEF Edge and MEF Center by calling the API provided by MEF Center.

- Alarm Information: Information about issues that have occurred since a specific point in time.
- Event Information: Information about issues that occurred at a past point in time. The MEF Center database stores a maximum of 50 events per MEF Edge device.

**Querying Alarm and Event Information Process Overview<a name="section124811062447"></a>**

The following is an example of the process for MEF Center to query alarm information by calling APIs. The steps for querying event information are the same, except that the API called is different.

1. Query the alarm list

    Submit a task to query the alarm or event list through the RESTful API. For details about the API, see [Querying the Alarm List](#querying-the-alarm-list).

    ```text
    https://{ip}:{port}/alarmmanager/v1/alarms?pageNum={value1}&pageSize={value2}&ifCenter={value3}&sn={value4}&groupId={value5}
    ```

2. (Optional) Query alarm details

    Query detailed alarm information using the alarm identifier in the MEF Center database returned by the query list. For details about the API, see [Querying Alarm Details](#querying-alarm-details).

    ```text
    https://{ip}:{port}/alarmmanager/v1/alarm?id={value1}
    ```

### Querying the Alarm List<a id="querying-the-alarm-list"></a>

**Command Function<a name="section135251624204320"></a>**

Queries system alarm data. The URL parameters specify the pagination control conditions, the alarm types to query, and the query operation type for this request, returning the alarm information already created in the MEF Center database.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/alarmmanager/v1/alarms?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&ifCenter=**_\{value3\}_**&sn=**_\{value4\}_**&groupId=**_\{value5\}_

Request Body: None

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

<table><thead align="left"><tr id="row1643491410534"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p1943421410537"><a name="p1943421410537"></a><a name="p1943421410537"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p113476328527"><a name="p113476328527"></a><a name="p113476328527"></a>Mandatory (Yes/No)</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p343413143534"><a name="p343413143534"></a><a name="p343413143534"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p18434201414535"><a name="p18434201414535"></a><a name="p18434201414535"></a>Value Requirement</p>
</th>
</tr>
</thead>
<tbody><tr id="row11236115220615"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p207789294419"><a name="p207789294419"></a><a name="p207789294419"></a>pageNum</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1477862910411"><a name="p1477862910411"></a><a name="p1477862910411"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1977814296415"><a name="p1977814296415"></a><a name="p1977814296415"></a>Page Number (Ordinal)</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p187782291042"><a name="p187782291042"></a><a name="p187782291042"></a>An integer ranging from 1 to 2^31-1.</p>
</td>
</tr>
<tr id="row19711641046"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1777710299410"><a name="p1777710299410"></a><a name="p1777710299410"></a>pageSize</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1177713299410"><a name="p1177713299410"></a><a name="p1177713299410"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p197771429244"><a name="p197771429244"></a><a name="p197771429244"></a>Page Size</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6777029447"><a name="p6777029447"></a><a name="p6777029447"></a>An integer ranging from 1 to 100.</p>
</td>
</tr>
<tr id="row10812924556"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p14778192919414"><a name="p14778192919414"></a><a name="p14778192919414"></a>ifCenter</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p0778182919415"><a name="p0778182919415"></a><a name="p0778182919415"></a>No</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p147785291248"><a name="p147785291248"></a><a name="p147785291248"></a>Indicates whether the queried node type is a <span id="ph388619511264"><a name="ph388619511264"></a><a name="ph388619511264"></a>MEF Center</span> node</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p17778152918411"><a name="p17778152918411"></a><a name="p17778152918411"></a>The value is true or false.</p>
<div class="note" id="note2862431999"><a name="note2862431999"></a><a name="note2862431999"></a><span class="notetitle">[!NOTE] Note</span><div class="notebody"><a name="ul172619261897"></a><a name="ul172619261897"></a><ul id="ul172619261897"><li>When ifCenter is set to <span class="parmvalue" id="parmvalue1231617358148"><a name="parmvalue1231617358148"></a><a name="parmvalue1231617358148"></a>"true"</span>, groupId and sn are ignored.</li><li>When ifCenter is set to <span class="parmvalue" id="parmvalue1996315374144"><a name="parmvalue1996315374144"></a><a name="parmvalue1996315374144"></a>"false"</span> and neither sn nor groupId is provided, all <span id="ph16491162049"><a name="ph16491162049"></a><a name="ph16491162049"></a>MEF Edge</span> node alarms are queried with pagination.</li></ul>
</div></div>
</td>
</tr>
<tr id="row1274362319416"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p37781229846"><a name="p37781229846"></a><a name="p37781229846"></a>sn</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1977812291141"><a name="p1977812291141"></a><a name="p1977812291141"></a>Optional, choose one</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p177785291045"><a name="p177785291045"></a><a name="p177785291045"></a>Specifies the serial number of the <span id="ph174153115216"><a name="ph174153115216"></a><a name="ph174153115216"></a>MEF Edge</span> device to query alarm information on that node</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p67781291347"><a name="p67781291347"></a><a name="p67781291347"></a>Supports lowercase letters, uppercase letters, digits, underscores, and hyphens; cannot start or end with an underscore or hyphen; maximum length is 64 bytes.</p>
</td>
</tr>
<tr id="row127648241644"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p477811292046"><a name="p477811292046"></a><a name="p477811292046"></a>groupId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p12778132911414"><a name="p12778132911414"></a><a name="p12778132911414"></a>Specifies the groupId to query alarm information on the nodes belonging to that node group</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1877852915417"><a name="p1877852915417"></a><a name="p1877852915417"></a>A 32-bit unsigned number. The minimum value is 1, and the maximum value is 2^32-1.</p>
</td>
</tr>
</tbody>
</table>

> [!NOTE]
>
>- When the three parameters ifCenter, sn, and groupId are all empty, alarms for all nodes are queried with pagination based on pageNum and pageSize.
>- When only sn or groupId is provided, ifCenter defaults to "false".

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/alarms?pageNum=1&pageSize=100&ifCenter=false&sn=xxxxxxxxxxxxxx
```

Request Body: None

Response Example:

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

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|Object|Query result|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|records|Array|Array of objects for paginated queries|
|total|Number|Total Results|

**Table 4** Description of the records field

|Parameter|Type|Description|
|--|--|--|
|id|Number|Alarm ID|
|alarmType|String|Alarm type. The value is alarm|
|createAt|String|Alarm creation time|
|ip|String|Device IP address|
|serialNumber|String|For MEF Edge alarms, this is the device serial number. For MEF Center alarms, this is an empty string|
|resource|String|Alarm source|
|severity|String|Alarm severity: <li>MINOR: Minor alarm</li><li>MAJOR: Major alarm</li><li>CRITICAL: Critical alarm</li>|

### Querying Alarm Details<a id="querying-alarm-details"></a>

**Command Function<a name="section135251624204320"></a>**

Queries alarm details. Returns the detailed information of the alarm based on the specified alarm ID.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/alarmmanager/v1/alarm?id=**<i>{value1}</i>

Request Body: None

**URL Parameters<a name="section1774293413516"></a>**

**Table 1**  Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|id|Yes|Alarm ID|An integer with a minimum value of 1 and a maximum value of 2^32-1.|

**Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/alarm?id=1430946159
```

Request Body: None

Response Example:

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

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** data Field Description

|Parameter|Type|Description|
|--|--|--|
|alarmId|String|Alarm ID|
|alarmName|String|Alarm name|
|alarmType|String|Alarm type. The value is alarm|
|createAt|String|Alarm creation time|
|detailedInformation|String|Detailed alarm information|
|id|Number|Alarm identifier|
|impact|String|Alarm impact description|
|ip|String|Device IP address|
|serialNumber|String|Device serial number for MEF Edge alarms, and an empty string for MEF Center alarms|
|perceivedSeverity|String|Alarm severity:<li>MINOR: Minor alarm</li><li>MAJOR: Major alarm</li><li>CRITICAL: Critical alarm</li>|
|reason|String|Alarm cause|
|resource|String|Alarm source|
|suggestion|String|Alarm handling suggestion|

### Querying the Event List<a name="ZH-CN_TOPIC_0000001691744917"></a>

**Command Function<a name="section135251624204320"></a>**

Queries system event data. The URL parameters specify the pagination control conditions, the event types to query, and the query operation type for this request. Returns the event information that has been created in the MEF Center database.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/alarmmanager/v1/events?pageNum=**_\{value1\}_**&pageSize=**_\{value2\}_**&ifCenter=**_\{value3\}_**&sn=**_\{value4\}_**&groupId=**_\{value5\}_

Request Body: None

**URL Parameters<a name="section1774293413516"></a>**

**Table 1**  Parameter description

<table><thead align="left"><tr id="row1643491410534"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p1943421410537"><a name="p1943421410537"></a><a name="p1943421410537"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p113476328527"><a name="p113476328527"></a><a name="p113476328527"></a>Mandatory (Yes/No)</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p343413143534"><a name="p343413143534"></a><a name="p343413143534"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p18434201414535"><a name="p18434201414535"></a><a name="p18434201414535"></a>Value Requirement</p>
</th>
</tr>
</thead>
<tbody><tr id="row11236115220615"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p207789294419"><a name="p207789294419"></a><a name="p207789294419"></a>pageNum</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1477862910411"><a name="p1477862910411"></a><a name="p1477862910411"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1977814296415"><a name="p1977814296415"></a><a name="p1977814296415"></a>Page Number (Ordinal)</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p187782291042"><a name="p187782291042"></a><a name="p187782291042"></a>An integer ranging from 1 to 2^31-1.</p>
</td>
</tr>
<tr id="row19711641046"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1777710299410"><a name="p1777710299410"></a><a name="p1777710299410"></a>pageSize</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1177713299410"><a name="p1177713299410"></a><a name="p1177713299410"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p197771429244"><a name="p197771429244"></a><a name="p197771429244"></a>Page Size</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6777029447"><a name="p6777029447"></a><a name="p6777029447"></a>An integer ranging from 1 to 100.</p>
</td>
</tr>
<tr id="row10812924556"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p14778192919414"><a name="p14778192919414"></a><a name="p14778192919414"></a>ifCenter</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p0778182919415"><a name="p0778182919415"></a><a name="p0778182919415"></a>No</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p147785291248"><a name="p147785291248"></a><a name="p147785291248"></a>Indicates whether the query node type is a <span id="ph388619511264"><a name="ph388619511264"></a><a name="ph388619511264"></a>MEF Center</span> node</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p17778152918411"><a name="p17778152918411"></a><a name="p17778152918411"></a>The value is true or false.</p>
<div class="note" id="note2862431999"><a name="note2862431999"></a><a name="note2862431999"></a><span class="notetitle">[!NOTE] Note</span><div class="notebody"><a name="ul172619261897"></a><a name="ul172619261897"></a><ul id="ul172619261897"><li>When ifCenter is set to <span class="parmvalue" id="parmvalue1231617358148"><a name="parmvalue1231617358148"></a><a name="parmvalue1231617358148"></a>"true"</span>, groupId and sn are ignored.</li><li>When ifCenter is set to <span class="parmvalue" id="parmvalue1996315374144"><a name="parmvalue1996315374144"></a><a name="parmvalue1996315374144"></a>"false"</span> and sn and groupId are not provided, all <span id="ph16491162049"><a name="ph16491162049"></a><a name="ph16491162049"></a>MEF Edge</span> node alarms are queried by page.</li></ul>
</div></div>
</td>
</tr>
<tr id="row1274362319416"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p37781229846"><a name="p37781229846"></a><a name="p37781229846"></a>sn</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1977812291141"><a name="p1977812291141"></a><a name="p1977812291141"></a>Optional, choose one</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p177785291045"><a name="p177785291045"></a><a name="p177785291045"></a>Specifies the <span id="ph174153115216"><a name="ph174153115216"></a><a name="ph174153115216"></a>MEF Edge</span> device serial number to query alarm information on that node</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p67781291347"><a name="p67781291347"></a><a name="p67781291347"></a>Supports lowercase letters, uppercase letters, digits, underscores, and hyphens; cannot start or end with an underscore or hyphen; maximum length is 64 bytes.</p>
</td>
</tr>
<tr id="row127648241644"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p477811292046"><a name="p477811292046"></a><a name="p477811292046"></a>groupId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p12778132911414"><a name="p12778132911414"></a><a name="p12778132911414"></a>Specifies the groupId to query alarm information on nodes belonging to that node group</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1877852915417"><a name="p1877852915417"></a><a name="p1877852915417"></a>A 32-bit unsigned number. The minimum value is 1, and the maximum value is 2^32-1.</p>
</td>
</tr>
</tbody>
</table>

> [!NOTE]
>
>- When the three parameters ifCenter, sn, and groupId are all empty, events for all nodes are queried by page based on pageNum and pageSize.
>- When only sn or groupId is provided, ifCenter defaults to "false".

**Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/events?pageNum=1&pageSize=100&ifCenter=false&groupId=1
```

Request Body: None

Response Example:

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

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|Object|Query result|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|records|Array|Array of objects for paginated queries|
|total|Number|Total Results|

**Table 4** Description of the records field

|Parameter|Type|Description|
|--|--|--|
|id|Number|Event sequence number|
|alarmType|String|Event type, with the value event|
|createAt|String|Event creation time|
|ip|String|Device IP address|
|serialNumber|String|For MEF Edge events, this is the device serial number; for MEF Center events, this is an empty string|
|resource|String|Source of the event|
|severity|String|Event severity:<li>MINOR: Minor event</li><li>MAJOR: Major event</li><li>CRITICAL: Critical event</li><li>OK: Normal event</li>|

### Querying Event Details<a name="ZH-CN_TOPIC_0000001643632794"></a>

**Command Function<a name="section135251624204320"></a>**

Viewes event details. It can return detailed information for the specified event based on the event sequence number.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL**: **https://**_\{ip\}:\{port\}_**/alarmmanager/v1/event?id=**<i>{value1}</i>

Request body: None

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|id|Yes|Event sequence number|An integer with a minimum value of 1 and a maximum value of 2^32-1.|

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/alarmmanager/v1/event?id=1430946159
```

Request Body: None

Response Example:

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

Response Status Code: 200

**Output Description<a name="section99341441896"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Query result|

**Table 3** data field description

|Parameter|Type|Description|
|--|--|--|
|alarmId|String|Event ID|
|alarmName|String|Event name|
|alarmType|String|Event type. The value is event.|
|createAt|String|Event creation time|
|detailedInformation|String|Event details|
|id|Number|Event sequence number|
|ip|String|Device IP address|
|impact|String|Event impact description|
|serialNumber|String|For MEF Edge events, this is the device serial number. For MEF Center events, this is an empty string.|
|perceivedSeverity|String|Event severity:<li>OK: Normal event</li><li>MINOR: Minor event</li><li>MAJOR: Major event</li><li>CRITICAL: Critical event</li>|
|reason|String|Event cause|
|resource|String|Event source|
|suggestion|String|Event handling suggestion|

## Configuration APIs<a id="ZH-CN_TOPIC_0000001526721288"></a>

### Overview<a id="configuration-api-introduction"></a>

The configuration API supports users in configuring the root certificates of third-party software repositories and image repositories, as well as configuring image download information. To use a third-party image repository, users must first call the [Importing the Root Certificate](#importing-the-root-certificate) API to import the image repository certificate, and then call the [Configuring Image Download Information](#configuring-image-download-information) API to configure the image download information.

### Importing the Root Certificate<a id="importing-the-root-certificate"></a>

**Command Function<a name="section135251624204320"></a>**

Imports the corresponding root certificates for third-party software repositories and image repositories. The root certificates of the software repository and image repository must be imported before using containerized applications.

> [!NOTE]
>
>- Repeatedly calling this API will update the root certificate.
>- The validity period of the root certificate should be greater than the [certificate alarm detection cycle](./common_operations.md#ZH-CN_TOPIC_0000001722295489) (default value is 7).
>- Re-importing a root certificate will back up the previously imported certificate.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/certmanager/v1/certificates/import**

Request Body:

```json
{
    "certName": certname,
    "cert": cert
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Type|Description|Value Requirement|
|--|--|--|--|
|certName|String|Purpose of the imported certificate|The value is one of the following parameters.<li>software: Software repository root certificate</li><li>image: Image repository root certificate</li>|
|cert|String|Base64-encoded PEM root certificate|<li>Must be a base64-encoded root certificate.</li><li>The certificate must be in PEM format.</li><li>The signature in the root CA certificate must be correct.</li><li>The root CA certificate must be within its validity period.</li><li>The certificate must be an X.509 V3 digital certificate. The "Basic Constraints" extension of the root CA certificate must indicate "CA", and the "Key Usage" extension must include "Certificate Signing".</li><li>The key algorithm must be RSA with a length of at least 3072, or ECDSA with a length of at least 256. The digest algorithm must be SHA256, SHA384, or SHA512.</li>|

**Example<a name="section9299576112"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/certmanager/v1/certificates/import
```

Request Body:

```json
{
    "certName": "software",
    "cert": "xxxxxxxxxxxxxxxxxxx..."
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "import certificate success"
}
```

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|

### Configuring Image Download Information<a id="configuring-image-download-information"></a>

**Command Function<a name="section135251624204320"></a>**

Configures the third-party image repository address and account credentials. The repository server address supports domain names or IP addresses. Repeatedly calling this API will update the existing image download information configuration.

> [!NOTICE]
>
>- The image download information will be transmitted by MEF Center to K8s and KubeEdge, which will store the data in K8s and the edgecore database of the MEF Edge device. Users can enhance the security of the image repository credentials by customizing K8s and KubeEdge as needed.
>- Users are advised to use trusted third-party image repositories. Configuring an untrusted third-party image repository address may lead to insecure transmission processes.
>- When using a domain name, you need to configure the mapping between the domain name and the IP address in the `/etc/hosts` file on the MEF Edge device. For details, see [Configuring Local Domain Name Mapping](./common_operations.md#ZH-CN_TOPIC_0000001722295397).

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/image/config**

Request Body:

```json
{
    "domain": domain,
    "ip": ip,
    "port": port,
    "account": account,
    "password": password
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value|
|--|--|--|--|
|domain|Optional|Domain name of the image repository server|String, a valid domain name. Supported length is 3 to 63 characters, consisting of uppercase and lowercase letters, digits, and symbols (.-). It must start and end with an uppercase or lowercase letter or digit, and cannot be all digits. The value cannot be localhost.<br>At least one of the domain and IP parameters must be provided. If both are provided, the domain name takes precedence.|
|ip|Optional|IP address of the image repository server|String, a valid IPv4 address. It cannot be all zeros or all 255s, cannot be the loopback address 127.0.0.1, and cannot be the host address of the MEF Edge device.<br>At least one of the domain and IP parameters must be provided. If both are provided, the domain name takes precedence.|
|port|Yes|Port number for the image repository service|Integer, ranging from 1 to 65535.|
|account|Yes|Image download account|String, maximum length of 256 characters. Supports uppercase and lowercase letters, digits, underscores (_), and hyphens (-). Underscores and hyphens cannot be at the beginning or end.|
|password|Yes|Image download password|Byte array, with an array length of [8, 20]. The image repository password does not support English colons.<div class="note"><span class="notetitle">[!NOTE] NOTE</span><div class="notebody">It is recommended that the password complexity meet the following requirements. If the configured password does not comply with the following rules, security risks may exist.<li>The password must be at least 8 characters long.</li><li>The password must contain at least two of the following character types:<ul><li>At least one lowercase letter.</li><li>At least one uppercase letter.</li><li>At least one digit.</li><li>At least one special character: `~!@#$%^&*()-_=+\|[{}];:'",<.>/? and space</li></ul></li><li>The password cannot be the same as the account.</li></div></div>|

**Example<a name="section9299576112"></a>**

Request example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/image/config
```

Request body:

```json
{
    "domain": "xxx.huawei.com",
    "ip": "10.10.10.10",
    "port": 6443,
    "account": "ImageRepository",
    "password": [72, 117, 97, 119, 101, 105, 49, 50, 35, 36]
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|

### Deleting the Root Certificate<a name="ZH-CN_TOPIC_0000001526881244"></a>

**Command Function<a name="section135251624204320"></a>**

Deletes root certificates that have been imported into third-party software repositories and image repositories.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/certmanager/v1/certificates/delete-cert**

Request Body:

```json
{
    "type": type
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Request Parameters

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|type|Yes|Type of certificate|The value must be one of software or image.<li>software: Software repository root certificate</li><li>image: Image repository root certificate</li>|

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/certmanager/v1/certificates/delete-cert
```

Request Body:

```json
{
    "type": "software"
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "delete ca file success"
}
```

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|

### Exporting the Root Certificate<a id="ZH-CN_TOPIC_0000001526721348"></a>

**Command Function<a name="section135251624204320"></a>**

Exports the root certificate of MEF Center.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_<b>/certmanager/v1/export?certName=</b>hub\_svr

**URL Parameters<a name="section1774293413516"></a>**

**Table 1** URL parameters

|Parameter|Mandatory|Description|Value Requirement|
|--|--|--|--|
|certName|Yes|Target type for exporting the root certificate|Currently, only the hub_svr type is supported for export. hub_svr: MEF Center root certificate|

**Usage Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/certmanager/v1/export?certName=hub_svr
```

Successful response example: Output the root.crt file

```text
-----BEGIN CERTIFICATE-----
xxxxxxxxxxxxxxxxxxx...
-----END CERTIFICATE-----
```

### Obtaining the Cloud-Edge Authentication Token<a id="ZH-CN_TOPIC_0000001566531326"></a>

**Command Function<a name="section135251624204320"></a>**

Obtains the token for cloud-edge authentication between MEF Center and MEF Edge.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_**/edgemanager/v1/token**

**Usage Example<a name="section9299576112"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/token
```

Response example:

```json
{
    "status": "00000000",
    "msg": "export token success",
    "data": "xxxxxxxxxxxxxxxxxxxx..."
}
```

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 1** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|String|Cloud-edge authentication token, valid for 7 days, automatically expires upon expiration|

### Obtaining Integrator Certificate Information<a id="ZH-CN_TOPIC_0000001621432513"></a>

**Syntax<a name="section6901955114320"></a>**

Operation Type: **GET**

**URL: https://**_\{ip\}:\{port\}_**/certmanager/v1/certificates/info?certName=**_\{**certName**\}_

**Request Parameters<a name="section1774293413516"></a>**

**Table 1**  Parameter description for obtaining integrator certificate information

|Parameter|Mandatory (Yes/No)|Description|Value|
|--|--|--|--|
|certName|Yes|Certificate name|String. Currently, only north is supported.|

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/certmanager/v1/certificates/info?certName=north
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|Integrator certificate information|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|FingerPrint|String|Certificate fingerprint|
|FingerPrintAlgorithm|String|Certificate fingerprint algorithm|
|Issuer|String|Certificate issuer|
|SerialNumber|String|Certificate serial number|
|Subject|String|Certificate holder|
|Validity|Object|Certificate validity, including NotBefore and NotAfter|
|NotAfter|String|Validity end time|
|NotBefore|String|Validity start time|

### Importing the CRL<a id="ZH-CN_TOPIC_0000001621540625"></a>

**Command Function<a name="section135251624204320"></a>**

Imports the corresponding Certificate Revocation List (CRL) chain of root certificates for service platforms, software repositories, image repositories, and other entities that interconnect with MEF Center. This can revoke RESTful request access permissions for interconnected third-party platform certificates that have been revoked. Repeatedly calling this API will update the imported CRL. After importing the CRL for the interconnected service platform, the user must manually restart MEF Center for it to take effect.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**URL: https://**_\{ip\}:\{port\}_**/certmanager/v1/crl/import**

Request body:

```json
{
    "crlName": crlName,
    "crl": crl
}
```

**Request Parameters<a name="section1774293413516"></a>**

**Table 1** Parameter description

|Parameter|Type|Description|Value Requirement|
|--|--|--|--|
|crlName|String|Purpose of the imported revocation list|<li>north: Integration party certificate revocation list</li><li>software: Software repository root certificate revocation list</li><li>image: Image repository certificate revocation list</li>|
|crl|String|Base64-encoded PEM format certificate revocation list chain|<li>Must be a base64-encoded CRL.</li><li>The CRL must be in PEM format.</li><li>The CRL must have the same number of levels as the certificate chain corresponding to crlName.</li><li>The CRL certificate must be within its validity period.</li><li>The CRL must contain the revocation list issued by each level of the certificate chain corresponding to the imported crlName.</li>|

**Example<a name="section9299576112"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/certmanager/v1/crl/import
```

Request Body:

```json
{
    "crlName": software,
    "crl": crl
}
```

Response Example:

```json
{
    "status": "00000000",
    "msg": "success"
}
```

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|

## Upgrade APIs<a id="ZH-CN_TOPIC_0000001527041092"></a>

### Overview<a name="ZH-CN_TOPIC_0000001577601037"></a>

MEF supports online upgrade, same-version upgrade, and version rollback of MEF Edge through the MEF Center software upgrade API.

**Pre-Upgrade Preparation<a name="section84371423133717"></a>**

1. Prepare the software to be upgraded. For online upgrades, the software to be upgraded must be prepared through a properly functioning third-party software repository. Users must ensure network connectivity between the MEF Edge device and the software repository, and ensure that the HTTPS URL specified when issuing the software download message can access the software to be upgraded. For details on software repository preparation and integration, see [Integrating Software Repository](./usage.md#ZH-CN_TOPIC_0000001674256310) and [Integrating Image Repository](./usage.md#ZH-CN_TOPIC_0000001674416006).
2. (Optional) Query node details. During a software upgrade, the request must include the device serial number of the edge device. If the user is unsure of the device serial number, they can query node details through the RESTful API to view the device serial number of the corresponding node. For details, see [Querying Node Details](#querying-node-details).

**Upgrade Process Overview<a name="section131291530113720"></a>**

The following is an example of the MEF Edge software upgrade process via API calls.

1. Send a software download message

    Send a software download message through the RESTful API to trigger the edge device to download the software to be upgraded. For details on the API for sending a software download message, see [Downloading Software](#downloading-software).

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/download
    ```

2. (Optional) Query software download progress

    When sending a software download message, you can query the software download progress through the RESTful API. For details on the API for querying software download progress, see [Querying the Software Download Progress](#querying-the-software-download-progress).

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/download-progress?serialNumber={value}
    ```

3. (Optional) Query software information

    After triggering the software download, you can query the current MEF Edge software information through the RESTful API, including the current software version and the version to be upgraded. For details about the software information query API, see [Querying Software Information](#querying-software-information).

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/version-info?serialNumber={value}
    ```

4. Deliver a software upgrade message

    Deliver a software upgrade message through the RESTful API to trigger the edge device to upgrade the MEF Edge software. For details about the software upgrade message delivery API, see [Upgrading Software](#upgrading-software).

    ```text
    https://{ip}:{port}/edgemanager/v1/software/edge/upgrade
    ```

### Downloading Software<a id="downloading-software"></a>

**Command Function<a name="section135251624204320"></a>**

Sends a message to download MEF Edge software for software download, configures the third-party software repository address and account credentials, and downloads the prepared software to be upgraded from the third-party software repository.

**Syntax<a name="section6901955114320"></a>**

Operation Type: **POST**

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/download**

Request Body:

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

**Request Parameters<a name="section1774293413516"></a>**

**Table 1**  Parameter description

<table><thead align="left"><tr id="row153763524260"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p837615213264"><a name="p837615213264"></a><a name="p837615213264"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p15376205202612"><a name="p15376205202612"></a><a name="p15376205202612"></a>Mandatory (Yes/No)</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p193761052142613"><a name="p193761052142613"></a><a name="p193761052142613"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p3376195215262"><a name="p3376195215262"></a><a name="p3376195215262"></a>Value Requirement</p>
</th>
</tr>
</thead>
<tbody><tr id="row137821209392"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p184721335322"><a name="p184721335322"></a><a name="p184721335322"></a>serialNumbers</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p17831206397"><a name="p17831206397"></a><a name="p17831206397"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1568916172332"><a name="p1568916172332"></a><a name="p1568916172332"></a>Device serial number</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1088711633416"><a name="p1088711633416"></a><a name="p1088711633416"></a>Array. Supports lowercase letters, uppercase letters, digits, underscores, and hyphens. Cannot start or end with an underscore or hyphen. Maximum length: 64 bytes. Array length: [1,2048].</p>
</td>
</tr>
<tr id="row183821834123910"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p144727383214"><a name="p144727383214"></a><a name="p144727383214"></a>softwareName</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1538243403917"><a name="p1538243403917"></a><a name="p1538243403917"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p8689161717331"><a name="p8689161717331"></a><a name="p8689161717331"></a>Name of the software to download</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p48161222151911"><a name="p48161222151911"></a><a name="p48161222151911"></a>Must be <span class="parmvalue" id="parmvalue361983132513"><a name="parmvalue361983132513"></a><a name="parmvalue361983132513"></a>"MEFEdge"</span>.</p>
</td>
</tr>
<tr id="row945705083116"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1845819504311"><a name="p1845819504311"></a><a name="p1845819504311"></a>downloadInfo</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p745865043119"><a name="p745865043119"></a><a name="p745865043119"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1745875043114"><a name="p1745875043114"></a><a name="p1745875043114"></a>Download information</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p194581550153117"><a name="p194581550153117"></a><a name="p194581550153117"></a>File download information from the software repository.</p>
</td>
</tr>
<tr id="row4160114413115"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p164722031320"><a name="p164722031320"></a><a name="p164722031320"></a>package</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p161611944183113"><a name="p161611944183113"></a><a name="p161611944183113"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p16689141763319"><a name="p16689141763319"></a><a name="p16689141763319"></a>Software file package</p>
<p id="p1068981783318"><a name="p1068981783318"></a><a name="p1068981783318"></a></p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p159771941173311"><a name="p159771941173311"></a><a name="p159771941173311"></a>String. Maximum URL length: 512 bytes. Must not contain special characters such as "\n!\\|;$&lt;&gt;@` ". The protocol must be HTTPS and the request method must be GET.</p>
<div class="note" id="note134513320143"><a name="note134513320143"></a><a name="note134513320143"></a><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><a name="ul17843314116"></a><a name="ul17843314116"></a><ul id="ul17843314116"><li>If a domain name is included, it must be a valid domain name conforming to the regular expression pattern ^[a-zA-Z0-9][a-zA-Z0-9.-]{1,61}[a-zA-Z0-9]$. The value cannot be all digits and cannot be localhost.</li><li>If an IP address is included, it must be a valid IPv4 address. It cannot be a loopback address (127.0.0.1) or the IP address of a <span id="ph1564518021910"><a name="ph1564518021910"></a><a name="ph1564518021910"></a>MEF Edge</span> node.</li></ul>
</div></div>
</td>
</tr>
<tr id="row32121849203113"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p34721736323"><a name="p34721736323"></a><a name="p34721736323"></a>userName</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1121374911312"><a name="p1121374911312"></a><a name="p1121374911312"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1068981714338"><a name="p1068981714338"></a><a name="p1068981714338"></a>Download account</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p10213154963115"><a name="p10213154963115"></a><a name="p10213154963115"></a>String. Length: 6–32. Can only contain lowercase letters, uppercase letters, and digits.</p>
</td>
</tr>
<tr id="row2851651163112"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1047212318323"><a name="p1047212318323"></a><a name="p1047212318323"></a>password</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p138565123119"><a name="p138565123119"></a><a name="p138565123119"></a>Yes</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p1168921714333"><a name="p1168921714333"></a><a name="p1168921714333"></a>Download password</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p208535118313"><a name="p208535118313"></a><a name="p208535118313"></a>Byte array. The download password must be included in the submitted information. Array length: [8,20].</p>
<div class="note" id="note107555422113"><a name="note107555422113"></a><a name="note107555422113"></a><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><div class="p" id="p82854163268"><a name="p82854163268"></a><a name="p82854163268"></a>It is recommended that the password complexity meet the following requirements. If the set password does not comply with the following rules, there may be security risks.<a name="ul753618188215"></a><a name="ul753618188215"></a><ul id="ul753618188215"><li>The password must be at least 8 characters long.</li><li>The password must contain a combination of at least two of the following character types:<a name="ul682642514227"></a><a name="ul682642514227"></a><ul id="ul682642514227"><li>At least one lowercase letter.</li><li>At least one uppercase letter.</li><li>At least one digit.</li><li>At least one special character: `~!@#$%^&amp;*()-_=+\|[{}];:'",&lt;.&gt;/? and space.</li></ul>
</li><li>The password cannot be the same as the account name.</li></ul>
</div>
</div></div>
</td>
</tr>
</tbody>
</table>

**Usage Example<a name="section9299576112"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/software/edge/download
```

Request Body:

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

Response Example:

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

Response Status Code: 200

**Output Description<a name="section127921251728"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|-|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|String List|List of successful serial numbers|
|failedInfos|Hash table, both key and value are strings|The key is the failed serial number, and the value is the reason for failure|

### Querying the Software Download Progress<a id="querying-the-software-download-progress"></a>

**Command Function<a name="section19727974820"></a>**

Queries the download progress.

**Syntax<a name="section101971341114914"></a>**

Operation Type: **GET**

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/download-progress?serialNumber=**_\{value\}_

**URL Parameters<a name="section12700104818484"></a>**

**Table 1** URL Parameters

|Parameter|Type|Description|Value Requirement|
|--|--|--|--|
|serialNumber|Mandatory|Device Serial Number|String, supports lowercase letters, uppercase letters, digits, underscores, and hyphens. Cannot start or end with an underscore or hyphen, maximum length of 64 bytes.|

**Usage Example<a name="section3878166125016"></a>**

Request Example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/software/edge/download-progress?serialNumber=2102312NSF10K8000130
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section1390714110514"></a>**

**Table 2**  Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description information|
|data|Object|-|

**Table 3**  Data field description

|Parameter|Type|Description|
|--|--|--|
|msg|String|Specific status information|
|progress|uint64|Download progress value|
|res|String|Download result, value is success or failed|

### Querying Software Information<a id="querying-software-information"></a>

**Command Function<a name="section82653133719"></a>**

Queries software information.

**Syntax<a name="section16835327372"></a>**

Operation Type: **GET**

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/version-info?serialNumber=**_\{value\}_

**URL Parameters<a name="section1130318261113"></a>**

**Table 1** URL parameter

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|serialNumber|Yes|Device serial number|String, supports lowercase letters, uppercase letters, digits, underscores, and hyphens. Cannot start or end with an underscore or hyphen. Maximum length is 64 bytes.|

**Usage Example<a name="section569514203919"></a>**

Request example:

```bash
GET https://10.10.10.10:30035/edgemanager/v1/software/edge/version-info?serialNumber=2102312NSF10K8000130
```

Response example:

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

Response Status Code: 200

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|-|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|InactiveVersion|String|Version information to be activated|
|Name|String|Name of the software to be activated|
|Version|String|Currently running software version|

### Upgrading Software<a id="upgrading-software"></a>

**Command Function<a name="section1612804915619"></a>**

Sends an upgrade message to upgrade the MEF Edge software.

**Syntax<a name="section121284496614"></a>**

Operation Type: POST

**https://**_\{ip\}:\{port\}_**/edgemanager/v1/software/edge/upgrade**

Request Body:

```json
{
    "serialNumbers": ["xxxxxxxxxxxxx"],
    "softwareName": "MEFEdge"
}
```

**Request Parameters<a name="section263114675317"></a>**

**Table 1** Parameter description

|Parameter|Mandatory (Yes/No)|Description|Value Requirement|
|--|--|--|--|
|serialNumbers|Yesy|Device serial number|Array, supports lowercase letters, uppercase letters, digits, underscores, and hyphens. Cannot start or end with an underscore or hyphen. Maximum length is 64 bytes. Array length is [1,2048].|
|softwareName|Yes|Name of the software to be downloaded|Value: MEFEdge.|

**Usage Example<a name="section9550421315"></a>**

Request Example:

```bash
POST https://10.10.10.10:30035/edgemanager/v1/software/edge/upgrade
```

Request Body:

```json
{
    "serialNumbers": ["xxxxxxxxxxxxx"],
    "softwareName": "MEFEdge"
}
```

Response Example:

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

Response Status Code: 200

**Output Description<a name="section851619548567"></a>**

**Table 2** Operation output description

|Parameter|Type|Description|
|--|--|--|
|status|String|Error code|
|msg|String|Description|
|data|Object|-|

**Table 3** Data field description

|Parameter|Type|Description|
|--|--|--|
|successIDs|String list|List of successful serial numbers|
|failedInfos|Hash table, with both key and value types being String|The key is the failed serial number, and the value is the reason for the failure.|

## Error Code Description<a name="ZH-CN_TOPIC_0000001526721244"></a>

### Value Range of Error Codes<a name="ZH-CN_TOPIC_0000001582102534"></a>

**Table 1**  Error code range description

|Error Code Range|Description|
|--|--|
|0000 0000|Common part|
|4000 0000|edge-manager module<li>4001 0000: Node management API</li><li>4002 0000: Containerized application management API</li>|
|5000 0000|alarm-manager module|
|6000 0000|cert-manager module|

### Common Error Codes<a name="ZH-CN_TOPIC_0000001577600997"></a>

**Table 1** Common error codes

|Error Code|Description|
|--|--|
|00000000|Request succeeded|
|00001001|Failed to parse request body|
|00001002|Failed to obtain request API result|
|00001003|Failed to send synchronization message to module|
|00001004|Request routing and forwarding failed|
|00001005|Parameter validation failed|
|00001006|Request structure conversion failed|
|00001007|Request body type determination failed|
|00001008|Failed to create request message|

### Node APIs<a name="ZH-CN_TOPIC_0000001527201156"></a>

**Table 1** Node management error codes

|Error Code|Description|
|--|--|
|00000000|Request succeeded|
|40011000|Node management specification validation failed|
|40012000|Data already exists|
|40012001|Failed to create node group API|
|40012002|Failed to obtain node group list|
|40012003|Failed to obtain node group details|
|40012004|Failed to modify node group information|
|40012005|Failed to obtain node group count|
|40012006|Failed to delete node group|
|40012007|Failed to obtain node details|
|40012008|Failed to modify node information|
|40012009|Failed to obtain node count by status|
|40012010|Failed to obtain managed node list|
|40012011|Failed to obtain unmanaged node list|
|40012012|Error adding node to node group|
|40012013|Error managing node|
|40012014|Failed to delete node|
|40012015|Failed to remove node from node group|
|40012016|Failed to send information to node|
|40012017|Failed to obtain token|
|40012018|Failed to obtain software version number on node when querying software information|

### Containerized Application APIs<a name="ZH-CN_TOPIC_0000001577401145"></a>

**Table 1**  Application management error codes

|Error Code|Description|
|--|--|
|00000000|Request succeeded|
|40021000|Application management specification verification failed|
|40021001|Parameter conversion failed|
|40021002|Failed to parse container parameters|
|40022000|Data already exists|
|40022001|Data does not exist|
|40022002|Failed to create application|
|40022003|Failed to query application details|
|40022004|Failed to query application list|
|40022005|Failed to deploy application|
|40022006|Failed to uninstall application|
|40022007|Failed to update application|
|40022008|Failed to delete application|
|40022009|Failed to query container instances by application|
|40022010|Failed to query container instances by node|
|40022011|Failed to query all instance list|
|40022012|Failed to query instance count under node group|

### Log Collection APIs<a name="ZH-CN_TOPIC_0000001640828986"></a>

**Table 1**  Log Collection API Error Code Description

| Error Code | Description |
|--|--|
| 40052002 | Log export service exception |
| 40051002 | Log export failed to obtain node information |

### Alarm Information APIs <a name="ZH-CN_TOPIC_0000001643483134"></a>

**Table 1**  Alarm API error codes

| Error Code | Description |
|--|--|
| 00000000 | Request successful |
| 50011001 | Failed to query MEF Center node alarms or events |
| 50011002 | Failed to query MEF Edge node alarm or event list |
| 50011003 | Failed to query alarms or events by node group |
| 50011004 | Failed to query alarms or events |
| 50011005 | Failed to query node group details |
| 50011006 | Failed to parse node group detail data |
| 50011007 | Failed to obtain alarm or event details |

### Configuration APIs <a name="ZH-CN_TOPIC_0000001527041140"></a>

**Table 1** Configuration API error codes

|Error Code|Description|
|--|--|
|00000000|Request succeeded|
|60001001|Failed to query the root certificate of the specified category|
|60001002|Failed to issue the root certificate|
|60001003|Failed to verify the root certificate content|
|60001004|Failed to save the root certificate|
|60001005|Failed to delete the root certificate|
|60001006|Failed to distribute the root certificate to the edge side|
|60001007|Failed to obtain the cloud-side secret|
|60001008|Failed to configure image download information|
|60001009|Failed to export the root certificate|
|60001010|Failed to obtain certificate information|
|60001011|Failed to import the revocation list|
|60001012|alarm-manager failed to obtain the imported certificate information|
|60002001|Failed to obtain the cloud-edge authentication token|
