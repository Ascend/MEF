# Alarm Handling<a name="ZH-CN_TOPIC_0000001674415906"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:19:52.122Z pushedAt=2026-06-09T01:46:35.730Z -->

## Alarm and Event Definitions<a name="ZH-CN_TOPIC_0000001674415930"></a>

- Alarm: An issue that begins at a specific point in time.
- Event: A specific occurrence that happened at a past point in time.

<table><thead align="left"><tr id="row5323212152618"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="p15323121210269"><a name="p15323121210269"></a><a name="p15323121210269"></a>Alarm Type</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="p1732341222611"><a name="p1732341222611"></a><a name="p1732341222611"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="p9323151242611"><a name="p9323151242611"></a><a name="p9323151242611"></a>Reference Link</p>
</th>
</tr>
</thead>
<tbody><tr id="row1732321220263"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p1532321212269"><a name="p1532321212269"></a><a name="p1532321212269"></a>Cloud-edge collaboration alarm event</p>
<p id="p7323151218262"><a name="p7323151218262"></a><a name="p7323151218262"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p33231412192613"><a name="p33231412192613"></a><a name="p33231412192613"></a>Common cloud-edge collaboration alarm event.</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p4323191213267"><a name="p4323191213267"></a><a name="p4323191213267"></a><a href="#common-alarms-and-events">Common Alarms and Events</a></p>
</td>
</tr>
<tr id="row7323131282612"><td class="cellrowborder" valign="top" headers="mcps1.1.4.1.1 "><p id="p183231712122613"><a name="p183231712122613"></a><a name="p183231712122613"></a><span id="ph133231212142610"><a name="ph133231212142610"></a><a name="ph133231212142610"></a>MEF Center</span> cloud-edge collaboration alarm event.</p>
<p id="p232317121261"><a name="p232317121261"></a><a name="p232317121261"></a>Alarms or events that occur when <span id="ph1232331218265"><a name="ph1232331218265"></a><a name="ph1232331218265"></a>MEF Edge</span> connects to <span id="ph1732311126262"><a name="ph1732311126262"></a><a name="ph1732311126262"></a>MEF Center</span> for cloud-edge collaboration.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.1.4.1.2 "><p id="p16323151222617"><a name="p16323151222617"></a><a name="p16323151222617"></a><a href="#mef-center-cloud-edge-collaboration-alarms-and-events">MEF Center cloud-edge collaboration alarm event</a></p>
</td>
</tr>
<tr id="row123241412142615"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p1832361211261"><a name="p1832361211261"></a><a name="p1832361211261"></a><span id="ph203231912162615"><a name="ph203231912162615"></a><a name="ph203231912162615"></a>MEF Center</span> device alarm event</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p820825516497"><a name="p820825516497"></a><a name="p820825516497"></a>Alarm events generated when <span id="ph120819554495"><a name="ph120819554495"></a><a name="ph120819554495"></a>MEF Center</span> devices connect to third-party management platforms, software repositories, or image repositories.</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p16324101242611"><a name="p16324101242611"></a><a name="p16324101242611"></a><a href="#mef-center-device-alarms-and-events">MEF Center Device Alarms and Events</a></p>
</td>
</tr>
</tbody>
</table>

## Common Alarms and Events<a id="common-alarms-and-events"></a>

### 0x00131001 Docker Engine Exception (Critical Alarm)<a name="ZH-CN_TOPIC_0000001722295449"></a>

**Alarm Description<a name="zh-cn_topic_0190357109_section10542121642"></a>**

Alarm description: The Docker engine is abnormal.

This alarm is generated when the Docker engine is not running properly. The alarm is cleared when the Docker engine returns to normal operation.

The module generating this alarm is MEF Edge.

**Alarm Attribute<a name="zh-cn_topic_0190357109_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x00131001|Critical|Yes|

**Impact on the system<a name="zh-cn_topic_0190357109_section85431015414"></a>**

Containers cannot be deployed or run.

**Possible cause<a name="zh-cn_topic_0190357109_section95431611040"></a>**

The Docker engine is not running properly.

**Handling procedure<a name="section1420215111613"></a>**

1. Log in to the device CLI, run the **systemctl restart docker** command to restart the Docker engine, and check whether the alarm is cleared.
2. If the alarm cannot be cleared, contact Huawei technical support.

### 0x00131011  MEF Edge Log Space Full (Minor Alarm) <a name="ZH-CN_TOPIC_0000001674415902"></a>

**Alarm Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Alarm description: The MEF Edge log space is about to be full.

This alarm is generated when the occupied space of the MEF Edge log and log dump file directory reaches 80% or above; the alarm is cleared when the occupied space falls below this threshold.

The module generating this alarm is MEF Edge.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x00131011|Minor|Yes|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

MEF Edge cannot record logs; edgecore cannot start properly.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

Insufficient storage space in the MEF Edge log directory.

**Handling procedure<a name="section1817111521104"></a>**

1. Check whether the MEF Edge log directory has sufficient storage space.
    - Run the **df** **-lh** command to check the usage of each partition. An alarm is reported when the partition usage exceeds 80%.

        For example, query the size of each file in the /var/alog directory.

        The following information is displayed: The usage of the /var/alog directory is 92%, and an alarm is generated.

        ```text
        Euler:/var/alog # df -lh
        Filesystem      Size  Used Avail Use% Mounted on
        /dev/root       5.9G  1.8G  3.8G  32% /
        devtmpfs        5.5G     0  5.5G   0% /dev
        tmpfs           5.7G     0  5.7G   0% /dev/shm
        tmpfs           5.7G  650M  5.1G  12% /run
        tmpfs           4.0M     0  4.0M   0% /sys/fs/cgroup
        tmpfs           5.7G   10M  5.7G   1% /tmp
        tmpfs           128M     0  128M   0% /var/IEF
        tmpfs           128M  118M   11M  92% /var/alog
        tmpfs           128M     0  128M   0% /var/dlog
        tmpfs           128M  128M     0 100% /var/log
        tmpfs           128M  320K  128M   1% /var/plog
        /dev/mmcblk0p5  974M   55M  853M   6% /home/log
        /dev/mmcblk0p4  974M  332M  576M  37% /home/data
        /dev/mmcblk0p7  2.9G  286M  2.5G  11% /home/package
        /dev/mmcblk0p6  2.0G  175M  1.7G  10% /usr/local/mindx
        ```

    - Run the <b>du -sh</b> _directory_<b>/*</b> command to check the size of each file in the directory where the partition usage exceeds 80%.

        For example, run **du -sh /var/alog/\*** to query the size of each file in the /var/alog directory.

        The displayed information is as follows:

        ```text
        Euler:~ # du -sh /var/alog/*
        90M /var/alog/big_file
        28M /var/alog/MEFEdge_log
        ```

2. After backing up data, handle the files in the directory with insufficient space to free up space.
    - Run the **mv** command to move the files in the corresponding directory.
    - Run the **rm** command to delete the files in the corresponding directory.

3. If the preceding operations do not clear the alarm, contact Huawei technical support.

### 0x00131012 NPU Exception (Critical Alarm)<a name="ZH-CN_TOPIC_0000001722375465"></a>

**Alarm Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Alarm description: The NPU chip is in an unhealthy state.

This alarm is generated when the NPU chip health status is abnormal. The alarm is cleared when the NPU chip recovers to a healthy state.

Module generating this alarm: MEF Edge.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x00131012|Critical|Yes|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

The inference container cannot be deployed or run.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

The NPU chip is in an unhealthy state.

**Handling procedure<a name="section1817111521104"></a>**

1. Log in to the device and use the **npu-smi info** command to check the unhealthy NPU information.
2. Contact Huawei technical support and provide the relevant information.

### 0x00131004 Containerized Application Restart Event (Normal Event) <a name="ZH-CN_TOPIC_0000001722375545"></a>

**Event Description<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section14304185863417"></a>**

Event description: The containerized application status is restarted.

This event is generated when MEF Edge detects that a containerized application deployed through the network management system is restarted.

Module generating this event: MEF Edge.

**Event attribute<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section123741058103416"></a>**

**Table 1**  Event information

|Event ID|Event level|Auto-cleared|
|--|--|--|
|0x00131004|Normal|No|

**Impact on the system<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section17423131514189"></a>**

This may cause abnormal status of system containerized applications.

**Possible cause<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section13451932141819"></a>**

The containerized application exited abnormally or a configuration change caused the container to restart.

**Handling procedure<a name="section18190182917212"></a>**

1. Log in to the device and troubleshoot the cause of the container restart.
    - If yes, the handling is complete.
    - No, go to [2](#zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341).

2. <a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341"></a>Contact Huawei technical support.

### 0x00131014 Database Exception (Major Alarm)<a name="ZH-CN_TOPIC_0000001722295521"></a>

**Alarm Description<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section14304185863417"></a>**

Alarm description: MEF Edge database exception.

This alarm is generated when MEF Edge detects an abnormal database format.

Module generating this alarm: MEF Edge.

**Alarm Attribute<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section123741058103416"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x00131014|Major|Yes|

**Impact on the System<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section17423131514189"></a>**

The edge system will stop running or run abnormally.

**Possible Cause<a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_section13451932141819"></a>**

The MEF Edge database format is incorrect.

**Handling Procedure<a name="zh-cn_topic_0000001182494244_section6849845132315"></a>**

1. Log in to the device and check whether the database format is normal.
    - If yes, the handling is complete.
    - If no, go to [2](#zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341).

2. <a name="zh-cn_topic_0000001182494244_zh-cn_topic_0176114047_zh-cn_topic_0176881261_li1380111588341"></a>Contact Huawei technical support.

## MEF Center Cloud-Edge Collaboration Alarms and Events<a id="mef-center-cloud-edge-collaboration-alarms-and-events"></a>

### 0x00131013 MEF Center Root Certificate Expired (Critical Alarm)<a name="ZH-CN_TOPIC_0000001674256230"></a>

**Alarm Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Alarm description: The MEF Center root certificate (used to issue the MEF Center service certificate) in the MEF Edge device is about to expire or has expired.

This alarm is generated when the MEF Center root certificate in the MEF Edge device is less than the certificate expiration alarm time threshold. After the certificate is updated to a valid certificate, this alarm disappears.

Module generating this alarm: MEF Edge.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x00131013|Major|Yes|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

The root certificate of the MEF Center may expire.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

The root certificate of MEF Center is about to expire or has expired.

**Handling procedure<a name="section1817111521104"></a>**

1. Obtain the MEF Center root certificate and cloud-edge authentication token again, and configure network management on the MEF Edge device. For details, see [Authentication and Integration Between MEF Center and MEF Edge](./usage.md#ZH-CN_TOPIC_0000001722295385).
2. If the alarm is not auto-cleared, contact Huawei technical support and provide the relevant information.

### 0x01000004 Root Certificate for Issuing MEF Edge Service Certificate About to Expire (Minor Alarm) <a name="ZH-CN_TOPIC_0000001722375589"></a>

Alarm description: The root certificate used to issue MEF Edge service certificates in the MEF Center device is about to expire or has expired, generating an alarm.

- When the root certificate in the MEF Center device is about to expire (24 hours < certificate expiration time < 15 days), this alarm is generated. The MEF Edge device reapplies for the issuance of a service certificate, and the alarm disappears after the update process ends.
- When the root certificate in the MEF Center device has expired or the certificate expiration time is less than or equal to 24 hours, the certificate in the MEF Center device is forcibly updated, and the alarm disappears.

Module generating this alarm: MEF Center.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x01000004|Minor|Yes|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

MEF Edge and MEF Center may fail to connect properly.

**Possible cause**<a name="zh-cn_topic_0176114050_section95431611040"></a>

The root certificate used to issue the MEF Edge service certificate is about to expire or has expired.

**Handling procedure**<a name="section1817111521104"></a>

1. The MEF Center certificate management module detects that the root certificate needs to be updated and successfully triggers the update process.
2. If the alarm is not auto-cleared, contact Huawei technical support.

### 0x01000005 MEF Center Root Certificate About to Expire (Minor Alarm)<a name="ZH-CN_TOPIC_0000001722375477"></a>

**Alarm Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Alarm description: The MEF Center root certificate (used to issue MEF Center service certificates) on the MEF Center device is about to expire or has expired, generating an alarm.

- When the MEF Center root certificate on the MEF Center device is about to expire (24 hours < certificate expiration time < 15 days), this alarm is generated. After the MEF Edge device completes the MEF Center root certificate update, the alarm disappears when the update process ends.
- When the MEF Center root certificate on the MEF Center device has expired or the certificate expiration time is less than or equal to 24 hours, the certificate on the MEF Center device is forcibly updated, and the alarm disappears.

The module generating this alarm is MEF Center.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1** Alarm information

|Alarm ID|Alarm Level|Auto-cleared|
|--|--|--|
|0x01000005|Minor|Yes|

**Impact on the System<a name="zh-cn_topic_0176114050_section85431015414"></a>**

MEF Edge and MEF Center may fail to connect properly.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

The root certificate of MEF Center (used to issue the MEF Center service certificate) is about to expire or has expired.

**Handling procedure<a name="section1817111521104"></a>**

1. The MEF Center certificate management module detects that the root certificate needs to be updated and successfully triggers the update process.
2. If the alarm is not auto-cleared, contact Huawei technical support.

### 0x01000006 Abnormal Automatic Update Process of the Root Certificate to Issue the MEF Edge Service Certificate (Major Event)<a name="ZH-CN_TOPIC_0000001722375473"></a>

**Event Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Event description: After the root certificate used for issuing MEF Edge service certificates in the MEF Center device is updated, there are still MEF Edge devices that have not successfully applied for the issuance of corresponding service certificates.

Module generating this event: MEF Center.

**Event attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Event information

|Event ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x01000006|Major|No|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

The MEF Edge device that fails to update the service certificate cannot connect to MEF Center again.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

The MEF Edge failed to complete the application for issuing the MEF Edge service certificate due to network interruption or other reasons.

**Handling procedure<a name="section1817111521104"></a>**

1. Refer to [MEF Edge Failed to Automatically Update MEF Center Root Certificate](./troubleshooting.md#ZH-CN_TOPIC_0000001722295437), obtain the MEF Center root certificate and cloud-edge authentication token again, and configure the network management on the MEF Edge device.
2. If the alarm is not auto-cleared, contact Huawei technical support and provide the corresponding information.

### 0x01000007  Abnormal Automatic Update of the MEF Center Root Certificate (Major Event)<a name="ZH-CN_TOPIC_0000001674415998"></a>

**Event Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Event description: After the MEF Center root certificate (used to issue MEF Center service certificates) in the MEF Center device is updated, there are still MEF Edge devices that have not successfully updated this MEF Center root certificate.

Module generating this event: MEF Center.

**Event Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1** Event information

|Event ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x01000007|Major|No|

**Impact on the System<a name="zh-cn_topic_0176114050_section85431015414"></a>**

The MEF Edge device that fails to update the MEF Center root certificate cannot connect to the MEF Center again.

**Possible Cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

MEF Edge failed to properly update the MEF Center root certificate (used for issuing the MEF Center service certificate) due to network interruption or other reasons.

**Handling procedure<a name="section1817111521104"></a>**

1. Refer to [MEF Edge Automatic Update of MEF Center Root Certificate Failure](./troubleshooting.md#ZH-CN_TOPIC_0000001722295437), re-obtain the MEF Center root certificate and cloud-edge authentication token, and configure network management on the MEF Edge device.
2. If the alarm is not auto-cleared, contact Huawei technical support and provide the relevant information.

## MEF Center Device Alarms and Events<a id="mef-center-device-alarms-and-events"></a>

### 0x01000001 Root Certificate of a Third-Party Management Platform Expired (Major Alarm)<a name="ZH-CN_TOPIC_0000001722295473"></a>

**Alarm Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Alarm description: The root certificate of the third-party management platform in the MEF Center device is about to expire or has expired.

When the certificate validity period is less than the certificate expiration alarm time threshold, this alarm is generated. After the certificate is updated to a valid certificate, this alarm is cleared.

Module generating this alarm: MEF Center.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x01000001|Major|Yes|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

After the certificate expires, the third-party management platform cannot access MEF Center.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

The certificate is about to expire or has already expired.

**Handling procedure<a name="section1817111521104"></a>**

1. Check whether the certificate has expired or is about to expire, and import a valid certificate again.
2. If the alarm is not auto-cleared, contact Huawei technical support and provide the corresponding information.

### 0x01000002 Root Certificate of a Software Repository Expired (Major Alarm)<a name="ZH-CN_TOPIC_0000001722375505"></a>

**Alarm Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Alarm description: The root certificate of the software repository in the MEF Center device is about to expire or has expired.

This alarm is generated when the certificate validity period is less than the certificate expiration alarm time threshold. The alarm disappears after the certificate is updated to a valid certificate.

Module generating this alarm: MEF Center.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x01000002|Major|Yes|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

After the certificate expires, MEF Center cannot access the software repository.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

The certificate is about to expire or has already expired.

**Handling procedure<a name="section1817111521104"></a>**

1. Check whether the certificate has expired or is about to expire, and re-import a valid certificate.
2. If the alarm is not auto-cleared, contact Huawei technical support and provide the relevant information.

### 0x01000003 Root Certificate of an Image Repository Expired (Major Alarm)<a name="ZH-CN_TOPIC_0000001674256222"></a>

**Alarm Description<a name="zh-cn_topic_0176114050_section10542121642"></a>**

Alarm description: The root certificate of the image repository in the MEF Center device is about to expire or has expired.

This alarm is generated when the certificate validity period is less than the certificate expiration alarm time threshold. After the certificate is updated to a valid certificate, this alarm is cleared.

Module generating this alarm: MEF Center.

**Alarm Attribute<a name="zh-cn_topic_0176114050_section1554219118413"></a>**

**Table 1**  Alarm information

|Alarm ID|Alarm Severity|Auto-cleared|
|--|--|--|
|0x01000003|Major|Yes|

**Impact on the system<a name="zh-cn_topic_0176114050_section85431015414"></a>**

After the certificate expires, MEF Center cannot access the image repository.

**Possible cause<a name="zh-cn_topic_0176114050_section95431611040"></a>**

The certificate is about to expire or has expired.

**Handling procedure<a name="section1817111521104"></a>**

1. Check whether the certificate has expired or is about to expire, and import a valid certificate again.
2. If the alarm is not auto-cleared, contact Huawei technical support and provide the corresponding information.
