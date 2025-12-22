class EdgeSystemUrls:
    BASE_URL = "/redfish/v1/Systems/"
    ETHERNET_INTERFACES = BASE_URL + "EthernetInterfaces/"
    SIMPLE_STORAGES = BASE_URL + "SimpleStorages/"
    PARTITIONS = BASE_URL + "Partitions/"
    LTE = BASE_URL + "LTE/ConfigInfo/"
    NFS = BASE_URL + "NfsManage/"
    SECURITY_LOAD = BASE_URL + "SecurityService/SecurityLoad/"
    SECURITY_IMPORT = BASE_URL + "SecurityService/SecurityLoad/Actions/SecurityLoad.Import"
    SECURITY_SERVICE = BASE_URL + "SecurityService/"
    RESET = BASE_URL + "Actions/RestoreDefaults.Reset/"

    def __init__(self):
        pass
