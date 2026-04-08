import { useEffect, useState } from 'react';
import { getAuthOptions } from '@/api/endpoints/auth';
import { getClientEnv } from '@/lib/env';

const fallbackSmsVerifyEnabled = (getClientEnv('SMS_VERIFY_ENABLED') ?? 'true') === 'true';

export function useSmsVerifyEnabled() {
  const [smsVerifyEnabled, setSmsVerifyEnabled] = useState(fallbackSmsVerifyEnabled);

  useEffect(() => {
    let active = true;

    getAuthOptions()
      .then((options) => {
        if (!active) {
          return;
        }
        setSmsVerifyEnabled(options.sms_verify_enabled);
      })
      .catch(() => {
        if (!active) {
          return;
        }
        setSmsVerifyEnabled(fallbackSmsVerifyEnabled);
      });

    return () => {
      active = false;
    };
  }, []);

  return smsVerifyEnabled;
}
