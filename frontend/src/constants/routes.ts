export const routePaths = {
  home: '/',
  about: '/about',
  blog: '/blog',
  login: '/login',
  register: '/register',
  profile: '/profile',
  profileInterfaces: '/profile/interfaces',
  profileGatewayLogs: '/profile/gateway-logs',
  profileChatRecords: '/profile/chat-records'
} as const;

export const adminRoutePaths = {
  root: '/admin',
  home: '/admin/home',
  systemUsers: '/admin/system/users',
  systemFiles: '/admin/system/files',
  systemConfig: '/admin/system/config',
  systemRole: '/admin/system/role',
  systemAudit: '/admin/system/audit'
} as const;

const protectedRoutes = new Set<string>([
  routePaths.home,
  routePaths.profile,
  routePaths.profileInterfaces,
  routePaths.profileGatewayLogs,
  routePaths.profileChatRecords,
  adminRoutePaths.root,
  adminRoutePaths.home,
  adminRoutePaths.systemUsers,
  adminRoutePaths.systemFiles,
  adminRoutePaths.systemConfig,
  adminRoutePaths.systemRole,
  adminRoutePaths.systemAudit
]);

const adminRoutes = new Set<string>([
  adminRoutePaths.root,
  adminRoutePaths.home,
  adminRoutePaths.systemUsers,
  adminRoutePaths.systemFiles,
  adminRoutePaths.systemConfig,
  adminRoutePaths.systemRole,
  adminRoutePaths.systemAudit
]);

export function isProtectedRoute(pathname: string): boolean {
  if (pathname.startsWith('/profile')) {
    return true;
  }
  return protectedRoutes.has(pathname);
}

export function isAdminRoute(pathname: string): boolean {
  return adminRoutes.has(pathname);
}
