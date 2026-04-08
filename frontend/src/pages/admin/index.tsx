import { Navigate } from '@/lib/router';
import { routePaths } from '@/constants/routes';
import { getFirstAccessibleAdminPath } from '@/lib/access';
import { useAuthStore } from '@/stores';

export default function AdminIndexPage() {
  const user = useAuthStore((state) => state.user);

  return <Navigate replace to={getFirstAccessibleAdminPath(user) ?? routePaths.home} />;
}
