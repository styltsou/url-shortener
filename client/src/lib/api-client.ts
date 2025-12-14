import { getApiBaseUrl } from "./env";

export interface ApiSuccessResponse<T> {
	data: T;
}

export interface ApiErrorResponse {
	error: {
		code?: string;
		message?: string;
		detail?: string;
	};
}

export type ApiResponse<T> = ApiSuccessResponse<T> | ApiErrorResponse;

/**
 * Custom error class for API errors
 */
export class ApiError extends Error {
	constructor(public status: number, public code?: string, message?: string) {
		super(message || `HTTP error! status: ${status}`);
		this.name = "ApiError";
	}
}

/**
 * Base fetch wrapper with auth
 */
async function apiFetch<T>(
	endpoint: string,
	options: RequestInit = {},
	token: string | null
): Promise<ApiSuccessResponse<T>> {
	const baseUrl = getApiBaseUrl();
	const url = `${baseUrl}${endpoint}`;

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
		let errorMessage = `HTTP error! status: ${response.status}`;
		let errorCode: string | undefined;

		try {
			const error: ApiErrorResponse = await response.json();
			errorMessage =
				error.error?.detail || error.error?.message || errorMessage;
			errorCode = error.error?.code;
		} catch {
			// If response is not JSON, use default error message
		}

		throw new ApiError(response.status, errorCode, errorMessage);
	}

	// Handle 204 No Content responses
	if (response.status === 204) {
		return {} as ApiSuccessResponse<T>;
	}

	return response.json();
}

/**
 * API client for making authenticated requests
 * All methods require a token to be passed
 */
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
