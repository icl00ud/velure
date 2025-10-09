import { Navigate } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireAuth?: boolean;
  redirectTo?: string;
}

export const ProtectedRoute = ({ 
  children, 
  requireAuth = true, 
  redirectTo = "/" 
}: ProtectedRouteProps) => {
  const { isAuthenticated, isInitializing } = useAuth();

  if (isInitializing) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (requireAuth && !isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (!requireAuth && isAuthenticated) {
    return <Navigate to={redirectTo} replace />;
  }

  return <>{children}</>;
};
