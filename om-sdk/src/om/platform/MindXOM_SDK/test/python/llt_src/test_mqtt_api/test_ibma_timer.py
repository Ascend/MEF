import mock
from bin.ibma_timer import IbmaTimer
from test_mqtt_api.get_log_info import GetLogInfo

getLog = GetLogInfo()


class TestCapability:
    OK = True
    FAILED = False

    def test_judge_effective_should_failed_when_timer_is_interrupted(self):
        timer = IbmaTimer()
        timer.interrupt = True
        ret = timer.judge_effective()
        assert ret == self.FAILED

    def test_judge_effective_should_failed_when_timer_is_disabled(self):
        timer = IbmaTimer()
        timer.interrupt = False
        timer.enabled = False
        ret = timer.judge_effective()
        assert ret == self.FAILED

    def test_judge_effective_should_failed_when_timer_interval_is_zero(self):
        timer = IbmaTimer()
        timer.interrupt = False
        timer.enabled = True
        timer.minIntervalTime = 0
        ret = timer.judge_effective()
        assert ret == self.FAILED

    def test_judge_effective_should_failed_when_timer_interval_is_beyond_max(self):
        timer = IbmaTimer()
        timer.interrupt = False
        timer.enabled = True
        timer.minIntervalTime = 5
        timer.maxIntervalTime = 86401
        ret = timer.judge_effective()
        assert ret == self.FAILED

    def test_judge_effective_should_failed_when_interval_not_between_min_and_max(self):
        timer = IbmaTimer()
        timer.interrupt = False
        timer.enabled = True
        timer.minIntervalTime = 5
        timer.maxIntervalTime = 864
        timer.intervalTime = 3
        ret = timer.judge_effective()
        assert ret == self.FAILED

    def test_judge_effective_should_failed_when_timer_is_stopped(self):
        timer = IbmaTimer()
        timer.interrupt = False
        timer.enabled = True
        timer.minIntervalTime = 5
        timer.maxIntervalTime = 864
        timer.intervalTime = 10
        timer.runTimes = -1
        ret = timer.judge_effective()
        assert ret == self.FAILED

    def test_judge_effective_should_ok(self):
        timer = IbmaTimer()
        timer.interrupt = False
        timer.enabled = True
        timer.minIntervalTime = 5
        timer.maxIntervalTime = 864
        timer.intervalTime = 10
        timer.runTimes = 1
        ret = timer.judge_effective()
        assert ret == self.OK

    @mock.patch.object(IbmaTimer, "judge_effective", mock.Mock(return_value=False))
    @getLog.clear_common_log
    def test_start_timer_should_fail_when_check_failed(self):
        IbmaTimer()._start_timer()
        ret = getLog.get_log()
        assert "Timer: start failed." in ret
