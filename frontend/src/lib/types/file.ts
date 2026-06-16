export interface NormalizedFile {
	id: string;
	name: string;
	isDir: boolean;
	isLocked: boolean;
	size: number;
	mimeType: string | null;
	fileCategory: string;
	isStarred: boolean;
	isSystem: boolean;
	systemKind?: string;
	createdAt: string;
	updatedAt: string;
}
