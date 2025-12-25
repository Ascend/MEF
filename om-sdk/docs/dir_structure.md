项目全量目录层级介绍如下：
```
om-sdk				                # 项目根目录
├── build				        # 构建相关目录
├── docs				        # 文档目录
│   └── images				        # 图片资源目录
└── src				                # 源码目录
    ├── om				        # OM SDK 组件代码
    │   ├── build	        		# 构建子目录
    │   │   └── web_assets			# Web资源目录
    │   ├── config		           	# 配置目录
    │   ├── doc				        # 文档目录
    │   ├── output	     		        # 输出目录
    │   ├── platform	 		        # 平台模块
    │   │   ├── cpp			        # C++平台代码
    │   │   └── MindXOM_SDK		        # MindXOM SDK子仓
    │   ├── src				        # 源码目录
    │   │   └── app			        # 应用程序目录
    │   └── test			        # 测试目录
    │       └── python			        # python测试目录
    └── om-web				        # OM Web 前端组件代码
        ├── build			        # 构建子目录
        ├── public			        # 公共资源
        │   ├── config			        # 配置文件
        │   ├── onlineHelp		        # 联机帮助
        │   └── WhiteboxConfig		        # 白盒配置
        ├── src				        # 源码目录
        │   ├── api			        # API接口
        │   ├── assets			        # 静态资源
        │   ├── components		        # 组件目录
        │   ├── router			        # 路由配置
        │   ├── utils			        # 工具函数
        │   └── views			        # 视图组件
        └── test			        # 测试目录
            ├── DT			        # DT测试
            ├── reports			        # 测试报告
            └── src			        # 测试源码
```