import { createFileRoute } from '@tanstack/react-router'
import { AuthenticateWithRedirectCallback } from '@clerk/clerk-react'

export const Route = createFileRoute('/sso-callback')({
  component: SSOCallbackPage,
})

function SSOCallbackPage() {
  // AuthenticateWithRedirectCallback automatically handles the OAuth callback
  // and redirects to the appropriate URL. This component renders nothing visible
  // and redirects immediately, so the user experience is seamless.
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <AuthenticateWithRedirectCallback
        afterSignInUrl={import.meta.env.VITE_CLERK_AFTER_SIGN_IN_URL || '/'}
        afterSignUpUrl={import.meta.env.VITE_CLERK_AFTER_SIGN_UP_URL || '/'}
      />
    </div>
  )
}
