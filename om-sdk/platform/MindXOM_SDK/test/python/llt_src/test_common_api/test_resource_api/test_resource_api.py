import os
from collections import namedtuple

from common.ResourceDefV1.accounts import RfAccountServiceObj
from common.ResourceDefV1.errorcolection import RfErrorObj
from common.ResourceDefV1.resource import RfResource
from common.ResourceDefV1.session import RfSessionServiceObj
from common.ResourceDefV1.systems import RfSystemsCollection
from common.ResourceDefV1.task_service import RfTaskServiceObj
from common.utils.app_common_method import AppCommonMethod
from conftest import TestBase

project_dir = AppCommonMethod.get_project_absolute_path()
base_path = os.path.join(project_dir, "common/MockupData/iBMAServerV1")

ResourceCreateCase = namedtuple("ResourceCreateCase", "tested_class, tested_class_filed, "
                                                      "tested_class_filed_filed, excepted, rel_path")

ResourcePatchGetRfCase = namedtuple("ResourcePatchGetRfCase", "tested_class, excepted, rel_path")


class TestResourceObj(TestBase):
    use_cases = {
        "test_create_sub_objects": {
            "RfAccountServiceObj": (RfAccountServiceObj, "account_service_resource", "Name",
                                    "Account Service", "redfish/v1/AccountService"),
            "RfErrorObj": (RfErrorObj, "errorcolection", "ID", "Base.1.0", "redfish/v1/ErrorCollection"),
            "RfSessionServiceObj": (RfSessionServiceObj, "session_service_resource", "Name", "Session Service",
                                    "redfish/v1/SessionService"),
            "RfTaskServiceObj": (RfTaskServiceObj, "TasksSet", "Name", "Task Collection", "redfish/v1/TaskService"),
        },
        "test_get_rf_resource": {
            "RfResource": (RfResource, "Base.1.0.errorID", "redfish/v1/ErrorCollection"),
            "RfSystemsCollection": (RfSystemsCollection, "redfish/v1/Systems", "redfish/v1/Systems")
        },
    }

    def test_create_sub_objects(self, model: ResourceCreateCase):
        obj = model.tested_class(base_path, model.rel_path)
        obj.create_sub_objects(base_path, model.rel_path)
        assert model.excepted == getattr(obj, model.tested_class_filed).resData[model.tested_class_filed_filed]

    def test_get_rf_resource(self, model: ResourcePatchGetRfCase):
        ret = model.tested_class(base_path, model.rel_path).get_resource()
        assert model.excepted in ret
