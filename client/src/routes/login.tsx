import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useSignIn, useAuth } from "@clerk/clerk-react";
import { Server } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useEffect, useState } from "react";

export const Route = createFileRoute("/login")({
	component: LoginPage,
});

function LoginPage() {
	const navigate = useNavigate();
	const { isSignedIn, isLoaded: authLoaded } = useAuth();
	const { signIn, isLoaded: signInLoaded } = useSignIn();
	const [loadingProvider, setLoadingProvider] = useState<
		"google" | "github" | null
	>(null);

	// Redirect if already signed in
	useEffect(() => {
		if (authLoaded && isSignedIn) {
			navigate({ to: import.meta.env.VITE_CLERK_AFTER_SIGN_IN_URL || "/" });
		}
	}, [authLoaded, isSignedIn, navigate]);

	const handleOAuth = async (provider: "google" | "github") => {
		if (!signIn || !signInLoaded || loadingProvider) return;

		setLoadingProvider(provider);
		try {
			// For OAuth, directly authenticate with redirect
			// redirectUrl should point to the callback route, not the final destination
			await signIn.authenticateWithRedirect({
				strategy: `oauth_${provider}`,
				redirectUrl: "/sso-callback",
				redirectUrlComplete:
					import.meta.env.VITE_CLERK_AFTER_SIGN_IN_URL || "/",
			});
			// Note: If redirect happens immediately, this code won't execute
			// But if there's a delay, the loading state will be visible
		} catch (err: any) {
			console.error("OAuth error:", err);
			setLoadingProvider(null);
			// You can add toast notifications here if needed
		}
	};

	if (!authLoaded || !signInLoaded) {
		return (
			<div className='flex min-h-screen items-center justify-center'>
				<div className='text-lg'>Loading...</div>
			</div>
		);
	}

	if (isSignedIn) {
		return null; // Will redirect via useEffect
	}

	return (
		<div className='flex min-h-screen items-center justify-center bg-background p-4'>
			<div className='w-full max-w-md space-y-8'>
				<div className='space-y-6 text-center'>
					<div className='flex justify-center'>
						<div className='bg-primary text-primary-foreground p-3 rounded-xl'>
							<Server className='w-8 h-8' />
						</div>
					</div>
					<div className='space-y-2'>
						<h1 className='text-2xl font-bold text-foreground'>
							Welcome to link4.it
						</h1>
						<p className='text-base text-muted-foreground'>
							Sign in to your account to continue
						</p>
					</div>
				</div>
				<div className='space-y-4'>
					<Button
						type='button'
						variant='outline'
						className='w-full h-11 text-base font-medium'
						onClick={() => handleOAuth("google")}
						disabled={loadingProvider !== null}
					>
						{loadingProvider === "google" ? (
							<Spinner className='mr-2 h-5 w-5' />
						) : (
							<svg className='mr-2 h-5 w-5' viewBox='0 0 24 24'>
								<path
									d='M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z'
									fill='#4285F4'
								/>
								<path
									d='M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z'
									fill='#34A853'
								/>
								<path
									d='M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z'
									fill='#FBBC05'
								/>
								<path
									d='M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z'
									fill='#EA4335'
								/>
							</svg>
						)}
						Continue with Google
					</Button>
					<Button
						type='button'
						variant='outline'
						className='w-full h-11 text-base font-medium'
						onClick={() => handleOAuth("github")}
						disabled={loadingProvider !== null}
					>
						{loadingProvider === "github" ? (
							<Spinner className='mr-2 h-5 w-5' />
						) : (
							<svg
								className='mr-2 h-5 w-5'
								fill='currentColor'
								viewBox='0 0 24 24'
							>
								<path
									fillRule='evenodd'
									d='M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z'
									clipRule='evenodd'
								/>
							</svg>
						)}
						Continue with GitHub
					</Button>
					<div className='relative'>
						<div className='absolute inset-0 flex items-center'>
							<span className='w-full border-t border-border' />
						</div>
						<div className='relative flex justify-center text-xs uppercase'>
							<span className='bg-background px-2 text-muted-foreground'>
								Secure authentication
							</span>
						</div>
					</div>
					<p className='text-xs text-center text-muted-foreground px-4'>
						By continuing, you agree to our Terms of Service and Privacy Policy.
						Your account will be created automatically on first sign in.
					</p>
				</div>
			</div>
		</div>
	);
}
