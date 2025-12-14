import { createFileRoute } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { useState } from "react";
import { MessageSquare, Bug, Lightbulb, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { LoadingState } from "@/components/shared/loading-state";
import { toast } from "sonner";

export const Route = createFileRoute("/feedback")({
	component: FeedbackPage,
});

type FeedbackType = "feedback" | "bug" | "feature";

function FeedbackPage() {
	const { isSignedIn, isLoaded } = useAuth();
	const [type, setType] = useState<FeedbackType>("feedback");
	const [message, setMessage] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to='/login' />;
	}

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!message.trim()) {
			toast.error("Please enter your feedback");
			return;
		}

		setIsSubmitting(true);

		// TODO: Replace with actual API call
		// For now, simulate API call
		setTimeout(() => {
			setIsSubmitting(false);
			toast.success("Thank you for your feedback!");
			setMessage("");
			setType("feedback");
		}, 1000);
	};

	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-2xl mx-auto'>
				<div className='mb-8'>
					<h1 className='text-3xl font-bold text-foreground mb-2'>
						Share Your Feedback
					</h1>
					<p className='text-muted-foreground'>
						We'd love to hear from you! Share your thoughts, report bugs, or
						suggest new features.
					</p>
				</div>

				<form onSubmit={handleSubmit} className='space-y-4'>
					<div className='space-y-2'>
						<Label htmlFor='type'>Type</Label>
						<Select
							value={type}
							onValueChange={(value) => setType(value as FeedbackType)}
						>
							<SelectTrigger id='type'>
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value='feedback'>
									<div className='flex items-center gap-2'>
										<MessageSquare className='w-4 h-4' />
										General Feedback
									</div>
								</SelectItem>
								<SelectItem value='bug'>
									<div className='flex items-center gap-2'>
										<Bug className='w-4 h-4' />
										Bug Report
									</div>
								</SelectItem>
								<SelectItem value='feature'>
									<div className='flex items-center gap-2'>
										<Lightbulb className='w-4 h-4' />
										Feature Request
									</div>
								</SelectItem>
							</SelectContent>
						</Select>
					</div>

					<div className='space-y-2'>
						<Label htmlFor='message'>
							{type === "bug"
								? "Describe the bug"
								: type === "feature"
								? "Describe your feature idea"
								: "Your feedback"}
						</Label>
						<Textarea
							id='message'
							placeholder={
								type === "bug"
									? "Please describe what happened, what you expected to happen, and steps to reproduce the bug..."
									: type === "feature"
									? "Tell us about the feature you'd like to see..."
									: "Share your thoughts, suggestions, or any other feedback..."
							}
							value={message}
							onChange={(e) => setMessage(e.target.value)}
							rows={8}
							className='resize-none'
							required
						/>
					</div>

					<div className='flex justify-end'>
						<Button type='submit' disabled={isSubmitting || !message.trim()}>
							{isSubmitting ? (
								<>
									<span className='mr-2'>Sending...</span>
								</>
							) : (
								<>
									<Send className='w-4 h-4 mr-2' />
									Send Feedback
								</>
							)}
						</Button>
					</div>
				</form>
			</div>
		</main>
	);
}
