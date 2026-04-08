import { useEffect, useMemo, useState } from 'react';
import { Button, Input, Select, Space, Spin, Table, Tag, Tooltip } from 'antd';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import AdminLayout from '../../layout';
import { formatDateTime, formatDurationMS, localizePolicyActionLabel, localizePolicyMenuLabel } from '../../common';
import { listAuditLogs, type AuditLogItem } from '@/api/endpoints/admin';
import { useI18n } from '@/i18n';
import { notifyApiError } from '@/lib/api-error';

type AuditQueryState = {
  keyword: string;
  menuKey?: string;
  statusCode?: number;
  page: number;
  pageSize: number;
};



const statusCodeOptions = [200, 201, 400, 401, 403, 404, 422, 500];

function getMethodColor(method: string): string {
  switch (method) {
    case 'GET':
      return 'blue';
    case 'POST':
      return 'green';
    case 'PUT':
      return 'gold';
    case 'DELETE':
      return 'red';
    default:
      return 'default';
  }
}

function getStatusColor(statusCode: number): string {
  if (statusCode >= 500) {
    return 'red';
  }
  if (statusCode >= 400) {
    return 'orange';
  }
  if (statusCode >= 300) {
    return 'blue';
  }
  return 'green';
}

export default function AdminAuditPage() {
  const { t } = useI18n();
  const [loading, setLoading] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedMenuKey, setSelectedMenuKey] = useState<string>();
  const [selectedStatusCode, setSelectedStatusCode] = useState<number>();
  const [query, setQuery] = useState<AuditQueryState>({ keyword: '', page: 1, pageSize: 20 });
  const [logs, setLogs] = useState<AuditLogItem[]>([]);
  const [total, setTotal] = useState(0);

  const moduleOptions = useMemo(
    () => [
      { value: 'auth', label: t('admin.auditModuleAuth') },
      { value: 'dashboard', label: t('admin.auditModuleDashboard') },
      { value: 'users', label: t('admin.auditModuleUsers') },
      { value: 'roles', label: t('admin.auditModuleRoles') },
      { value: 'configs', label: t('admin.auditModuleConfigs') },
      { value: 'audits', label: t('admin.auditModuleAudits') },
      { value: 'profile', label: t('admin.auditModuleProfile') }
    ],
    [t]
  );

  const loadAuditLogs = async (nextQuery: AuditQueryState) => {
    setLoading(true);
    try {
      const response = await listAuditLogs({
        keyword: nextQuery.keyword || undefined,
        menu_key: nextQuery.menuKey || undefined,
        status_code: nextQuery.statusCode,
        page: nextQuery.page,
        page_size: nextQuery.pageSize
      });
      setLogs(response.list);
      setTotal(response.total);
      setQuery({
        keyword: nextQuery.keyword,
        menuKey: nextQuery.menuKey,
        statusCode: nextQuery.statusCode,
        page: response.page,
        pageSize: response.page_size
      });
    } catch (error) {
      notifyApiError(error, t('admin.auditLoadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadAuditLogs({ keyword: '', page: 1, pageSize: 20 });
  }, [t]);

  const columns = useMemo<ColumnsType<AuditLogItem>>(
    () => [
      {
        title: t('admin.tableUsername'),
        dataIndex: 'username',
        width: 140
      },
      {
        title: t('admin.tableModule'),
        width: 140,
        render: (_value, record) => localizePolicyMenuLabel(record.menu_key, record.menu_label, t) || '-'
      },
      {
        title: t('admin.tableOperation'),
        width: 260,
        render: (_value, record) => (
          <div className="space-y-1">
            <div className="font-medium text-slate-700">
              {localizePolicyActionLabel(record.operation_id, record.operation_name, t) || t('admin.auditEmptyOperation')}
            </div>
            <div className="text-xs text-slate-500">{record.operation_id || '-'}</div>
          </div>
        )
      },
      {
        title: t('admin.tableMethod'),
        dataIndex: 'method',
        width: 100,
        render: (value: string) => <Tag color={getMethodColor(value)}>{value}</Tag>
      },
      {
        title: t('admin.tableRoutePath'),
        dataIndex: 'route_path',
        width: 240,
        render: (value: string) => <code className="text-xs text-slate-600">{value || '-'}</code>
      },
      {
        title: t('admin.tableRequestPath'),
        dataIndex: 'request_path',
        width: 240,
        render: (value: string) => <code className="text-xs text-slate-600">{value || '-'}</code>
      },
      {
        title: t('admin.tableStatusCode'),
        dataIndex: 'status_code',
        width: 110,
        render: (value: number) => <Tag color={getStatusColor(value)}>{value}</Tag>
      },
      {
        title: t('admin.tableClientIp'),
        dataIndex: 'client_ip',
        width: 140,
        render: (value: string) => value || '-'
      },
      {
        title: t('admin.tableDuration'),
        dataIndex: 'duration_ms',
        width: 110,
        render: (value: number) => formatDurationMS(value)
      },
      {
        title: t('admin.tableCreatedAt'),
        dataIndex: 'created_at',
        width: 180,
        render: (value: string) => formatDateTime(value)
      },
      {
        title: t('admin.tableUserAgent'),
        dataIndex: 'user_agent',
        width: 260,
        render: (value: string) =>
          value ? (
            <Tooltip title={value}>
              <span className="block overflow-hidden text-ellipsis whitespace-nowrap text-xs text-slate-500">{value}</span>
            </Tooltip>
          ) : (
            '-'
          )
      }
    ],
    [t]
  );

  const onSearch = async () => {
    await loadAuditLogs({
      keyword: searchKeyword.trim(),
      menuKey: selectedMenuKey,
      statusCode: selectedStatusCode,
      page: 1,
      pageSize: query.pageSize
    });
  };

  const onReset = async () => {
    setSearchKeyword('');
    setSelectedMenuKey(undefined);
    setSelectedStatusCode(undefined);
    await loadAuditLogs({ keyword: '', page: 1, pageSize: query.pageSize });
  };

  const onTableChange = async (pagination: TablePaginationConfig) => {
    await loadAuditLogs({
      keyword: query.keyword,
      menuKey: query.menuKey,
      statusCode: query.statusCode,
      page: pagination.current ?? query.page,
      pageSize: pagination.pageSize ?? query.pageSize
    });
  };

  return (
    <AdminLayout activeKey="audits">
      {loading ? (
        <div className="flex justify-center py-10">
          <Spin size="large" />
        </div>
      ) : (
        <div className="space-y-5">
          <section className="rounded-2xl border border-blue-200/60 bg-white/70 p-4">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 className="text-lg font-semibold text-sky-900">{t('admin.auditPanelTitle')}</h3>
                <p className="mt-1 text-sm text-slate-500">{t('admin.auditPanelDesc')}</p>
              </div>
              <Space wrap>
                <Input
                  allowClear
                  className="ant-surface-input min-w-[220px] sm:min-w-[280px]"
                  placeholder={t('admin.auditSearchPlaceholder')}
                  value={searchKeyword}
                  onChange={(event) => setSearchKeyword(event.target.value)}
                  onPressEnter={() => void onSearch()}
                />
                <Select
                  allowClear
                  className="ant-surface-select min-w-[180px]"
                  options={moduleOptions}
                  placeholder={t('admin.auditModulePlaceholder')}
                  value={selectedMenuKey}
                  onChange={(value) => setSelectedMenuKey(value)}
                />
                <Select
                  allowClear
                  className="ant-surface-select min-w-[160px]"
                  options={statusCodeOptions.map((code) => ({ value: code, label: String(code) }))}
                  placeholder={t('admin.auditStatusPlaceholder')}
                  value={selectedStatusCode}
                  onChange={(value) => setSelectedStatusCode(value)}
                />
                <Button className="ant-surface-btn-primary !h-11" onClick={() => void onSearch()} type="primary">
                  {t('admin.searchButton')}
                </Button>
                <Button className="ant-surface-btn-outline !h-11" onClick={() => void onReset()}>
                  {t('admin.resetButton')}
                </Button>
              </Space>
            </div>
          </section>

          <Table
            rowKey="id"
            columns={columns}
            dataSource={logs}
            scroll={{ x: 1900 }}
            pagination={{
              current: query.page,
              pageSize: query.pageSize,
              total,
              showSizeChanger: true,
              showTotal: (count) => `${t('admin.auditTotal')}: ${count}`
            }}
            onChange={onTableChange}
          />
        </div>
      )}
    </AdminLayout>
  );
}
