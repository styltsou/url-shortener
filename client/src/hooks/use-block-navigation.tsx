import { createContext, useContext, useEffect, useRef, useCallback, useState, ReactNode } from "react";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface BlockNavigationState {
	shouldBlock: boolean;
	title?: string;
	message?: string;
	confirmButtonLabel?: string;
	cancelButtonLabel?: string;
	blockBrowserNavigation?: boolean;
	blockBackForward?: boolean;
	onConfirm?: () => void;
	onCancel?: () => void;
}

interface NavigationBlockerContextValue {
	registerBlocker: (id: string, state: BlockNavigationState) => void;
	unregisterBlocker: (id: string) => void;
}

const NavigationBlockerContext = createContext<NavigationBlockerContextValue | null>(null);

export interface NavigationBlockerProviderProps {
	children: ReactNode;
}

/**
 * Provider component that manages navigation blocking state and renders the dialog.
 * Wrap your app with this provider to enable navigation blocking functionality.
 *
 * @example
 * ```tsx
 * function App() {
 *   return (
 *     <NavigationBlockerProvider>
 *       <YourApp />
 *     </NavigationBlockerProvider>
 *   );
 * }
 * ```
 */
export function NavigationBlockerProvider({ children }: NavigationBlockerProviderProps) {
	const [blockers, setBlockers] = useState<Map<string, BlockNavigationState>>(new Map());
	const [showDialog, setShowDialog] = useState(false);
	const [currentBlocker, setCurrentBlocker] = useState<BlockNavigationState | null>(null);
	const shouldGoBackRef = useRef(false);

	// Get the most relevant blocker (first one that should block)
	const activeBlocker = Array.from(blockers.values()).find((b) => b.shouldBlock);

	// Block browser navigation (refresh, tab close, etc.)
	useEffect(() => {
		if (!activeBlocker?.blockBrowserNavigation || !activeBlocker?.shouldBlock) {
			return;
		}

		const handleBeforeUnload = (e: BeforeUnloadEvent) => {
			e.preventDefault();
			e.returnValue = "";
			return "";
		};

		window.addEventListener("beforeunload", handleBeforeUnload);

		return () => {
			window.removeEventListener("beforeunload", handleBeforeUnload);
		};
	}, [activeBlocker]);

	// Block browser back/forward navigation
	useEffect(() => {
		if (!activeBlocker?.blockBackForward || !activeBlocker?.shouldBlock) {
			return;
		}

		const currentUrl = window.location.href;
		const currentState = window.history.state;

		const handlePopState = (e: PopStateEvent) => {
			if (activeBlocker?.shouldBlock) {
				window.history.pushState(currentState, "", currentUrl);
				setShowDialog(true);
				setCurrentBlocker(activeBlocker);
				shouldGoBackRef.current = true;
			}
		};

		window.history.pushState(null, "", window.location.href);
		window.addEventListener("popstate", handlePopState);

		return () => {
			window.removeEventListener("popstate", handlePopState);
		};
	}, [activeBlocker]);

	const registerBlocker = useCallback((id: string, state: BlockNavigationState) => {
		setBlockers((prev) => {
			const next = new Map(prev);
			next.set(id, state);
			return next;
		});
	}, []);

	const unregisterBlocker = useCallback((id: string) => {
		setBlockers((prev) => {
			const next = new Map(prev);
			next.delete(id);
			return next;
		});
	}, []);

	const handleConfirm = useCallback(() => {
		setShowDialog(false);
		currentBlocker?.onConfirm?.();

		if (shouldGoBackRef.current) {
			window.history.back();
			shouldGoBackRef.current = false;
		}

		setCurrentBlocker(null);
	}, [currentBlocker]);

	const handleCancel = useCallback(() => {
		setShowDialog(false);
		shouldGoBackRef.current = false;
		currentBlocker?.onCancel?.();
		setCurrentBlocker(null);
	}, [currentBlocker]);

	const contextValue: NavigationBlockerContextValue = {
		registerBlocker,
		unregisterBlocker,
	};

	return (
		<NavigationBlockerContext.Provider value={contextValue}>
			{children}
			<AlertDialog open={showDialog} onOpenChange={(open) => {
				if (!open) {
					handleCancel();
				}
			}}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>
							{currentBlocker?.title || "Confirm Navigation"}
						</AlertDialogTitle>
						<AlertDialogDescription>
							{currentBlocker?.message || "Are you sure you want to leave? Your changes may be lost."}
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel onClick={handleCancel}>
							{currentBlocker?.cancelButtonLabel || "Cancel"}
						</AlertDialogCancel>
						<AlertDialogAction onClick={handleConfirm}>
							{currentBlocker?.confirmButtonLabel || "Leave"}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</NavigationBlockerContext.Provider>
	);
}

export interface UseBlockNavigationOptions {
	/**
	 * Whether navigation should be blocked
	 */
	shouldBlock: boolean;
	/**
	 * Title for the navigation confirmation dialog
	 * @default "Confirm Navigation"
	 */
	title?: string;
	/**
	 * Description/message for the navigation confirmation dialog
	 * @default "Are you sure you want to leave? Your changes may be lost."
	 */
	message?: string;
	/**
	 * Label for the "Confirm" button (allows navigation)
	 * @default "Leave"
	 */
	confirmButtonLabel?: string;
	/**
	 * Label for the "Cancel" button (blocks navigation)
	 * @default "Cancel"
	 */
	cancelButtonLabel?: string;
	/**
	 * Whether to block browser navigation (refresh, tab close, etc.)
	 * @default true
	 */
	blockBrowserNavigation?: boolean;
	/**
	 * Whether to block browser back/forward button navigation
	 * @default true
	 */
	blockBackForward?: boolean;
	/**
	 * Callback fired when user confirms they want to navigate
	 */
	onConfirm?: () => void;
	/**
	 * Callback fired when user cancels navigation
	 */
	onCancel?: () => void;
}

/**
 * A hook that blocks browser-initiated navigation (back/forward buttons, refresh, close)
 * based on a condition and shows a customizable alert dialog.
 *
 * This hook only blocks browser-initiated navigation:
 * - Browser back/forward buttons
 * - Browser refresh/close
 *
 * Programmatic navigation (navigate() calls, Link clicks) is NOT blocked.
 *
 * Requires NavigationBlockerProvider to be wrapped around your app.
 *
 * @example
 * ```tsx
 * // In your root component:
 * <NavigationBlockerProvider>
 *   <App />
 * </NavigationBlockerProvider>
 *
 * // In any component:
 * const [hasChanges, setHasChanges] = useState(false);
 * useBlockNavigation({
 *   shouldBlock: hasChanges,
 *   title: "Unsaved Changes",
 *   message: "You have unsaved edits. Are you sure you want to leave?",
 * });
 *
 * // Programmatic navigation works normally
 * const handleNavigation = () => {
 *   navigate({ to: "/other-page" }); // Always works, no blocking
 * };
 * ```
 */
export function useBlockNavigation(options: UseBlockNavigationOptions) {
	const context = useContext(NavigationBlockerContext);
	const blockerIdRef = useRef<string>(`blocker-${Math.random().toString(36).substr(2, 9)}`);

	if (!context) {
		throw new Error(
			"useBlockNavigation must be used within a NavigationBlockerProvider. " +
			"Wrap your app with <NavigationBlockerProvider>."
		);
	}

	const {
		shouldBlock,
		title = "Confirm Navigation",
		message = "Are you sure you want to leave? Your changes may be lost.",
		confirmButtonLabel = "Leave",
		cancelButtonLabel = "Cancel",
		blockBrowserNavigation = true,
		blockBackForward = true,
		onConfirm,
		onCancel,
	} = options;

	// Register/unregister blocker when options change
	useEffect(() => {
		context.registerBlocker(blockerIdRef.current, {
			shouldBlock,
			title,
			message,
			confirmButtonLabel,
			cancelButtonLabel,
			blockBrowserNavigation,
			blockBackForward,
			onConfirm,
			onCancel,
		});

		return () => {
			context.unregisterBlocker(blockerIdRef.current);
		};
	}, [
		context,
		shouldBlock,
		title,
		message,
		confirmButtonLabel,
		cancelButtonLabel,
		blockBrowserNavigation,
		blockBackForward,
		onConfirm,
		onCancel,
	]);
}
