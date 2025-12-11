import { useAuth } from "@clerk/clerk-react";

const API_BASE_URL =
	import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

export interface ApiSuccessResponse<T> {
	data: T;
}

export interface ApiErrorResponse {
	error: {
		code: string;
		message: string;
	};
}

export type ApiResponse<T> = ApiSuccessResponse<T> | ApiErrorResponse;

// Helper to get auth token from Clerk
export async function getAuthToken(): Promise<string | null> {
	// This will be called from React components where useAuth is available
	// For now, we'll pass the token directly
	return null;
}

// Base fetch wrapper with auth
async function apiFetch<T>(
	endpoint: string,
	options: RequestInit = {},
	token: string | null
): Promise<ApiSuccessResponse<T>> {
	const url = `${API_BASE_URL}${endpoint}`;

	const headers: HeadersInit = {
		"Content-Type": "application/json",
		...options.headers,
	};

	if (token) {
		headers["Authorization"] = `Bearer ${token}`;
	}

	const response = await fetch(url, {
		...options,
		headers,
	});

	if (!response.ok) {
		const error: ApiErrorResponse = await response.json();
		throw new Error(
			error.error?.message || `HTTP error! status: ${response.status}`
		);
	}

	return response.json();
}

export const apiClient = {
	async get<T>(
		endpoint: string,
		token: string | null
	): Promise<ApiSuccessResponse<T>> {
		return apiFetch<T>(endpoint, { method: "GET" }, token);
	},

	async post<T>(
		endpoint: string,
		body: unknown,
		token: string | null
	): Promise<ApiSuccessResponse<T>> {
		return apiFetch<T>(
			endpoint,
			{
				method: "POST",
				body: JSON.stringify(body),
			},
			token
		);
	},

	async patch<T>(
		endpoint: string,
		body: unknown,
		token: string | null
	): Promise<ApiSuccessResponse<T>> {
		return apiFetch<T>(
			endpoint,
			{
				method: "PATCH",
				body: JSON.stringify(body),
			},
			token
		);
	},

	async delete(endpoint: string, token: string | null): Promise<void> {
		await apiFetch(endpoint, { method: "DELETE" }, token);
	},
};
