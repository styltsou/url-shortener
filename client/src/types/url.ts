export interface Tag {
	id: string;
	name: string;
}

export interface Url {
	id: string;
	originalUrl: string;
	shortCode: string;
	createdAt: Date;
	expiresAt: Date | null;
	tags: Tag[];
	isActive?: boolean; // Active/inactive status
}
