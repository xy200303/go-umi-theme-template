import type { RoleItem, RolePolicy, SystemConfigItem } from '@/api/endpoints/admin';
import { extractApiErrorMessage } from '@/lib/api-error';
import type { AuthUser } from '@/types/auth';

export { extractApiErrorMessage } from '@/lib/api-error';
export { notifyApiError } from '@/lib/api-error';

export type UserFormValues = {
  username: string;
  phone: string;
  password?: string;
  email: string;
  avatar_url: string;
  avatar_file_id: string;
  signature: string;
  gender: string;
  age?: number | null;
  is_active: boolean;
};

export type PasswordFormValues = {
  password: string;
};

export type RoleFormValues = {
  name: string;
  display_name: string;
  description: string;
};

export type ConfigFormValues = {
  config_group: string;
  config_key: string;
  config_val: string;
  remark: string;
};

export type PolicyTemplate = {
  key: string;
  menuKey: string;
  menuLabel: string;
  actionLabel: string;
  description?: string;
  method: string;
  path: string;
};

export type PolicyTemplateSection = {
  menuKey: string;
  menuLabel: string;
  items: PolicyTemplate[];
};

export type PolicyTemplateScopeKey = 'admin' | 'user' | 'other';

type TranslateFn = (key: string, vars?: Record<string, string | number>) => string;

export const formLabelClassName = 'mb-1 block text-sm font-medium text-slate-600';
export const usernamePattern = /^[A-Za-z0-9_]+$/;

const POLICY_METHOD_COLORS: Record<string, string> = {
  GET: 'blue',
  POST: 'green',
  PUT: 'gold',
  DELETE: 'red'
};

const SECTION_AGGREGATE_POLICIES: Record<string, RolePolicy[]> = {
  dashboard: [{ method: 'GET', path: '/api/v1/admin/stats' }],
  audits: [{ method: 'GET', path: '/api/v1/admin/audit-logs' }],
  files: [
    { method: 'GET', path: '/api/v1/admin/files' },
    { method: 'GET', path: '/api/v1/admin/files/stats' }
  ],
  users: [
    { method: '(GET|POST)', path: '/api/v1/admin/users' },
    { method: '(PUT|DELETE)', path: '/api/v1/admin/users/*' }
  ],
  roles: [
    { method: '(GET|POST)', path: '/api/v1/admin/roles' },
    { method: '(GET|PUT|DELETE)', path: '/api/v1/admin/roles/*' }
  ],
  configs: [{ method: '(GET|PUT)', path: '/api/v1/admin/system-configs' }],
  profile: [
    { method: '(GET|PUT)', path: '/api/v1/user/profile' },
    { method: 'POST', path: '/api/v1/user/*' }
  ]
};

const POLICY_MENU_I18N_KEYS: Record<string, string> = {
  auth: 'admin.auditModuleAuth',
  dashboard: 'admin.auditModuleDashboard',
  users: 'admin.auditModuleUsers',
  roles: 'admin.auditModuleRoles',
  configs: 'admin.auditModuleConfigs',
  audits: 'admin.auditModuleAudits',
  files: 'admin.menuFiles',
  profile: 'admin.auditModuleProfile'
};

const POLICY_OPERATION_I18N_KEYS: Record<string, { label: string; description?: string }> = {
  'dashboard.stats.get': {
    label: 'admin.policyActionViewStats',
    description: 'admin.policyDescViewStats'
  },
  'audits.list': {
    label: 'admin.policyActionListAudits',
    description: 'admin.policyDescListAudits'
  },
  'files.admin.list': {
    label: 'admin.policyActionListFiles',
    description: 'admin.policyDescListFiles'
  },
  'files.admin.stats': {
    label: 'admin.policyActionViewFileStats',
    description: 'admin.policyDescViewFileStats'
  },
  'files.upload': {
    label: 'admin.policyActionUploadFile',
    description: 'admin.policyDescUploadFile'
  },
  'files.direct.init': {
    label: 'admin.policyActionInitDirectUpload',
    description: 'admin.policyDescInitDirectUpload'
  },
  'files.direct.complete': {
    label: 'admin.policyActionCompleteDirectUpload',
    description: 'admin.policyDescCompleteDirectUpload'
  },
  'users.list': {
    label: 'admin.policyActionListUsers',
    description: 'admin.policyDescListUsers'
  },
  'users.create': {
    label: 'admin.policyActionCreateUser',
    description: 'admin.policyDescCreateUser'
  },
  'users.update': {
    label: 'admin.policyActionUpdateUser',
    description: 'admin.policyDescUpdateUser'
  },
  'users.delete': {
    label: 'admin.policyActionDeleteUser',
    description: 'admin.policyDescDeleteUser'
  },
  'users.password': {
    label: 'admin.policyActionResetUserPassword',
    description: 'admin.policyDescResetUserPassword'
  },
  'users.roles': {
    label: 'admin.policyActionAssignUserRoles',
    description: 'admin.policyDescAssignUserRoles'
  },
  'roles.list': {
    label: 'admin.policyActionListRoles',
    description: 'admin.policyDescListRoles'
  },
  'roles.create': {
    label: 'admin.policyActionCreateRole',
    description: 'admin.policyDescCreateRole'
  },
  'roles.update': {
    label: 'admin.policyActionUpdateRole',
    description: 'admin.policyDescUpdateRole'
  },
  'roles.delete': {
    label: 'admin.policyActionDeleteRole',
    description: 'admin.policyDescDeleteRole'
  },
  'roles.policies.get': {
    label: 'admin.policyActionViewRolePolicies',
    description: 'admin.policyDescViewRolePolicies'
  },
  'roles.policies.set': {
    label: 'admin.policyActionSaveRolePolicies',
    description: 'admin.policyDescSaveRolePolicies'
  },
  'configs.list': {
    label: 'admin.policyActionListConfigs',
    description: 'admin.policyDescListConfigs'
  },
  'configs.save': {
    label: 'admin.policyActionSaveConfigs',
    description: 'admin.policyDescSaveConfigs'
  },
  'profile.view': {
    label: 'admin.policyActionViewProfile',
    description: 'admin.policyDescViewProfile'
  },
  'profile.update': {
    label: 'admin.policyActionUpdateProfile',
    description: 'admin.policyDescUpdateProfile'
  },
  'profile.password': {
    label: 'admin.policyActionResetOwnPassword',
    description: 'admin.policyDescResetOwnPassword'
  },
  'profile.phone': {
    label: 'admin.policyActionChangePhone',
    description: 'admin.policyDescChangePhone'
  },
  'profile.avatar': {
    label: 'admin.policyActionUploadAvatar',
    description: 'admin.policyDescUploadAvatar'
  }
};

function translateKnownKey(t: TranslateFn, key?: string, fallback = ''): string {
  if (!key) {
    return fallback;
  }
  const translated = t(key);
  return translated === key ? fallback : translated;
}

export function createEmptyUserFormValues(): UserFormValues {
  return {
    username: '',
    phone: '',
    password: '',
    email: '',
    avatar_url: '',
    avatar_file_id: '',
    signature: '',
    gender: '',
    age: undefined,
    is_active: true
  };
}

export function mapUserToFormValues(user: AuthUser): UserFormValues {
  return {
    username: user.username,
    phone: user.phone,
    password: '',
    email: user.email ?? '',
    avatar_url: user.avatar_url ?? '',
    avatar_file_id: '',
    signature: user.signature ?? '',
    gender: user.gender ?? '',
    age: user.age,
    is_active: user.is_active ?? true
  };
}

export function createEmptyPasswordFormValues(): PasswordFormValues {
  return {
    password: ''
  };
}

export function createEmptyRoleFormValues(): RoleFormValues {
  return {
    name: '',
    display_name: '',
    description: ''
  };
}

export function mapRoleToFormValues(role: RoleItem): RoleFormValues {
  return {
    name: role.name,
    display_name: role.display_name,
    description: role.description
  };
}

export function createEmptyConfigFormValues(): ConfigFormValues {
  return {
    config_group: '',
    config_key: '',
    config_val: '',
    remark: ''
  };
}

export function mapConfigToFormValues(config: SystemConfigItem): ConfigFormValues {
  return {
    config_group: config.config_group ?? '',
    config_key: config.config_key,
    config_val: config.config_val,
    remark: config.remark ?? ''
  };
}

export function formatDateTime(value?: string): string {
  if (!value) {
    return '-';
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return parsed.toLocaleString();
}

export function formatDurationMS(value?: number): string {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return '-';
  }
  return `${value} ms`;
}

export function getPolicyMethodColor(method: string): string {
  return POLICY_METHOD_COLORS[method] ?? 'default';
}

export function uniquePolicies(policies: RolePolicy[]): RolePolicy[] {
  const seen = new Set<string>();
  return policies.filter((policy) => {
    const key = `${policy.method} ${policy.path}`;
    if (seen.has(key)) {
      return false;
    }
    seen.add(key);
    return true;
  });
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function createPolicyPathPattern(path: string): RegExp {
  const escaped = escapeRegExp(path).replace(/\/:([a-zA-Z0-9_]+)/g, '/[^/]+').replace(/\\\*/g, '.*');
  return new RegExp(`^${escaped}$`);
}

export function matchesPolicyTemplate(policy: RolePolicy, template: PolicyTemplate): boolean {
  try {
    const methodMatched = new RegExp(`^${policy.method}$`).test(template.method);
    if (!methodMatched) {
      return false;
    }
  } catch {
    if (policy.method !== template.method) {
      return false;
    }
  }

  return createPolicyPathPattern(policy.path).test(template.path);
}

export function buildSectionAggregatePolicies(menuKey: string): RolePolicy[] {
  return SECTION_AGGREGATE_POLICIES[menuKey] ?? [];
}

export function getPolicyTemplateScopeKey(path: string): PolicyTemplateScopeKey {
  if (path.startsWith('/api/v1/admin/')) {
    return 'admin';
  }
  if (path.startsWith('/api/v1/user/')) {
    return 'user';
  }
  return 'other';
}

export function buildPoliciesFromSelection(selectedKeys: string[], sections: PolicyTemplateSection[]): RolePolicy[] {
  const selectedKeySet = new Set(selectedKeys);
  const policies: RolePolicy[] = [];

  sections.forEach((section) => {
    section.items
      .filter((item) => selectedKeySet.has(item.key))
      .forEach((item) => {
        policies.push({ path: item.path, method: item.method });
      });
  });

  return uniquePolicies(policies);
}

export function localizePolicyMenuLabel(menuKey: string, fallback: string | undefined, t: TranslateFn): string {
  return translateKnownKey(t, POLICY_MENU_I18N_KEYS[menuKey], fallback?.trim() || menuKey);
}

export function localizePolicyActionLabel(operationId: string, fallback: string | undefined, t: TranslateFn): string {
  return translateKnownKey(t, POLICY_OPERATION_I18N_KEYS[operationId]?.label, fallback?.trim() || operationId);
}

export function localizePolicyDescription(operationId: string, fallback: string | undefined, t: TranslateFn): string | undefined {
  const translated = translateKnownKey(t, POLICY_OPERATION_I18N_KEYS[operationId]?.description, fallback?.trim() || '');
  return translated || undefined;
}

export function localizePolicyTemplate(template: PolicyTemplate, t: TranslateFn): PolicyTemplate {
  return {
    ...template,
    menuLabel: localizePolicyMenuLabel(template.menuKey, template.menuLabel, t),
    actionLabel: localizePolicyActionLabel(template.key, template.actionLabel, t),
    description: localizePolicyDescription(template.key, template.description, t)
  };
}
