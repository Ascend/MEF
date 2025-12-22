import { describe, expect, test } from 'vitest';
import {
  ipWithMaskChecker,
  macChecker, validateAccount, validateAssetTag, validateCertAlarmTime, validateHostName,
  validateIp,
  validateNodeId,
  validatePort,
  validateServerName, validateSessionTimeout, validatePasswordExpirationDays, validateNetworkUsage,
  validateVlanId, validateServerPath
} from '@/utils/validator';
import i18n from '@/utils/locale';

let callbackData;
const mockCallback = (data) => {
  callbackData = data;
};

describe('test macChecker suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '1A-2B-3C-4D-5E-6F';
    macChecker(rule, value, mockCallback);
    const expected = new Error('MAC地址格式错误');
    expect(callbackData).toStrictEqual(expected);
  });

  test('null test', () => {
    const rule = '';
    const value = '';
    macChecker(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('xx:xx:xx type test', () => {
    // MAC地址:支持两种格式，xx:xx:xx 和 xx:xx:xx:xx:xx:xx
    const rule = '';
    const value = '1A:2B:3C';
    macChecker(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('xx:xx:xx:xx:xx:xx type test', () => {
    // MAC地址:支持两种格式，xx:xx:xx 和 xx:xx:xx:xx:xx:xx
    const rule = '';
    const value = '1A:2B:3C:4D:5E:6F';
    macChecker(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test ipWithMaskChecker suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '256.168.1.1/8';
    ipWithMaskChecker(rule, value, mockCallback);
    const expected = new Error('IP地址格式错误');
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '192.168.1.1';
    ipWithMaskChecker(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    ipWithMaskChecker(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateIp suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '256.168.1';
    validateIp(rule, value, mockCallback);
    const expected = new Error('该字段格式错误');
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '192.168.1.1';
    validateIp(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateIp(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateNodeId suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '89abcdef-89ab-89ab-89ab-123456abcde';
    validateNodeId(rule, value, mockCallback);
    const expected = new Error('请输入合法的节点ID');
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '89abcdef-89ab-89ab-89ab-123456abcdef';
    validateNodeId(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateNodeId(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validatePort suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '123456';
    validatePort(rule, value, mockCallback);
    const expected = new Error('请输入合法的端口号');
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '12345';
    validatePort(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validatePort(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateServerName suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '89_abcdef.';
    validateServerName(rule, value, mockCallback);
    const expected = new Error('请输入合法的服务器名称');
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '89-abcdef.';
    validateServerName(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateServerName(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateAccount suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '89_abc.ABC';
    validateAccount(rule, value, mockCallback);
    const expected = new Error('最大长度256个字符，由大小写字母、数字、短横线(-)、下划线(_)组成');
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '89-abc_ABC';
    validateAccount(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateAccount(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateHostName suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '89_abc.ABC';
    validateHostName(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('information.systemInfo.hostnameErrTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = 'abc-123';
    validateHostName(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateHostName(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateAssetTag suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '中文';
    validateAssetTag(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('information.systemInfo.AssetTagErrTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = 'abc-123';
    validateAssetTag(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateAssetTag(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateSessionTimeout suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = 4;
    validateSessionTimeout(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('common.errorRangeTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = 6;
    validateSessionTimeout(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateSessionTimeout(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateCertAlarmTime suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = 4;
    validateCertAlarmTime(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('common.errorRangeTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = 8;
    validateCertAlarmTime(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateCertAlarmTime(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validatePasswordExpirationDays suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = 366;
    validatePasswordExpirationDays(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('common.errorRangeTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = 8;
    validatePasswordExpirationDays(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validatePasswordExpirationDays(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateNetworkUsage suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = 'validateNetworkUsagevalidateNetworkUsagevalidateNetworkUsagevalidateNetworkUsage';
    validateNetworkUsage(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('network.networkWired.ipUsageErrorTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = 'web';
    validateNetworkUsage(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateNetworkUsage(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateVlanId suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '4095';
    validateVlanId(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('network.networkWired.vlanIdErrorTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '50';
    validateVlanId(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateVlanId(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});

describe('test validateServerPath suite', () => {
  test('Failed Test', () => {
    const rule = '';
    const value = '/home/测试';
    validateServerPath(rule, value, mockCallback);
    const expected = new Error(i18n.global.t('disk.common.serverPathTip'));
    expect(callbackData).toStrictEqual(expected);
  });

  test('Success Test', () => {
    const rule = '';
    const value = '/home/data';
    validateServerPath(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });

  test('null Test', () => {
    const rule = '';
    const value = '';
    validateServerPath(rule, value, mockCallback);
    const expected = undefined;
    expect(callbackData).toStrictEqual(expected);
  });
});