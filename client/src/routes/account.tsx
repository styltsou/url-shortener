import { createFileRoute, Navigate } from "@tanstack/react-router";
import { useAuth, useUser } from "@clerk/clerk-react";
import { LoadingState } from "@/components/shared/loading-state";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Mail, Calendar, Shield, ExternalLink } from "lucide-react";
import { DATETIME_FORMAT_OPTIONS } from "@/lib/constants";

export const Route = createFileRoute("/account")({
	component: AccountPage,
});

function AccountPage() {
	const { isSignedIn, isLoaded } = useAuth();
	const { user } = useUser();

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to='/login' />;
	}

	const primaryEmail = user?.primaryEmailAddress?.emailAddress;
	const createdAt = user?.createdAt
		? new Date(user.createdAt).toLocaleDateString("en-US", DATETIME_FORMAT_OPTIONS)
		: null;

	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-2xl mx-auto'>
				<div className='mb-8'>
					<h1 className='text-3xl font-bold text-foreground'>Account</h1>
					<p className='text-muted-foreground mt-2'>
						Manage your account information and preferences.
					</p>
				</div>

				<div className='space-y-6'>
					<Card>
						<CardHeader>
							<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
								Profile
							</CardTitle>
						</CardHeader>
						<CardContent className='flex items-center gap-4'>
							<Avatar className='h-16 w-16'>
								<AvatarImage src={user?.imageUrl} />
								<AvatarFallback>
									{user?.firstName?.[0]}
									{user?.lastName?.[0]}
								</AvatarFallback>
							</Avatar>
							<div>
								<p className='text-lg font-semibold text-foreground'>
									{user?.fullName}
								</p>
								<p className='text-sm text-muted-foreground'>{primaryEmail}</p>
							</div>
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
								Details
							</CardTitle>
						</CardHeader>
						<CardContent className='space-y-4'>
							<div className='flex items-center gap-3'>
								<Mail className='h-4 w-4 text-muted-foreground shrink-0' />
								<div>
									<p className='text-sm font-medium text-foreground'>Email</p>
									<p className='text-sm text-muted-foreground'>
										{primaryEmail}
									</p>
								</div>
							</div>
							<div className='flex items-center gap-3'>
								<Calendar className='h-4 w-4 text-muted-foreground shrink-0' />
								<div>
									<p className='text-sm font-medium text-foreground'>
										Member since
									</p>
									<p className='text-sm text-muted-foreground'>{createdAt}</p>
								</div>
							</div>
							<div className='flex items-center gap-3'>
								<Shield className='h-4 w-4 text-muted-foreground shrink-0' />
								<div>
									<p className='text-sm font-medium text-foreground'>
										Authentication
									</p>
									<div className='flex gap-1.5 mt-0.5'>
										{user?.externalAccounts?.map((account) => (
											<Badge
												key={account.id}
												variant='secondary'
												className='text-xs'
											>
												{account.provider}
											</Badge>
										))}
										{(!user?.externalAccounts ||
											user.externalAccounts.length === 0) && (
											<Badge variant='secondary' className='text-xs'>
												Email
											</Badge>
										)}
									</div>
								</div>
							</div>
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
								Manage Account
							</CardTitle>
						</CardHeader>
						<CardContent>
							<p className='text-sm text-muted-foreground mb-4'>
								Update your profile, security settings, and connected accounts
								through Clerk's account portal.
							</p>
							<a
								href='https://clerk.com/docs'
								target='_blank'
								rel='noopener noreferrer'
								className='inline-flex items-center gap-2 text-sm text-primary hover:underline'
							>
								<ExternalLink className='h-4 w-4' />
								Manage in Clerk
							</a>
						</CardContent>
					</Card>
				</div>
			</div>
		</main>
	);
}
