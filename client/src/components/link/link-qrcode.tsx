import { useState } from "react";
import { Download } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { getApiBaseUrl } from "@/lib/env";

interface LinkQRCodeProps {
	shortcode: string;
}

export function LinkQRCode({ shortcode }: LinkQRCodeProps) {
	const [size] = useState(256);
	const apiBase = getApiBaseUrl();
	const qrUrl = `${apiBase}/api/v1/links/${shortcode}/qrcode?size=${size}`;

	const handleDownload = async () => {
		try {
			const response = await fetch(qrUrl);
			const blob = await response.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = `qr-${shortcode}.png`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch {
			// Silently fail
		}
	};

	return (
		<Card>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
					QR Code
				</CardTitle>
			</CardHeader>
			<CardContent className='flex flex-col items-center gap-4'>
				<img
					src={qrUrl}
					alt={`QR code for link ${shortcode}`}
					className='rounded-lg'
					width={size}
					height={size}
				/>
				<Button variant='outline' size='sm' onClick={handleDownload}>
					<Download className='w-4 h-4 mr-2' />
					Download
				</Button>
			</CardContent>
		</Card>
	);
}
