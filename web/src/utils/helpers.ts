export function joinPath(...parts: string[]): string {
  let path = '/';

  parts.map((part: string) => {
    if (part.startsWith('/')) {
      path += part.slice(1);
    } else {
      path += part;
    }

    if (!path.endsWith('/')) {
      path += '/';
    }
  });

  return path;
}
