import re
from collections import namedtuple

from common.log.log_handlers import TrimSensitiveLoggerHandler
from conftest import TestBase

TrimSensitiveLoggerHandlerCase = namedtuple("TrimSensitiveLoggerHandler", "func, expected, input_param")


class TestHandler(TestBase):
    use_cases = {
        "test_trim_sensitive_logger_handler_func": {
            "_path_mask_less_3": (TrimSensitiveLoggerHandler._path_mask, "12", re.match("[0-9]*", "12")),
            "_path_mask_more_3": (TrimSensitiveLoggerHandler._path_mask, "***4", re.match("[0-9]*", "1234")),
            "_remove_abspath_str": (TrimSensitiveLoggerHandler._remove_abspath, "***34ab/c", "/1234ab/c"),
            "_remove_abspath_not_str": (TrimSensitiveLoggerHandler._remove_abspath, [1, 2], [1, 2]),
            "_remove_url_pword_str": (TrimSensitiveLoggerHandler._remove_url_pword,
                                       "https://:******@", "https://1234@"),
            "_remove_url_pword_not_str": (TrimSensitiveLoggerHandler._remove_url_pword, [1, 2], [1, 2]),
            "_remove_pword_str": (TrimSensitiveLoggerHandler._remove_pword, "pwd'******'", "password://1234@"),
            "_remove_pword_not_str": (TrimSensitiveLoggerHandler._remove_pword, [1, 2], [1, 2]),
            "_remove_pbkdf2_str": (TrimSensitiveLoggerHandler._remove_pbkdf2, "pbkdf2:******", "pbkdf2://1234@"),
            "_remove_pbkdf2_not_str": (TrimSensitiveLoggerHandler._remove_pbkdf2, [1, 2], [1, 2]),
            "_remove_crlf_str": (TrimSensitiveLoggerHandler._remove_crlf, "abcd*abcd*", "abcd\rabcd\n"),
            "_remove_crlf_not_str": (TrimSensitiveLoggerHandler._remove_crlf, [1, 2], [1, 2]),
            "handle_sensitive_log": (TrimSensitiveLoggerHandler().handle_sensitive_log,
                                     "pwd'******'", "password://1234@"),
        }
    }

    def test_trim_sensitive_logger_handler_func(self, model: TrimSensitiveLoggerHandlerCase):
        ret = model.func(model.input_param)
        assert model.expected == ret
