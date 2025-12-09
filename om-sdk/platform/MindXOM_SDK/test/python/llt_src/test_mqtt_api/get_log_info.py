import os


class GetLogInfo:
    file = "/var/plog/manager/manager_run.log"

    def clear_common_log(self, fun):
        def decorator(*args, **kw):
            if os.path.exists(self.file):
                with open(self.file, mode='w') as file2:
                    file2.write('')
            return fun(*args, **kw)

        return decorator

    def clear_log(self):
        if os.path.exists(self.file):
            with open(self.file, mode='w') as file2:
                file2.write('')

    def get_log(self):
        res = ""
        if os.path.exists(self.file):
            with open(self.file) as file2:
                res = file2.read()
        return res


class GetOperationLog(GetLogInfo):
    file = "/var/plog/manager/manager_operate.log"
