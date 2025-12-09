import abc
from abc import ABC


class RestfulTestBase:
    tests_list = []
    labels_list = []

    def __init__(self, **kwargs):
        """
        1、测试基类，完成参数的获取以及和测试用例交互，所有的restful测试类都必须基于此基类
        2、该类派生出了确定请求方式的子类(PostTest、GetTest、DeleteTest)，请继承这些子类，用于开发
           例如：class NetManagerGetTestSuccess(GetTest)
        3、每个最终的业务子类都必须实现call_back_assert函数，该函数的作用是处理接口测试的返回消息(该函数内部必须存在assert判断语句)
           如果不需要任何操作实现：则可以直接实现为assert True
        4、如果需要在测试前打桩或者设置文件等操作，请子类重写before函数，在该函数内的操作，会在接口访问前进行，测试完成后需要
            恢复操作，则实现after函数，该函数会在接口访问后进行操作

        :param kwargs:
        url:为访问的路径，只需要填写api接口上route内的url即可，例如/NetManager
        method:为请求的方式，通过继承(PostTest、GetTest、DeleteTest)这些子类，method会自动设置好
        data:为请求的数据，GET请求，不需要data，其他方式请求需要data，类型为字典型的
        code: 为期望接口返回的状态码，int类型
        header: 为请求的消息头，默认会增加{"X-Real-IP": "127.0.0.1"}，用于token校验的IP地址校验
        label: 为测试用例名称
        """
        self._url: str = kwargs.get("url", "")
        self._method: str = kwargs.get("method", "GET")
        self._data: dict = kwargs.get("data")
        self._code: int = kwargs.get("code", 0)
        self._label: str = kwargs.get("label", "restful base test")  # 测试用例名称
        self._header: dict = kwargs.get("header", {})
        # 以下两条是收集有多少用例会进行测试，基类每被初始化一次，则会增加一次测试
        RestfulTestBase.tests_list.append(self)
        RestfulTestBase.labels_list.append(self._label)

    def __str__(self):
        return f"label={self._label}, url={self._url}, method={self._method}, data={self._data}, code={self._code}"

    def get_url(self) -> str:
        return self._url

    def get_method(self) -> str:
        return self._method

    def get_data(self) -> dict:
        return self._data

    def get_code(self) -> int:
        return self._code

    def get_header(self) -> dict:
        if "X-Real-IP" not in self._header:
            self._header.update({"X-Real-IP": "127.0.0.1"})

        return self._header

    def before(self):
        """测试前置条件"""
        pass

    def after(self):
        """测试后恢复操作等"""
        pass

    def get_call_back(self):
        return self.call_back_assert

    @abc.abstractmethod
    def call_back_assert(self, test_response: str):
        """返回消息回调处理，需要实现assert的操作判断"""
        pass


class PostTest(RestfulTestBase, ABC):
    """请求方式为POST的接口，请继承该类"""

    def __init__(self, url: str, code: int, data: dict, label: str, **kwargs):
        super().__init__(url=url, code=code, data=data, label=label, method="POST", **kwargs)


class GetTest(RestfulTestBase, ABC):
    """请求方式为GET的接口，请继承该类"""

    def __init__(self, url: str, code: int, label: str, **kwargs):
        super().__init__(url=url, code=code, label=label, method="GET", **kwargs)


class DeleteTest(RestfulTestBase, ABC):
    """请求方式为DELETE的接口，请继承该类"""

    def __init__(self, url: str, code: int, data: dict, label: str, **kwargs):
        super().__init__(url=url, code=code, data=data, label=label, method="DELETE", **kwargs)


class PatchTest(RestfulTestBase, ABC):
    """请求方式为PATCH的接口，请继承该类"""

    def __init__(self, url: str, code: int, data: dict, label: str, **kwargs):
        super().__init__(url=url, code=code, data=data, label=label, method="PATCH", **kwargs)