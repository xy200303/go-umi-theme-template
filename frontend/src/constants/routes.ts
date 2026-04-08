export const routePaths = {
  home: '/',
  about: '/about',
  blog: '/blog',
  login: '/login',
  register: '/register',
  profile: '/profile'
} as const;

export const adminRoutePaths = {
  root: '/admin',
  home: '/admin/home',
  systemUsers: '/admin/system/users',
  systemConfig: '/admin/system/config',
  systemRole: '/admin/system/role'
} as const;

const protectedRoutes = new Set<string>([
  routePaths.home,
  routePaths.profile,
  adminRoutePaths.root,
  adminRoutePaths.home,
  adminRoutePaths.systemUsers,
  adminRoutePaths.systemConfig,
  adminRoutePaths.systemRole
]);

const adminRoutes = new Set<string>([
  adminRoutePaths.root,
  adminRoutePaths.home,
  adminRoutePaths.systemUsers,
  adminRoutePaths.systemConfig,
  adminRoutePaths.systemRole
]);

export function isProtectedRoute(pathname: string): boolean {
  return protectedRoutes.has(pathname);
}

export function isAdminRoute(pathname: string): boolean {
  return adminRoutes.has(pathname);
}
