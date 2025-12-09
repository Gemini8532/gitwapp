import { describe, it, expect } from 'vitest';
import { shortenPath } from './pathUtils';

describe('shortenPath', () => {
  it('should shorten paths starting with /home/user', () => {
    expect(shortenPath('/home/jules/projects/gitwapp')).toBe('~jules/projects/gitwapp');
  });

  it('should shorten path which is exactly /home/user', () => {
    expect(shortenPath('/home/jules')).toBe('~jules');
  });

  it('should shorten path /home/user/', () => {
    expect(shortenPath('/home/jules/')).toBe('~jules/');
  });

  it('should not change paths that do not start with /home/', () => {
    expect(shortenPath('/var/www/html')).toBe('/var/www/html');
    expect(shortenPath('relative/path')).toBe('relative/path');
  });

  it('should not change paths that are just /home', () => {
    expect(shortenPath('/home')).toBe('/home');
  });

  it('should not change paths that look like /home but are not directories under /home', () => {
      // Depending on implementation, but strict /home/ check
      expect(shortenPath('/home_dir')).toBe('/home_dir');
  });
});
