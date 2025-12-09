export const shortenPath = (path: string): string => {
  return path.replace(/^\/home\/([^/]+)/, '~$1');
};
