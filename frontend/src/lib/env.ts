type ClientEnvKey = 'API_BASE_URL' | 'SMS_VERIFY_ENABLED';

function readEnv(name: string): string | undefined {
  if (typeof process === 'undefined') {
    return undefined;
  }

  return process.env[name];
}

export function getClientEnv(key: ClientEnvKey): string | undefined {
  return readEnv(`UMI_APP_${key}`);
}

export function isDev(): boolean {
  return readEnv('NODE_ENV') === 'development';
}
