declare module '*.module.css' {
  const classes: Record<string, string>;
  export default classes;
}

declare module 'md5.js' {
  export default class MD5 {
    update(data: ArrayBufferView | ArrayBuffer | string, encoding?: string): MD5;
    digest(encoding?: 'hex'): string;
  }
}
