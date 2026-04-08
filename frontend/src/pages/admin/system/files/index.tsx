import { useEffect, useMemo, useState } from 'react';
import { Button, Input, Select, Space, Spin, Table, Tag } from 'antd';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import AdminLayout from '../../layout';
import { extractApiErrorMessage, formatDateTime } from '../../common';
import { getAdminFileStats, listAdminFiles, type AdminFileItem, type AdminFileStats } from '@/api/endpoints/admin';
import TechStatCard from '@/components/ui/TechStatCard';
import { useI18n } from '@/i18n';
import { notifyError } from '@/lib/notify';

type FileQueryState = {
  keyword: string;
  uploadStatus?: string;
  page: number;
  pageSize: number;
};

const initialStats: AdminFileStats = {
  total_count: 0,
  uploaded_count: 0,
  bound_count: 0,
  deleted_count: 0
};

function formatFileSize(size: number): string {
  if (!Number.isFinite(size) || size <= 0) {
    return '0 B';
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = size;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value >= 10 || unitIndex === 0 ? value.toFixed(0) : value.toFixed(1)} ${units[unitIndex]}`;
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'bound':
      return 'green';
    case 'deleted':
      return 'red';
    case 'uploaded':
      return 'blue';
    default:
      return 'default';
  }
}

export default function AdminFilesPage() {
  const { t } = useI18n();
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<AdminFileStats>(initialStats);
  const [files, setFiles] = useState<AdminFileItem[]>([]);
  const [total, setTotal] = useState(0);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedUploadStatus, setSelectedUploadStatus] = useState<string>();
  const [query, setQuery] = useState<FileQueryState>({ keyword: '', page: 1, pageSize: 20 });

  const uploadStatusOptions = useMemo(
    () => [
      { value: 'uploaded', label: t('admin.fileStatusUploaded') },
      { value: 'bound', label: t('admin.fileStatusBound') },
      { value: 'deleted', label: t('admin.fileStatusDeleted') }
    ],
    [t]
  );

  const loadFilesPage = async (nextQuery: FileQueryState) => {
    setLoading(true);
    try {
      const params = {
        keyword: nextQuery.keyword || undefined,
        upload_status: nextQuery.uploadStatus || undefined
      };

      const [statsResp, listResp] = await Promise.all([
        getAdminFileStats(params),
        listAdminFiles({
          ...params,
          page: nextQuery.page,
          page_size: nextQuery.pageSize
        })
      ]);

      setStats(statsResp);
      setFiles(listResp.list);
      setTotal(listResp.total);
      setQuery({
        keyword: nextQuery.keyword,
        uploadStatus: nextQuery.uploadStatus,
        page: listResp.page,
        pageSize: listResp.page_size
      });
    } catch (error) {
      notifyError(extractApiErrorMessage(error) ?? t('admin.fileLoadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadFilesPage({ keyword: '', page: 1, pageSize: 20 });
  }, [t]);

  const columns = useMemo<ColumnsType<AdminFileItem>>(
    () => [
      {
        title: t('admin.tableFileName'),
        dataIndex: 'original_name',
        width: 220
      },
      {
        title: t('admin.tableFileStatus'),
        dataIndex: 'upload_status',
        width: 120,
        render: (value: string) => <Tag color={getStatusColor(value)}>{t(`admin.fileStatusValue.${value}`)}</Tag>
      },
      {
        title: t('admin.tableFileSize'),
        dataIndex: 'size',
        width: 120,
        render: (value: number) => formatFileSize(value)
      },
      {
        title: t('admin.tableStorageDriver'),
        dataIndex: 'storage_driver',
        width: 120,
        render: (value: string) => value || '-'
      },
      {
        title: t('admin.tableMimeType'),
        dataIndex: 'mime_type',
        width: 180,
        render: (value: string) => value || '-'
      },
      {
        title: t('admin.tableUploadedBy'),
        dataIndex: 'uploaded_by',
        width: 120
      },
      {
        title: t('admin.tableCreatedAt'),
        dataIndex: 'created_at',
        width: 180,
        render: (value: string) => formatDateTime(value)
      },
      {
        title: t('admin.tableActions'),
        width: 160,
        fixed: 'right',
        render: (_value, record) => (
          <Space>
            <Button
              className="ant-surface-btn-outline !h-9"
              disabled={!record.file_url}
              href={record.file_url || undefined}
              rel="noreferrer"
              target="_blank"
              type="default"
            >
              {t('admin.fileDownloadButton')}
            </Button>
          </Space>
        )
      }
    ],
    [t]
  );

  const onSearch = async () => {
    await loadFilesPage({
      keyword: searchKeyword.trim(),
      uploadStatus: selectedUploadStatus,
      page: 1,
      pageSize: query.pageSize
    });
  };

  const onReset = async () => {
    setSearchKeyword('');
    setSelectedUploadStatus(undefined);
    await loadFilesPage({ keyword: '', page: 1, pageSize: query.pageSize });
  };

  const onTableChange = async (pagination: TablePaginationConfig) => {
    await loadFilesPage({
      keyword: query.keyword,
      uploadStatus: query.uploadStatus,
      page: pagination.current ?? query.page,
      pageSize: pagination.pageSize ?? query.pageSize
    });
  };

  return (
    <AdminLayout activeKey="files">
      {loading ? (
        <div className="flex justify-center py-10">
          <Spin size="large" />
        </div>
      ) : (
        <div className="space-y-5">
          <section className="rounded-2xl border border-blue-200/60 bg-white/70 p-4">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 className="text-lg font-semibold text-sky-900">{t('admin.filePanelTitle')}</h3>
                <p className="mt-1 text-sm text-slate-500">{t('admin.filePanelDesc')}</p>
              </div>
              <Space wrap>
                <Input
                  allowClear
                  className="ant-surface-input min-w-[220px] sm:min-w-[280px]"
                  placeholder={t('admin.fileSearchPlaceholder')}
                  value={searchKeyword}
                  onChange={(event) => setSearchKeyword(event.target.value)}
                  onPressEnter={() => void onSearch()}
                />
                <Select
                  allowClear
                  className="ant-surface-select min-w-[160px]"
                  options={uploadStatusOptions}
                  placeholder={t('admin.fileStatusPlaceholder')}
                  value={selectedUploadStatus}
                  onChange={(value) => setSelectedUploadStatus(value)}
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

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
            <TechStatCard title={t('admin.fileStatsTotal')} value={stats.total_count} />
            <TechStatCard title={t('admin.fileStatsUploaded')} value={stats.uploaded_count} />
            <TechStatCard title={t('admin.fileStatsBound')} value={stats.bound_count} />
            <TechStatCard title={t('admin.fileStatsDeleted')} value={stats.deleted_count} />
          </div>

          <Table
            rowKey="id"
            columns={columns}
            dataSource={files}
            scroll={{ x: 1500 }}
            pagination={{
              current: query.page,
              pageSize: query.pageSize,
              total,
              showSizeChanger: true,
              showTotal: (count) => `${t('admin.fileTotal')}: ${count}`
            }}
            onChange={onTableChange}
          />
        </div>
      )}
    </AdminLayout>
  );
}
