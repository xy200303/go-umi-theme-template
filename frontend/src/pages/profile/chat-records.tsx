import { ReloadOutlined } from '@ant-design/icons';
import { Button, Input, Select, Space, Table, Tag, Tooltip } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import { listChatRecords, listUserInterfaces, type ChatRecordItem, type UserInterfaceItem } from '@/api/endpoints/user-interface';
import { notifyApiError } from '@/lib/api-error';
import { useI18n } from '@/i18n';
import { ProfileShell } from './common';

type ChatRecordQueryState = {
  keyword: string;
  interfaceId?: number;
  statusCode?: number;
  page: number;
  pageSize: number;
};

function formatDateTime(value?: string): string {
  if (!value) return '-';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
}

function formatDurationMS(value?: number): string {
  if (typeof value !== 'number' || Number.isNaN(value)) return '-';
  return `${value} ms`;
}

function getStatusColor(statusCode: number): string {
  if (statusCode >= 500) return 'red';
  if (statusCode >= 400) return 'orange';
  if (statusCode >= 300) return 'blue';
  return 'green';
}

const statusCodeOptions = [200, 201, 400, 401, 403, 404, 422, 429, 500];

export default function ProfileChatRecordsPage() {
  const { t } = useI18n();
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<ChatRecordItem[]>([]);
  const [interfaces, setInterfaces] = useState<UserInterfaceItem[]>([]);
  const [total, setTotal] = useState(0);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedInterfaceId, setSelectedInterfaceId] = useState<number>();
  const [selectedStatusCode, setSelectedStatusCode] = useState<number>();
  const [query, setQuery] = useState<ChatRecordQueryState>({ keyword: '', page: 1, pageSize: 20 });

  const interfaceOptions = useMemo(() => interfaces.map((item) => ({ value: item.id, label: item.name })), [interfaces]);
  const interfaceNameMap = useMemo(() => new Map(interfaces.map((item) => [item.id, item.name])), [interfaces]);

  const loadRecords = async (nextQuery: ChatRecordQueryState) => {
    setLoading(true);
    try {
      const response = await listChatRecords({
        keyword: nextQuery.keyword || undefined,
        interface_id: nextQuery.interfaceId,
        status_code: nextQuery.statusCode,
        page: nextQuery.page,
        page_size: nextQuery.pageSize
      });
      setItems(response.list);
      setTotal(response.total);
      setQuery({
        keyword: nextQuery.keyword,
        interfaceId: nextQuery.interfaceId,
        statusCode: nextQuery.statusCode,
        page: response.page,
        pageSize: response.page_size
      });
    } catch (error) {
      notifyApiError(error, t('profile.chatRecordsLoadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const bootstrap = async () => {
      try {
        const interfaceList = await listUserInterfaces();
        setInterfaces(interfaceList);
      } catch (error) {
        notifyApiError(error, t('profile.chatRecordsLoadFailed'));
      }
      await loadRecords({ keyword: '', page: 1, pageSize: 20 });
    };
    void bootstrap();
  }, [t]);

  const columns = useMemo<ColumnsType<ChatRecordItem>>(
    () => [
      {
        title: t('profile.chatRecordsTableInterface'),
        width: 180,
        render: (_value, record) => interfaceNameMap.get(record.user_interface_id) || `#${record.user_interface_id}`
      },
      {
        title: t('profile.chatRecordsTableModel'),
        dataIndex: 'model',
        width: 180,
        render: (value: string) => value || '-'
      },
      {
        title: t('profile.chatRecordsTableInput'),
        dataIndex: 'user_input',
        width: 360,
        render: (value: string) =>
          value ? (
            <Tooltip title={value}>
              <span className="block overflow-hidden text-ellipsis whitespace-nowrap text-xs text-slate-600">{value}</span>
            </Tooltip>
          ) : (
            '-'
          )
      },
      {
        title: t('profile.chatRecordsTableOutput'),
        dataIndex: 'model_output',
        width: 420,
        render: (value: string) =>
          value ? (
            <Tooltip title={value}>
              <span className="block overflow-hidden text-ellipsis whitespace-nowrap text-xs text-slate-600">{value}</span>
            </Tooltip>
          ) : (
            '-'
          )
      },
      {
        title: t('profile.chatRecordsTableStatus'),
        dataIndex: 'status_code',
        width: 110,
        render: (value: number) => <Tag color={getStatusColor(value)}>{value}</Tag>
      },
      {
        title: t('profile.chatRecordsTableDuration'),
        dataIndex: 'duration_ms',
        width: 120,
        render: (value: number) => formatDurationMS(value)
      },
      {
        title: t('profile.chatRecordsTableError'),
        dataIndex: 'error_message',
        width: 260,
        render: (value: string) =>
          value ? (
            <Tooltip title={value}>
              <span className="block overflow-hidden text-ellipsis whitespace-nowrap text-xs text-rose-500">{value}</span>
            </Tooltip>
          ) : (
            '-'
          )
      },
      {
        title: t('profile.chatRecordsTableCreatedAt'),
        dataIndex: 'created_at',
        width: 180,
        render: (value: string) => formatDateTime(value)
      }
    ],
    [interfaceNameMap, t]
  );

  const onSearch = async () => {
    await loadRecords({
      keyword: searchKeyword.trim(),
      interfaceId: selectedInterfaceId,
      statusCode: selectedStatusCode,
      page: 1,
      pageSize: query.pageSize
    });
  };

  const onReset = async () => {
    setSearchKeyword('');
    setSelectedInterfaceId(undefined);
    setSelectedStatusCode(undefined);
    await loadRecords({ keyword: '', page: 1, pageSize: query.pageSize });
  };

  const onTableChange = async (pagination: TablePaginationConfig) => {
    await loadRecords({
      keyword: query.keyword,
      interfaceId: query.interfaceId,
      statusCode: query.statusCode,
      page: pagination.current ?? query.page,
      pageSize: pagination.pageSize ?? query.pageSize
    });
  };

  return (
    <ProfileShell activeKey="chatRecords">
      <div className="space-y-5">
        <section className="rounded-[28px] border border-slate-200 bg-white/90 px-6 py-7 shadow-[0_12px_40px_rgba(15,23,42,0.04)]">
          <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h2 className="text-[20px] font-semibold tracking-tight text-slate-900">{t('profile.chatRecordsTitle')}</h2>
              <p className="mt-3 text-[15px] leading-7 text-slate-500">{t('profile.chatRecordsDesc')}</p>
            </div>
            <Space wrap>
              <Input
                allowClear
                className="ant-surface-input min-w-[220px] sm:min-w-[280px]"
                placeholder={t('profile.chatRecordsSearchPlaceholder')}
                value={searchKeyword}
                onChange={(event) => setSearchKeyword(event.target.value)}
                onPressEnter={() => void onSearch()}
              />
              <Select
                allowClear
                className="ant-surface-select min-w-[180px]"
                options={interfaceOptions}
                placeholder={t('profile.chatRecordsInterfacePlaceholder')}
                value={selectedInterfaceId}
                onChange={(value) => setSelectedInterfaceId(value)}
              />
              <Select
                allowClear
                className="ant-surface-select min-w-[160px]"
                options={statusCodeOptions.map((code) => ({ value: code, label: String(code) }))}
                placeholder={t('profile.chatRecordsStatusPlaceholder')}
                value={selectedStatusCode}
                onChange={(value) => setSelectedStatusCode(value)}
              />
              <Button className="ant-surface-btn-primary !h-11" onClick={() => void onSearch()} type="primary">
                {t('admin.searchButton')}
              </Button>
              <Button className="ant-surface-btn-outline !h-11" icon={<ReloadOutlined />} onClick={() => void loadRecords(query)}>
                {t('profile.chatRecordsRefresh')}
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
          dataSource={items}
          loading={loading}
          scroll={{ x: 1900 }}
          pagination={{
            current: query.page,
            pageSize: query.pageSize,
            total,
            showSizeChanger: true,
            showTotal: (count) => `${t('profile.chatRecordsTotal')}: ${count}`
          }}
          onChange={onTableChange}
        />
      </div>
    </ProfileShell>
  );
}
