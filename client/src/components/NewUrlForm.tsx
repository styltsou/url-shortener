import { useState } from "react";
import {
	Link as LinkIcon,
	ArrowRight,
	ChevronRight,
	Loader2,
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
	InputGroup,
	InputGroupText,
	InputGroupInput,
} from "@/components/ui/input-group";

interface NewUrlFormProps {
	onShorten: (
		originalUrl: string,
		customCode?: string,
		expirationDate?: string
	) => Promise<void>;
	isLoading: boolean;
}

export function NewUrlForm({ onShorten, isLoading }: NewUrlFormProps) {
	const [originalUrl, setOriginalUrl] = useState("");
	const [showOptions, setShowOptions] = useState(false);
	const [customCode, setCustomCode] = useState("");
	const [expirationDate, setExpirationDate] = useState("");

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!originalUrl) return;
		await onShorten(originalUrl, customCode, expirationDate);
		setOriginalUrl("");
		setCustomCode("");
		setExpirationDate("");
		setShowOptions(false);
	};

	return (
		<div className='w-full max-w-3xl mx-auto mb-12'>
			<div className='text-center mb-8 space-y-2'>
				<h2 className='text-4xl font-extrabold text-foreground tracking-tight'>
					Shorten your links
				</h2>
				<p className='text-muted-foreground text-lg'>
					Detailed analytics and custom branding included.
				</p>
			</div>

			<form onSubmit={handleSubmit} className='w-full'>
				<div className='flex flex-col sm:flex-row gap-3'>
					<div className='relative flex-1'>
						<div className='absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground'>
							<LinkIcon className='w-5 h-5' />
						</div>
						<Input
							type='url'
							value={originalUrl}
							onChange={(e) => setOriginalUrl(e.target.value)}
							placeholder='Paste a long URL here...'
							required
							className='h-12 pl-12 text-base md:text-lg w-full transition-all focus-visible:ring-2 focus-visible:ring-offset-2'
						/>
					</div>
					<Button
						type='submit'
						disabled={isLoading}
						className='h-12 px-8 text-base font-semibold shrink-0 relative transition-colors'
					>
						<span
							className={`flex items-center ${
								isLoading ? "opacity-0" : "opacity-100"
							}`}
						>
							Shorten <ArrowRight className='w-5 h-5 ml-2' />
						</span>
						{isLoading && (
							<div className='absolute inset-0 flex items-center justify-center'>
								<Loader2 className='w-5 h-5 animate-spin' />
							</div>
						)}
					</Button>
				</div>

				<div className='mt-3 text-center'>
					<Button
						type='button'
						variant='ghost'
						onClick={() => setShowOptions(!showOptions)}
						className='text-sm font-medium text-muted-foreground hover:text-foreground'
					>
						{showOptions ? "Hide Options" : "Show customization options"}
						<ChevronRight
							className={`w-3 h-3 transition-transform ${
								showOptions ? "rotate-90" : "rotate-0"
							}`}
						/>
					</Button>
				</div>

				<div
					className={`overflow-hidden transition-all duration-300 ease-in-out ${
						showOptions ? "max-h-40 opacity-100 mt-6" : "max-h-0 opacity-0"
					}`}
				>
					<div className='grid grid-cols-1 md:grid-cols-2 gap-6 px-1'>
						<div>
							<label className='block text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2'>
								Custom Alias
							</label>
							<InputGroup>
								<InputGroupText className='text-primary'>
									short.ly/
								</InputGroupText>
								<InputGroupInput
									type='text'
									value={customCode}
									onChange={(e) => setCustomCode(e.target.value)}
									placeholder='alias'
								/>
							</InputGroup>
						</div>
						<div>
							<label className='block text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2'>
								Expiration Date
							</label>
							<Input
								type='date'
								value={expirationDate}
								onChange={(e) => setExpirationDate(e.target.value)}
								className='bg-transparent'
							/>
						</div>
					</div>
				</div>
			</form>
		</div>
	);
}
