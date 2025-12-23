// myModule.test.js
import { test, vi, expect } from 'vitest';
import { handleResponseError } from '@/utils/axios/handleResponseError';

const clearMock = vi.fn();
const getItemMock = vi.fn();
const setItemMock = vi.fn();
const alertMock = vi.fn().mockResolvedValue();

global.sessionStorage = {
  ...global.sessionStorage,
  clear: clearMock,
  getItem: getItemMock,
  setItem: setItemMock,
};
global.ElMessageBox = {
  ...global.ElMessageBox,
  alert: alertMock,
};
test('SessionTimeLimitExceeded', async () => {
  const response = {
    data: {
      error: {
        code: 'SessionTimeLimitExceeded',
        '@Message.ExtendedInfo': [
          {
            Oem: { status: 11020 },
            Message: '',
          },
        ],
      },
    },
  };

  await handleResponseError(response);

  expect(clearMock).toHaveBeenCalled();
});

test('AccountForSessionNoLongerExists', async () => {
  const response = {
    data: {
      error: {
        code: 'AccountForSessionNoLongerExists',
        '@Message.ExtendedInfo': [
          {
            Oem: { status: 11020 },
            Message: '',
          },
        ],
      },
    },
  };

  await handleResponseError(response);

  expect(clearMock).toHaveBeenCalled();
});

test('SESSION.ERROR_PASSWORD_EXPIRE', async () => {
  const response = {
    data: {
      error: {
        code: 'SessionTimeLimitExceeded',
        '@Message.ExtendedInfo': [
          {
            Oem: { status: 110204 },
            Message: '',
          },
        ],
      },
    },
  };

  await handleResponseError(response);

  expect(clearMock).toHaveBeenCalled();
});
