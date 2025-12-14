/**
 * Environment variable validation and access
 */

/**
 * Validates that required environment variables are present
 * Throws an error if any required variables are missing
 */
export function validateEnv(): void {
	const clerkPubKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;

	if (!clerkPubKey) {
		throw new Error(
			"Missing required environment variable: VITE_CLERK_PUBLISHABLE_KEY"
		);
	}
}

/**
 * Gets the Clerk publishable key from environment variables
 * @throws Error if the key is not set
 */
export function getClerkPublishableKey(): string {
	const key = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;
	if (!key) {
		throw new Error(
			"Missing required environment variable: VITE_CLERK_PUBLISHABLE_KEY"
		);
	}
	return key;
}

/**
 * Gets the API base URL from environment variables
 * Falls back to localhost if not set
 */
export function getApiBaseUrl(): string {
	return import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
}

/**
 * Gets the short domain from environment variables
 * Falls back to link4.it if not set
 */
export function getShortDomain(): string {
	return import.meta.env.VITE_SHORT_DOMAIN || "link4.it";
}
