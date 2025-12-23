import { describe, expect, test } from 'vitest';
import Error from '@/utils/axios/error';

describe('test error.js suite', () => {
  test('ok', () => {
    const err = {
      response: {
        status: '400',
        data: {
          status: '400',
        },
      },
    };
    const expected = {
      'code': '400',
      'message': '连接出错,状态码：(400)!',
      'response': {
        'data': {
          'status': '400',
        },
        'status': '400',
      },
      'status': '400',
    };
    const ret = Error(err);
    expect(ret).toEqual(expected);
  });
});