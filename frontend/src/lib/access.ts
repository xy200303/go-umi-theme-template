import { adminRoutePaths, routePaths } from '@/constants/routes';
import type { AuthUser } from '@/types/auth';

export type AccessKey = 'profile' | 'admin.dashboard' | 'admin.users' | 'admin.files' | 'admin.roles' | 'admin.configs' | 'admin.audits';

const accessOperationIds: Record<AccessKey, string[]> = {
  profile: ['profile.view'],
  'admin.dashboard': ['dashboard.stats.get'],
  'admin.users': ['users.list'],
  'admin.files': ['files.admin.list', 'files.admin.stats'],
  'admin.roles': ['roles.list'],
  'admin.configs': ['configs.list'],
  'admin.audits': ['audits.list']
};

const adminAccessEntries: Array<{ accessKey: AccessKey; path: string }> = [
  { accessKey: 'admin.dashboard', path: adminRoutePaths.home },
  { accessKey: 'admin.users', path: adminRoutePaths.systemUsers },
  { accessKey: 'admin.files', path: adminRoutePaths.systemFiles },
  { accessKey: 'admin.roles', path: adminRoutePaths.systemRole },
  { accessKey: 'admin.configs', path: adminRoutePaths.systemConfig },
  { accessKey: 'admin.audits', path: adminRoutePaths.systemAudit }
];

export function hasAccess(user: AuthUser | null | undefined, accessKey: AccessKey): boolean {
  if (!user) {
    return false;
  }

  if (user.roles.includes('admin')) {
    return true;
  }

  if (accessKey === 'profile' && !(user.operation_ids?.length)) {
    return true;
  }

  const grantedOperationIds = new Set(user.operation_ids ?? []);
  return accessOperationIds[accessKey].some((operationId) => grantedOperationIds.has(operationId));
}

export function getFirstAccessibleAdminPath(user: AuthUser | null | undefined): string | null {
  const matchedEntry = adminAccessEntries.find((entry) => hasAccess(user, entry.accessKey));
  return matchedEntry?.path ?? null;
}

export function hasAnyAdminAccess(user: AuthUser | null | undefined): boolean {
  return getFirstAccessibleAdminPath(user) !== null;
}

export function canAccessRoute(user: AuthUser | null | undefined, pathname: string): boolean {
  if (!user) {
    return false;
  }

  if (pathname === routePaths.home) {
    return true;
  }
  if (pathname === routePaths.profile) {
    return hasAccess(user, 'profile');
  }
  if (pathname === adminRoutePaths.root) {
    return hasAnyAdminAccess(user);
  }
  if (pathname === adminRoutePaths.home) {
    return hasAccess(user, 'admin.dashboard');
  }
  if (pathname === adminRoutePaths.systemUsers) {
    return hasAccess(user, 'admin.users');
  }
  if (pathname === adminRoutePaths.systemFiles) {
    return hasAccess(user, 'admin.files');
  }
  if (pathname === adminRoutePaths.systemRole) {
    return hasAccess(user, 'admin.roles');
  }
  if (pathname === adminRoutePaths.systemConfig) {
    return hasAccess(user, 'admin.configs');
  }
  if (pathname === adminRoutePaths.systemAudit) {
    return hasAccess(user, 'admin.audits');
  }

  return true;
}
