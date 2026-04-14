import { useState, type ReactNode } from 'react';
import { ApiOutlined, HistoryOutlined, MessageOutlined, UserOutlined } from '@ant-design/icons';
import MainNavbar from '@/components/layout/MainNavbar';
import SidebarLayout from '@/components/layout/SidebarLayout';
import { routePaths } from '@/constants/routes';
import { useNavigate } from '@/lib/router';
import { useI18n } from '@/i18n';

export const profileSectionPaths = {
  profile: routePaths.profile,
  interfaces: routePaths.profileInterfaces,
  gatewayLogs: routePaths.profileGatewayLogs,
  chatRecords: routePaths.profileChatRecords
} as const;

export type ProfileSectionKey = keyof typeof profileSectionPaths;

export function ProfileShell({ activeKey, children }: { activeKey: ProfileSectionKey; children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const { t } = useI18n();

  const profileMenuItems = [
    {
      key: 'profile',
      label: t('profile.menuProfile'),
      icon: <UserOutlined />
    },
    {
      key: 'interfaces',
      label: t('profile.menuInterfaces'),
      icon: <ApiOutlined />
    },
    {
      key: 'gatewayLogs',
      label: t('profile.menuGatewayLogs'),
      icon: <HistoryOutlined />
    },
    {
      key: 'chatRecords',
      label: t('profile.menuChatRecords'),
      icon: <MessageOutlined />
    }
  ];

  return (
    <div className="app-shell p-3 pb-8">
      <MainNavbar />
      <SidebarLayout
        title={t('nav.profile')}
        items={profileMenuItems}
        activeKey={activeKey}
        onChange={(key) => navigate(profileSectionPaths[key as ProfileSectionKey])}
        collapsed={collapsed}
        onToggle={() => setCollapsed((value) => !value)}
        collapseLabel={t('profile.menuCollapse')}
        expandLabel={t('profile.menuExpand')}
      >
        {children}
      </SidebarLayout>
    </div>
  );
}
