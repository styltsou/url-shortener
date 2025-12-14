import { useRef } from "react";

/**
 * A hook that returns a debounced version of a callback function.
 * 
 * @param callback - The function to debounce
 * @param delay - The delay in milliseconds (default: 300)
 * @returns A debounced version of the callback
 * 
 * @example
 * ```tsx
 * const debouncedSearch = useDebouncedCallback((value: string) => {
 *   updateSearchParams({ query: value });
 * }, 300);
 * 
 * <Input onChange={(e) => debouncedSearch(e.target.value)} />
 * ```
 */
export function useDebouncedCallback<T extends (...args: any[]) => any>(
	callback: T,
	delay: number = 300
): T {
	const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	// React Compiler automatically memoizes this function
	const debouncedCallback = ((...args: Parameters<T>) => {
		if (timeoutRef.current) {
			clearTimeout(timeoutRef.current);
		}

		timeoutRef.current = setTimeout(() => {
			callback(...args);
		}, delay);
	}) as T;

	return debouncedCallback;
}

