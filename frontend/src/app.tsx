import { StrictMode, useEffect, type ReactNode } from 'react';
import { App as AntdApp, ConfigProvider } from 'antd';
import { I18nProvider } from '@/i18n';
import { registerMessageApi } from '@/lib/notify';
import { adminRoutePaths, isAdminRoute, isProtectedRoute, routePaths } from '@/constants/routes';
import { useAuthStore } from '@/stores';
import 'antd/dist/reset.css';
import '@/styles/globals.css';

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
  const { user, isAdmin } = useAuthStore.getState();

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

  if (isAdminRoute(pathname) && !isAdmin()) {
    redirectTo(routePaths.home);
    return;
  }

  if (pathname === adminRoutePaths.root) {
    redirectTo(adminRoutePaths.home);
  }
}
