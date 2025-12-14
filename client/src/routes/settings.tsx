import { createFileRoute } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { LoadingState } from "@/components/shared/loading-state";

export const Route = createFileRoute("/settings")({
	component: SettingsPage,
});

function SettingsPage() {
	const { isSignedIn, isLoaded } = useAuth();

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to='/login' />;
	}

	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-6xl mx-auto'>
				<div className='mb-8'>
					<h1 className='text-3xl font-bold text-foreground'>Settings</h1>
					<p className='text-muted-foreground mt-2'>
						Manage your application settings and preferences.
					</p>
				</div>
				<div className='grid gap-6'>
					<div className='border border-border rounded-lg p-8 text-center'>
						<p className='text-muted-foreground'>
							Settings content coming soon...
						</p>
					</div>
				</div>
			</div>
		</main>
	);
}
