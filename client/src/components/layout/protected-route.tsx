import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { LoadingState } from "@/components/shared/loading-state";

interface ProtectedRouteProps {
	children: ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
	const { isSignedIn, isLoaded } = useAuth();

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to="/login" />;
	}

	return <>{children}</>;
}
