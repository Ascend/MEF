import { expect, it, vi } from 'vitest';
import { showLoading } from '@/utils/axios/loading';
import * as elementPlus from 'element-plus';

vi.mock('element-plus', () => ({
  ElLoading: {
    service: vi.fn(),
  },
}));

it('should test showLoading', () => {
  const spy = vi.spyOn(elementPlus.ElLoading, 'service');
  showLoading();
  expect(spy).toHaveBeenCalledWith({
    lock: true,
    background: 'rgba(0, 0, 0, 0.6)',
  });

  showLoading();
  expect(spy).toHaveBeenCalledTimes(1);
});
