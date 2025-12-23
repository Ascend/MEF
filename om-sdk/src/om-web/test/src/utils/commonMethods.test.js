import { describe, expect, it, test, vi } from 'vitest';
import * as commonMethodsModule from '@/utils/commonMethods';
import {
  checkModuleType,
  clearSessionAndJumpToLogin,
  convertToGB,
  dateFormat,
  deepCopyDictList,
  getLanguageDefaultChinese,
  handleOperationResponseError,
  handleUploadError,
  isValidUploadFilename,
  switchToTargetTimezone
} from '@/utils/commonMethods';


describe('test convertToGB suite', () => {
  test('ok', () => {
    const bytes = 1024 * 1024 * 1024;
    const expected = '1.00 GB';
    const ret = convertToGB(bytes);
    expect(ret).toEqual(expected);
  });
});

describe('test dateFormat suite', () => {
  test('test yyyyMMddhhmmss format', () => {
    const expected = '20230708111845';
    const date = new Date(2023, 6, 8, 11, 18, 45);
    const format = 'yyyyMMddhhmmss';
    const ret = dateFormat(date, format);
    expect(ret).toEqual(expected);
  });
  test('test yyyy-MM-dd hh:mm:ss format', () => {
    const expected = '2023-07-08 11:18:45';
    const date = new Date(2023, 6, 8, 11, 18, 45);
    const format = 'yyyy-MM-dd hh:mm:ss';
    const ret = dateFormat(date, format);
    expect(ret).toEqual(expected);
  });
});

describe('test getLanguageDefaultChinese suite', () => {
  test('ok', () => {
    const expected = 'zh';
    const ret = getLanguageDefaultChinese();
    expect(ret).toEqual(expected);
  });
});

describe('test clearSessionAndJumpToLogin', () => {
  test('ok', () => {
    const router = {
      push: vi.fn(),
    };
    clearSessionAndJumpToLogin(router);
    expect(router.push).toHaveBeenCalledWith('/login');
  });
});

describe('test deepCopyDictList suite', () => {
  test('ok', () => {
    const src = [
      {
        'enable': 'false',
        'start_time': null,
        'end_time': null,
        'ip_addr': '123456',
        'mac_addr': null,
      },
    ];
    const expected = [
      {
        'enable': 'false',
        'start_time': null,
        'end_time': null,
        'ip_addr': '123456',
        'mac_addr': null,
      },
    ];
    const ret = deepCopyDictList(src);
    expect(ret).toEqual(expected);
  });
});

describe('test switchToTargetTimezone suite', () => {
  test('ok', () => {
    const currDate = new Date(2023, 6, 8, 10);
    const currTz = 'Etc/GMT-9';
    const targetTz = 'Etc/GMT-8';
    const expected = new Date(2023, 6, 8, 9);
    const ret = switchToTargetTimezone(currDate, currTz, targetTz);
    expect(ret).toEqual(expected);
  });
});

describe('test handleUploadError', () => {
  it('test in errorTipMapper', () => {
    const handleUploadErrorMock = vi.spyOn(commonMethodsModule, 'handleUploadError');

    const error = {
      message: JSON.stringify({
        error: {
          '@Message.ExtendedInfo': [
            {
              Oem: {
                status: 100025,
              },
            },
          ],
        },
      }),
    };

    handleUploadError(error);
    expect(handleUploadErrorMock).toHaveBeenCalled();
  });

  it('test not in errorTipMapper', () => {
    const handleUploadErrorMock = vi.spyOn(commonMethodsModule, 'handleUploadError');

    const error = {
      message: JSON.stringify({
        error: {
          '@Message.ExtendedInfo': [
            {
              Oem: {
                status: 'someStatus',
              },
            },
          ],
        },
      }),
    };

    handleUploadError(error);
    expect(handleUploadErrorMock).toHaveBeenCalled();
  });

  it('test uploadError', () => {
    const handleUploadErrorMock = vi.spyOn(commonMethodsModule, 'handleUploadError');

    const error = {
      message: JSON.stringify({
        error: {},
      }),
    };

    handleUploadError(error);
    expect(handleUploadErrorMock).toHaveBeenCalled();
  });
});

describe('test isValidUploadFilename suite', () => {
  test('Correct file name', () => {
    const fileName = 'WeakPas_1686055117342.conf';
    const expected = true;
    const ret = isValidUploadFilename(fileName);
    expect(ret).toEqual(expected);
  });

  it('Wrong file name', () => {
    const fileName = 'WeakPas 1686055117342.conf';
    const expected = null;
    const ret = isValidUploadFilename(fileName);
    expect(ret).toEqual(expected);
  });
});

describe('test checkModuleType suite', () => {
  test('ok', () => {
    const Module = 'MCU';
    const expected = true;
    const ret = checkModuleType(Module, 'mcu');
    expect(ret).toEqual(expected);
  });
});

describe('test handleOperationResponseError', () => {
  it('test in errorTipMapper', () => {
    const handleOperationResponseErrorMock = vi.spyOn(commonMethodsModule, 'handleOperationResponseError');

    const err = {
      response: {
        data: {
          error: {
            '@Message.ExtendedInfo': [
              {
                Oem: {
                  status: 'someStatus',
                },
              },
            ],
          },
        },
      },
    };

    handleOperationResponseError(err);
    expect(handleOperationResponseErrorMock).toHaveBeenCalled();
  });

  it('test not in errorTipMapper', () => {
    const handleOperationResponseErrorMock = vi.spyOn(commonMethodsModule, 'handleOperationResponseError');

    const err = {
      response: {
        data: {
          error: {
            '@Message.ExtendedInfo': [
              {
                Oem: {
                  status: 100025,
                },
              },
            ],
          },
        },
      },
    };

    handleOperationResponseError(err);
    expect(handleOperationResponseErrorMock).toHaveBeenCalled();
  });
});