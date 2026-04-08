import { Navigate } from '@/lib/router';
import { adminRoutePaths } from '@/constants/routes';

export default function AdminIndexPage() {
  return <Navigate replace to={adminRoutePaths.home} />;
}
