import { StrictMode, useEffect, type ReactNode } from 'react';
import { App as AntdApp, ConfigProvider } from 'antd';
import { I18nProvider } from '@/i18n';
import { registerMessageApi } from '@/lib/notify';
import { adminRoutePaths, isProtectedRoute, routePaths } from '@/constants/routes';
import { canAccessRoute, getFirstAccessibleAdminPath } from '@/lib/access';
import { useAuthStore } from '@/stores';

function AntdMessageRegistrar() {
  const { message } = AntdApp.useApp();

  useEffect(() => {
    registerMessageApi(message);
  }, [message]);

  return null;
}

function RootProviders({ children }: { children: ReactNode }) {
  return (
    <StrictMode>
      <ConfigProvider
        theme={{
          token: {
            colorPrimary: '#1d6df5',
            colorInfo: '#1d6df5',
            borderRadius: 12,
            colorText: '#14345f',
            colorTextSecondary: '#365a87',
            colorBgLayout: '#f3f8ff',
            colorBgContainer: '#ffffff'
          }
        }}
      >
        <I18nProvider>
          <AntdApp>
            <AntdMessageRegistrar />
            {children}
          </AntdApp>
        </I18nProvider>
      </ConfigProvider>
    </StrictMode>
  );
}

function normalizePathname(pathname: string): string {
  if (pathname === '/') return pathname;
  return pathname.replace(/\/+$/, '') || '/';
}

function redirectTo(pathname: string): void {
  if (typeof window === 'undefined') return;
  const current = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  if (current === pathname) return;
  window.location.replace(pathname);
}

export function rootContainer(container: ReactNode) {
  return <RootProviders>{container}</RootProviders>;
}

export function onRouteChange({ location }: { location: { pathname: string } }) {
  const pathname = normalizePathname(location.pathname);
  const { user } = useAuthStore.getState();

  if ((pathname === routePaths.login || pathname === routePaths.register) && user) {
    redirectTo(routePaths.home);
    return;
  }

  if (!isProtectedRoute(pathname)) {
    return;
  }

  if (!user) {
    redirectTo(routePaths.login);
    return;
  }

  if (pathname === adminRoutePaths.root) {
    redirectTo(getFirstAccessibleAdminPath(user) ?? routePaths.home);
    return;
  }

  if (!canAccessRoute(user, pathname)) {
    redirectTo(pathname.startsWith('/admin') ? (getFirstAccessibleAdminPath(user) ?? routePaths.home) : routePaths.home);
  }
}
