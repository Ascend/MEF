import { describe, expect, test } from 'vitest';
import timeZone from '@/utils/timeZone';
import timeZoneMapper from '@/utils/timeZoneMapper';

describe('test timeZone.js suite', () => {
  test('ok', () => {
    const expected = 'zh';
    expect(timeZone).toHaveProperty(expected);
  })
})

describe('test timeZoneMapper.js suite', () => {
  test('ok', () => {
    const expected = 'AFRICA_ABIDJAN';
    expect(timeZoneMapper).toHaveProperty(expected);
  })
})