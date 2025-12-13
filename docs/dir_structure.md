项目全量目录层级介绍如下：

    mind-edge				        # 项目根目录
    ├── build				        # 构建相关目录
    ├── common-utils				# 公共工具库
    │   ├── backuputils				# 备份工具
    │   ├── build					# 构建相关目录
    │   ├── cache					# 缓存管理
    │   ├── checker					# 校验工具
    │   ├── cmsverify				# CMS 验证工具
    │   ├── database				# 数据库操作工具
    │   ├── envutils				# 环境变量工具
    │   ├── fileutils				# 文件操作工具
    │   ├── httpsmgr				# HTTPS 管理工具
    │   ├── hwlog					# 日志工具
    │   ├── k8stool					# Kubernetes 工具
    │   ├── kmc					# KMC 工具
    │   ├── limiter					# 限流工具
    │   ├── logmgmt					# 日志管理
    │   ├── modulemgr				# 模块管理器
    │   ├── rand					# 随机数工具
    │   ├── terminal				# 终端工具
    │   ├── test					# 测试目录
    │   ├── tls					# TLS 工具
    │   ├── utils					# 通用工具
    │   ├── websocketmgr				# WebSocket 管理器
    │   ├── x509					# X.509 工具
    │   └── xcrypto					# 加密工具
    ├── device-plugin				# 设备插件组件
    │   ├── build					# 构建目录
    │   ├── doc				        # 文档
    │   └── pkg				        # 主程序代码
    ├── mef-center					# MEFCenter 中心组件代码
    │   ├── alarm-manager				# 告警管理
    │   ├── build					# 构建配置
    │   ├── cert-manager				# 证书管理
    │   ├── common					# 公共模块
    │   ├── edge-manager				# 边缘管理器
    │   ├── mef-center-install			# MEF 安装工具
    │   ├── nginx-manager				# Nginx 管理器
    │   ├── opensource				# 开源组件目录
    │   └── platform				# 平台模块目录
    ├── mef-edge					# MEFEdge 边缘组件代码
    │   ├── build					# 构建配置
    │   └── edge-installer				# 边缘组件目录
    │       ├── build				# 构建配置
    │       ├── cmd					# 主程序入口
    │       ├── config				# 配置目录
    │       ├── pkg					# 主程序代码
    │       ├── script				# 脚本目录
    │       └── tool				# 工具目录
    ├── om-sdk					# OM SDK 组件代码
    │   ├── application				# 应用程序
    │   ├── build					# 构建配置
    │   ├── config					# 配置目录
    │   ├── doc				        # 文档目录
    │   ├── output					# 输出目录
    │   ├── platform				# 平台模块
    │   ├── src				        # 源码目录
    │   └── test				        # 测试目录
    └── om-web					# OM Web 前端组件代码
    │    ├── build					# 构建脚本
    │    ├── public					# 公共资源
    │    ├── src					# 源码目录
    │    └── test					# 测试目录