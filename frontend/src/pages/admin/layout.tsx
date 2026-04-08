import { useState, type ReactNode } from 'react';
import { DashboardOutlined, DatabaseOutlined, FileSearchOutlined, FolderOpenOutlined, SafetyCertificateOutlined, SettingOutlined, TeamOutlined } from '@ant-design/icons';
import { useNavigate } from '@/lib/router';
import MainNavbar from '@/components/layout/MainNavbar';
import SidebarLayout from '@/components/layout/SidebarLayout';
import { adminRoutePaths } from '@/constants/routes';
import { useI18n } from '@/i18n';
import { hasAccess } from '@/lib/access';
import { useAuthStore } from '@/stores';

const adminSectionPaths = {
  dashboard: adminRoutePaths.home,
  users: adminRoutePaths.systemUsers,
  files: adminRoutePaths.systemFiles,
  roles: adminRoutePaths.systemRole,
  configs: adminRoutePaths.systemConfig,
  audits: adminRoutePaths.systemAudit
} as const;

export type AdminSectionKey = keyof typeof adminSectionPaths;

type AdminPageProps = {
  activeKey: AdminSectionKey;
  children: ReactNode;
};

export default function AdminLayout({ activeKey, children }: AdminPageProps) {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const { t } = useI18n();
  const user = useAuthStore((state) => state.user);
  const menuItems = [
    hasAccess(user, 'admin.dashboard')
      ? {
          key: 'dashboard',
          label: t('nav.home'),
          icon: <DashboardOutlined />
        }
      : null,
    (() => {
      const children = [
        hasAccess(user, 'admin.users') ? { key: 'users', label: t('admin.menuUsers'), icon: <TeamOutlined /> } : null,
        hasAccess(user, 'admin.files') ? { key: 'files', label: t('admin.menuFiles'), icon: <FolderOpenOutlined /> } : null,
        hasAccess(user, 'admin.roles') ? { key: 'roles', label: t('admin.menuRoles'), icon: <SafetyCertificateOutlined /> } : null,
        hasAccess(user, 'admin.configs') ? { key: 'configs', label: t('admin.menuConfigs'), icon: <DatabaseOutlined /> } : null,
        hasAccess(user, 'admin.audits') ? { key: 'audits', label: t('admin.menuAudits'), icon: <FileSearchOutlined /> } : null
      ].filter(Boolean);

      if (!children.length) {
        return null;
      }

      return {
        key: 'system',
        label: t('nav.admin'),
        icon: <SettingOutlined />,
        children
      };
    })()
  ].filter(Boolean) as Array<{
    key: string;
    label: string;
    icon: ReactNode;
    children?: Array<{ key: string; label: string; icon: ReactNode }>;
  }>;

  return (
    <div className="app-shell p-3 pb-8">
      <MainNavbar />
      <SidebarLayout
        title={t('nav.admin')}
        items={menuItems}
        activeKey={activeKey}
        onChange={(key) => navigate(adminSectionPaths[key])}
        collapsed={collapsed}
        onToggle={() => setCollapsed((value) => !value)}
        collapseLabel={t('admin.menuCollapse')}
        expandLabel={t('admin.menuExpand')}
      >
        {children}
      </SidebarLayout>
    </div>
  );
}
