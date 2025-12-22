import { describe, expect, test, beforeEach, afterEach, vi } from 'vitest';
import { saveFile } from '@/utils/downloadFile';

describe('saveFile', () => {
  let createObjectURLMock;
  let revokeObjectURLMock;

  beforeEach(() => {
    createObjectURLMock = vi.fn();
    revokeObjectURLMock = vi.fn();

    global.window = {
      ...global.window,
      URL: {
        createObjectURL: createObjectURLMock,
        revokeObjectURL: revokeObjectURLMock,
      },
    };
  });

  afterEach(() => {
    delete global.window;
  });

  test('saves the file with the given data and file name', () => {
    const data = 'test data';
    const fileName = 'test.txt';

    saveFile(data, fileName);

    expect(createObjectURLMock).toHaveBeenCalled();
    expect(revokeObjectURLMock).toHaveBeenCalled();
  });
});

describe('saveFile', () => {
  test('saveFile', () => {
    const createElement = document.createElement;
    const createObjectURLMock = vi.fn();
    const revokeObjectURLMock = vi.fn();

    let anchor;
    document.createElement = tagName => {
      if (tagName === 'a') {
        anchor = createElement.call(document, tagName);
        return anchor;
      }
      return createElement.call(document, tagName);
    };
    global.window = {
      ...global.window,
      URL: {
        createObjectURL: createObjectURLMock,
        revokeObjectURL: revokeObjectURLMock,
      },
    };

    saveFile('data', 'file.txt');

    expect(createObjectURLMock).toHaveBeenCalled();
    expect(revokeObjectURLMock).toHaveBeenCalled();
    expect(anchor.download).toBe('file.txt');

    document.createElement = createElement;
  });
});
