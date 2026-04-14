import { useEffect, useMemo, useState } from 'react';
import { CheckCircleOutlined, CopyOutlined, DeleteOutlined, EditOutlined, PlusOutlined, RedoOutlined } from '@ant-design/icons';
import { Button, Input, Modal, Popconfirm, Select, Space, Switch, Table, Tag, Tooltip, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  createUserInterface,
  deleteUserInterface,
  getGatewayDisplayConfig,
  listUserInterfaces,
  regenerateUserInterfaceGatewayKey,
  testUserInterfaceConnection,
  updateUserInterface,
  type InterfaceType,
  type UserInterfaceItem
} from '@/api/endpoints/user-interface';
import { notifyApiError } from '@/lib/api-error';
import { notifySuccess, notifyWarning } from '@/lib/notify';
import { useI18n } from '@/i18n';
import { getClientEnv } from '@/lib/env';
import { ProfileShell } from './common';

type CompatCard = {
  key: InterfaceType;
  title: string;
  desc: string;
  endpoint: string;
  authLabel: string;
  online: boolean;
};

type InterfaceFormState = {
  id?: number;
  name: string;
  interface_type: InterfaceType;
  target_base_url: string;
  target_api_key: string;
  default_model: string;
  enabled: boolean;
};

const interfaceTypeOptions: Array<{ value: InterfaceType; label: string }> = [
  { value: 'openai_api', label: 'OpenAI Chat Completions' },
  { value: 'openai_response', label: 'OpenAI Responses' },
  { value: 'claude', label: 'Anthropic Claude Messages' },
  { value: 'gemini', label: 'Google Gemini' }
];

const labelCls = 'mb-1 block text-sm font-medium text-slate-600';

function createEmptyForm(): InterfaceFormState {
  return {
    id: undefined,
    name: '',
    interface_type: 'openai_api',
    target_base_url: '',
    target_api_key: '',
    default_model: '',
    enabled: true
  };
}

function mapRecordToForm(record: UserInterfaceItem): InterfaceFormState {
  return {
    id: record.id,
    name: record.name,
    interface_type: record.interface_type,
    target_base_url: record.target_base_url,
    target_api_key: '',
    default_model: record.default_model ?? '',
    enabled: record.enabled
  };
}

function getInterfaceTypeLabel(value: InterfaceType): string {
  return interfaceTypeOptions.find((item) => item.value === value)?.label ?? value;
}

function getAPIOriginFromEnv(): string {
  const apiBaseURL = getClientEnv('API_BASE_URL')?.trim();
  if (!apiBaseURL) {
    return '';
  }

  if (apiBaseURL.startsWith('http://') || apiBaseURL.startsWith('https://')) {
    try {
      const parsed = new URL(apiBaseURL);
      return parsed.origin;
    } catch {
      return '';
    }
  }

  return '';
}

function normalizeGatewayBaseURL(raw?: string): string {
  const trimmed = raw?.trim();
  if (!trimmed) {
    return getAPIOriginFromEnv();
  }
  return trimmed.replace(/\/$/, '');
}

function getSuggestedFullPath(interfaceType: InterfaceType): string {
  switch (interfaceType) {
    case 'openai_api':
      return '/v1/chat/completions';
    case 'openai_response':
      return '/v1/responses';
    case 'claude':
      return '/v1/messages';
    case 'gemini':
      return '/v1beta/models/{model}:generateContent';
    default:
      return '/v1/chat/completions';
  }
}

function splitPathSegments(pathname: string): string[] {
  return pathname
    .split('/')
    .map((segment) => segment.trim())
    .filter(Boolean);
}

function longestSuffixPrefixOverlap(currentSegments: string[], requiredSegments: string[]): number {
  const maxOverlap = Math.min(currentSegments.length, requiredSegments.length);
  for (let overlap = maxOverlap; overlap >= 1; overlap -= 1) {
    const start = currentSegments.length - overlap;
    let matched = true;
    for (let index = 0; index < overlap; index += 1) {
      if (currentSegments[start + index] !== requiredSegments[index]) {
        matched = false;
        break;
      }
    }
    if (matched) {
      return overlap;
    }
  }
  return 0;
}

function getRequiredPathSegments(interfaceType: InterfaceType): string[] {
  switch (interfaceType) {
    case 'openai_response':
      return ['v1', 'responses'];
    case 'claude':
      return ['v1', 'messages'];
    case 'gemini':
      return ['v1beta', 'models', '{model}:generateContent'];
    case 'openai_api':
    default:
      return ['v1', 'chat', 'completions'];
  }
}

function normalizeTargetPathWithPrefix(pathname: string, interfaceType: InterfaceType): string {
  const cleanedPath = pathname.trim().replace(/\/+$/, '');
  if (!cleanedPath || cleanedPath === '/') {
    return getSuggestedFullPath(interfaceType);
  }

  if (interfaceType === 'gemini') {
    const lower = cleanedPath.toLowerCase();
    if (lower.includes('/models/') && lower.includes(':generatecontent')) {
      return cleanedPath.startsWith('/') ? cleanedPath : `/${cleanedPath}`;
    }
  }

  const currentSegments = splitPathSegments(cleanedPath);
  const requiredSegments = getRequiredPathSegments(interfaceType);
  const overlap = longestSuffixPrefixOverlap(currentSegments, requiredSegments);
  const nextSegments = [...currentSegments, ...requiredSegments.slice(overlap)];
  return `/${nextSegments.join('/')}`;
}

function getNormalizedTargetBaseURLPreview(raw: string, interfaceType: InterfaceType): string {
  const trimmed = raw.trim();
  if (!trimmed) {
    return '';
  }

  const withProtocol = trimmed.includes('://') ? trimmed : `https://${trimmed}`;

  try {
    const parsed = new URL(withProtocol);
    parsed.pathname = normalizeTargetPathWithPrefix(parsed.pathname, interfaceType);
    parsed.search = '';
    parsed.hash = '';
    return parsed.toString().replace(/\/$/, '');
  } catch {
    return '';
  }
}

function buildCompatCards(items: UserInterfaceItem[], gatewayBaseURL: string): CompatCard[] {
  const origin = normalizeGatewayBaseURL(gatewayBaseURL);
  const hasEnabled = (type: InterfaceType) => items.some((item) => item.interface_type === type && item.enabled);
  const buildEndpoint = (path: string) => (origin ? `${origin}${path}` : path);

  return [
    {
      key: 'openai_api',
      title: 'OpenAI Chat Completions',
      desc: '兼容 OpenAI Chat Completions 调用格式。',
      endpoint: buildEndpoint('/v1/chat/completions'),
      authLabel: 'Authorization: Bearer <API_KEY>',
      online: hasEnabled('openai_api')
    },
    {
      key: 'openai_response',
      title: 'OpenAI Responses',
      desc: '兼容 OpenAI Responses 调用格式。',
      endpoint: buildEndpoint('/v1/responses'),
      authLabel: 'Authorization: Bearer <API_KEY>',
      online: hasEnabled('openai_response')
    },
    {
      key: 'claude',
      title: 'Anthropic Claude Messages',
      desc: '兼容 Claude / Anthropic Messages 调用格式。',
      endpoint: buildEndpoint('/anthropic/v1/messages'),
      authLabel: 'x-api-key + anthropic-version',
      online: hasEnabled('claude')
    },
    {
      key: 'gemini',
      title: 'Google Gemini',
      desc: '兼容 Gemini 原生接口调用格式。',
      endpoint: buildEndpoint('/gemini/v1beta/models/{model}:generateContent'),
      authLabel: 'x-goog-api-key 或 key=API_KEY',
      online: hasEnabled('gemini')
    }
  ];
}

export default function ProfileInterfacesPage() {
  const { t } = useI18n();
  const [items, setItems] = useState<UserInterfaceItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [modalOpen, setModalOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testingId, setTestingId] = useState<number | null>(null);
  const [gatewayBaseURL, setGatewayBaseURL] = useState('');
  const [formData, setFormData] = useState<InterfaceFormState>(createEmptyForm());

  const compatCards = useMemo(() => buildCompatCards(items, gatewayBaseURL), [gatewayBaseURL, items]);
  const pagedItems = useMemo(() => {
    const start = (page - 1) * pageSize;
    return items.slice(start, start + pageSize);
  }, [items, page, pageSize]);
  const normalizedTargetBaseURLPreview = useMemo(
    () => getNormalizedTargetBaseURLPreview(formData.target_base_url, formData.interface_type),
    [formData.interface_type, formData.target_base_url]
  );

  const columns = useMemo<ColumnsType<UserInterfaceItem>>(
    () => [
      {
        title: t('profile.interfacesTableName'),
        dataIndex: 'name',
        key: 'name',
        width: 180
      },
      {
        title: t('profile.interfacesTableType'),
        dataIndex: 'interface_type',
        key: 'interface_type',
        width: 240,
        render: (value: InterfaceType) => <Tag color="blue">{getInterfaceTypeLabel(value)}</Tag>
      },
      {
        title: t('profile.interfacesTableBaseUrl'),
        dataIndex: 'target_base_url',
        key: 'target_base_url',
        width: 260,
        ellipsis: true
      },
      {
        title: t('profile.interfacesTableModel'),
        dataIndex: 'default_model',
        key: 'default_model',
        width: 140,
        render: (value: string) => value || '-'
      },
      {
        title: t('profile.interfacesTableKey'),
        dataIndex: 'gateway_key_prefix',
        key: 'gateway_key_prefix',
        width: 420,
        render: (value: string, record) => {
          const displayKey = record.gateway_key || `${value}...`;
          return (
            <Space size="small" wrap={false}>
              <Typography.Text
                code
                className="!mb-0 !inline-block !whitespace-nowrap"
                style={{ maxWidth: 280 }}
                ellipsis={{ tooltip: displayKey }}
              >
                {displayKey}
              </Typography.Text>
              <Button
                icon={<CopyOutlined />}
                onClick={() => copyText(record.gateway_key || displayKey, t('profile.interfacesKeyCopied'), t('profile.interfacesKeyCopyFailed'))}
                type="text"
              >
                {t('profile.interfacesCopyKey')}
              </Button>
            </Space>
          );
        }
      },
      {
        title: t('profile.interfacesTableStatus'),
        dataIndex: 'enabled',
        key: 'enabled',
        width: 120,
        render: (value: boolean) => <Tag color={value ? 'green' : 'default'}>{value ? t('profile.interfaceEnabled') : t('profile.interfaceDisabled')}</Tag>
      },
      {
        title: t('profile.interfacesTableActions'),
        key: 'actions',
        width: 150,
        render: (_, record) => (
          <Space size="small" wrap={false}>
            <Tooltip title={t('profile.interfacesEdit')}>
              <Button
                icon={<EditOutlined />}
                onClick={() => openEditModal(record)}
                type="text"
              />
            </Tooltip>
            <Tooltip title={t('profile.interfacesRegenerateKey')}>
              <Popconfirm
                title={t('profile.interfacesRegenerateKeyConfirm', { name: record.name })}
                onConfirm={() => handleRegenerateKey(record.id)}
                okText={t('profile.confirm')}
                cancelText={t('profile.cancel')}
              >
                <Button icon={<RedoOutlined />} type="text" />
              </Popconfirm>
            </Tooltip>
            <Tooltip title={t('profile.interfacesTest')}>
              <Button
                icon={<CheckCircleOutlined />}
                loading={testingId === record.id}
                onClick={() => handleTestConnection(record.id)}
                type="text"
              />
            </Tooltip>
            <Tooltip title={t('profile.interfacesDelete')}>
              <Popconfirm
                title={t('profile.interfacesDeleteConfirm', { name: record.name })}
                onConfirm={() => handleDelete(record.id)}
                okText={t('profile.confirm')}
                cancelText={t('profile.cancel')}
              >
                <Button danger icon={<DeleteOutlined />} type="text" />
              </Popconfirm>
            </Tooltip>
          </Space>
        )
      }
    ],
    [t]
  );

  const loadItems = async () => {
    setLoading(true);
    try {
      const [data, displayConfig] = await Promise.all([listUserInterfaces(), getGatewayDisplayConfig()]);
      setItems(data);
      setGatewayBaseURL(displayConfig.gateway_base_url);
    } catch (error) {
      notifyApiError(error, t('profile.interfacesLoadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadItems();
  }, []);

  useEffect(() => {
    const maxPage = Math.max(1, Math.ceil(items.length / pageSize));
    if (page > maxPage) {
      setPage(maxPage);
    }
  }, [items.length, page, pageSize]);

  const copyText = async (text: string, successMessage: string, failedMessage: string) => {
    if (!text) {
      notifyWarning(failedMessage);
      return;
    }
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text);
      } else {
        throw new Error('clipboard api unavailable');
      }
      notifySuccess(successMessage);
    } catch {
      try {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.setAttribute('readonly', 'true');
        textarea.style.position = 'fixed';
        textarea.style.left = '-9999px';
        document.body.appendChild(textarea);
        textarea.select();
        textarea.setSelectionRange(0, textarea.value.length);
        const copied = document.execCommand('copy');
        document.body.removeChild(textarea);
        if (!copied) {
          throw new Error('execCommand copy failed');
        }
        notifySuccess(successMessage);
      } catch {
        notifyWarning(failedMessage);
      }
    }
  };

  const openCreateModal = () => {
    setFormData(createEmptyForm());
    setModalOpen(true);
  };

  const openEditModal = (record: UserInterfaceItem) => {
    setFormData(mapRecordToForm(record));
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setSaving(false);
    setFormData(createEmptyForm());
  };

  const updateFormData = (patch: Partial<InterfaceFormState>) => {
    setFormData((prev) => ({ ...prev, ...patch }));
  };

  const handleCreate = async () => {
    if (!formData.name.trim() || !formData.target_base_url.trim()) {
      notifyWarning(t('profile.interfacesRequired'));
      return;
    }
    if (!formData.id && !formData.target_api_key.trim()) {
      notifyWarning(t('profile.interfacesApiKeyRequired'));
      return;
    }

    setSaving(true);
    try {
      if (formData.id) {
        const updated = await updateUserInterface(formData.id, {
          name: formData.name,
          interface_type: formData.interface_type,
          target_base_url: formData.target_base_url,
          target_api_key: formData.target_api_key || undefined,
          default_model: formData.default_model,
          enabled: formData.enabled
        });
        setItems((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
        notifySuccess(t('profile.interfacesUpdated'));
        closeModal();
        return;
      }

      const created = await createUserInterface({
        name: formData.name,
        interface_type: formData.interface_type,
        target_base_url: formData.target_base_url,
        target_api_key: formData.target_api_key,
        default_model: formData.default_model,
        enabled: formData.enabled
      });
      setItems((prev) => [created.interface, ...prev]);
      setPage(1);
      notifySuccess(t('profile.interfacesCreated'));
      closeModal();
    } catch (error) {
      notifyApiError(error, t('profile.interfacesCreateFailed'));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteUserInterface(id);
      setItems((prev) => prev.filter((item) => item.id !== id));
      notifySuccess(t('profile.interfacesDeleted'));
    } catch (error) {
      notifyApiError(error, t('profile.interfacesDeleteFailed'));
    }
  };

  const handleRegenerateKey = async (id: number) => {
    try {
      const result = await regenerateUserInterfaceGatewayKey(id);
      setItems((prev) => prev.map((item) => (item.id === result.interface.id ? result.interface : item)));
      notifySuccess(t('profile.interfacesRegenerated'));
      await copyText(result.gateway_key, t('profile.interfacesKeyCopied'), t('profile.interfacesKeyCopyFailed'));
    } catch (error) {
      notifyApiError(error, t('profile.interfacesRegenerateFailed'));
    }
  };

  const handleTestConnection = async (id: number) => {
    setTestingId(id);
    try {
      const result = await testUserInterfaceConnection(id);
      if (result.ok) {
        notifySuccess(`${t('profile.interfacesTestSuccess')} (${result.status_code})`);
      } else {
        notifyWarning(`${t('profile.interfacesTestFailed')} (${result.status_code || '-'}) ${result.message}`);
      }
      await loadItems();
    } catch (error) {
      notifyApiError(error, t('profile.interfacesTestFailed'));
    } finally {
      setTestingId(null);
    }
  };

  return (
    <ProfileShell activeKey="interfaces">
      <div className="flex min-h-full flex-col gap-4">
        <section className="rounded-[28px] border border-slate-200 bg-white/90 px-6 py-7 shadow-[0_12px_40px_rgba(15,23,42,0.05)]">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div>
              <h2 className="text-[20px] font-semibold tracking-tight text-slate-900">{t('profile.compatTitle')}</h2>
              <p className="mt-3 text-[15px] leading-7 text-slate-500">{t('profile.compatDesc')}</p>
            </div>
            <Button
              className="ant-surface-btn-outline !h-12 !rounded-full !px-6"
              icon={<CopyOutlined />}
              onClick={() =>
                copyText(normalizeGatewayBaseURL(gatewayBaseURL), t('profile.interfacesBaseUrlCopied'), t('profile.interfacesBaseUrlCopyFailed'))
              }
              type="default"
            >
              {t('profile.interfacesCopyBaseUrl')}
            </Button>
          </div>

          <div className="mt-8 grid grid-cols-1 gap-4 xl:grid-cols-2">
            {compatCards.map((card) => (
              <div key={card.key} className="rounded-[28px] border border-slate-200 bg-white px-6 py-6 shadow-[0_8px_26px_rgba(15,23,42,0.03)]">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <h3 className="text-[18px] font-semibold tracking-tight text-slate-900">{card.title}</h3>
                    <p className="mt-3 text-[15px] leading-7 text-slate-500">{card.desc}</p>
                  </div>
                  <span
                    className={`inline-flex min-w-[84px] justify-center rounded-full px-4 py-2 text-[14px] font-medium ${
                      card.online ? 'bg-lime-100 text-lime-700' : 'bg-slate-100 text-slate-500'
                    }`}
                  >
                    {card.online ? 'online' : 'offline'}
                  </span>
                </div>

                <div className="mt-8 rounded-[22px] bg-slate-50 px-5 py-5">
                  <div>
                    <div className="flex items-center justify-between gap-3">
                      <p className="text-[16px] font-semibold text-slate-900">Endpoint</p>
                      <Button
                        icon={<CopyOutlined />}
                        onClick={() => copyText(card.endpoint, t('profile.interfacesEndpointCopied'), t('profile.interfacesEndpointCopyFailed'))}
                        size="small"
                        type="text"
                      >
                        {t('profile.interfacesCopyEndpoint')}
                      </Button>
                    </div>
                    <p className="mt-2 break-all font-mono text-[14px] leading-7 text-slate-700">{card.endpoint}</p>
                  </div>
                  <div className="mt-5">
                    <p className="text-[16px] font-semibold text-slate-900">Auth</p>
                    <p className="mt-2 break-all text-[14px] leading-7 text-slate-700">{card.authLabel}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="rounded-[28px] border border-slate-200 bg-white/90 px-6 py-7 shadow-[0_12px_40px_rgba(15,23,42,0.04)]">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div>
              <h2 className="text-[20px] font-semibold tracking-tight text-slate-900">{t('profile.keysTitle')}</h2>
              <p className="mt-3 text-[15px] leading-7 text-slate-500">{t('profile.keysDesc')}</p>
            </div>
            <Button
              className="ant-surface-btn-primary !h-12 !rounded-full !px-6"
              icon={<PlusOutlined />}
              onClick={openCreateModal}
              type="primary"
            >
              {t('profile.interfacesCreate')}
            </Button>
          </div>

          <div className="mt-6 rounded-2xl border border-slate-200 bg-white">
            <Table
              columns={columns}
              dataSource={pagedItems}
              loading={loading}
              pagination={{
                current: page,
                pageSize,
                total: items.length,
                showSizeChanger: true,
                pageSizeOptions: ['5', '10', '20', '50'],
                onChange: (nextPage, nextPageSize) => {
                  setPage(nextPage);
                  setPageSize(nextPageSize);
                }
              }}
              rowKey="id"
              scroll={{ x: 1200 }}
            />
          </div>
        </section>
      </div>

      <Modal
        open={modalOpen}
        onCancel={closeModal}
        onOk={handleCreate}
        okText={formData.id ? t('profile.interfacesSave') : t('profile.interfacesCreate')}
        cancelText={t('profile.cancel')}
        confirmLoading={saving}
        title={formData.id ? t('profile.interfacesEditTitle') : t('profile.interfacesCreateTitle')}
      >
        <div className="space-y-4 pt-2">
          <label className="block">
            <span className={labelCls}>{t('profile.interfacesFieldName')}</span>
            <Input
              className="ant-surface-input"
              placeholder={t('profile.interfacesFieldNamePlaceholder')}
              value={formData.name}
              onChange={(e) => updateFormData({ name: e.target.value })}
            />
          </label>

          <label className="block">
            <span className={labelCls}>{t('profile.interfacesFieldType')}</span>
            <Select
              className="ant-surface-select"
              options={interfaceTypeOptions}
              value={formData.interface_type}
              onChange={(value) => updateFormData({ interface_type: value })}
            />
          </label>

          <label className="block">
            <span className={labelCls}>{t('profile.interfacesFieldBaseUrl')}</span>
            <Input
              className="ant-surface-input"
              placeholder={t('profile.interfacesFieldBaseUrlPlaceholder')}
              value={formData.target_base_url}
              onChange={(e) => updateFormData({ target_base_url: e.target.value })}
            />
            <p className="mt-2 text-xs leading-6 text-slate-500">{t('profile.interfacesFieldBaseUrlHint')}</p>
            {normalizedTargetBaseURLPreview ? (
              <div className="mt-2 rounded-lg border border-sky-100 bg-sky-50 px-3 py-2 text-xs leading-6 text-sky-700">
                <span className="font-medium">{t('profile.interfacesFieldBaseUrlPreviewLabel')}</span>
                <span className="ml-2 break-all font-mono">{normalizedTargetBaseURLPreview}</span>
              </div>
            ) : null}
          </label>

          <label className="block">
            <span className={labelCls}>{t('profile.interfacesFieldApiKey')}</span>
            <Input.Password
              className="ant-surface-input"
              placeholder={formData.id ? t('profile.interfacesFieldApiKeyEditPlaceholder') : t('profile.interfacesFieldApiKeyPlaceholder')}
              value={formData.target_api_key}
              onChange={(e) => updateFormData({ target_api_key: e.target.value })}
            />
          </label>

          <label className="block">
            <span className={labelCls}>{t('profile.interfacesFieldModel')}</span>
            <Input
              className="ant-surface-input"
              placeholder={t('profile.interfacesFieldModelPlaceholder')}
              value={formData.default_model}
              onChange={(e) => updateFormData({ default_model: e.target.value })}
            />
          </label>

          <div className="flex items-center justify-between rounded-lg border border-slate-200 bg-slate-50 px-3 py-3">
            <div>
              <p className="text-sm font-medium text-slate-700">{t('profile.interfacesFieldEnabled')}</p>
              <p className="text-xs text-slate-500">{t('profile.interfacesFieldEnabledHint')}</p>
            </div>
            <Switch checked={formData.enabled} onChange={(checked) => updateFormData({ enabled: checked })} />
          </div>
        </div>
      </Modal>
    </ProfileShell>
  );
}
