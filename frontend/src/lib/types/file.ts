export interface NormalizedFile {
	id: string;
	slug: string;
	name: string;
	isDir: boolean;
	/** Directory has a password set (server-side, persistent). */
	hasPassword: boolean;
	size: number;
	mimeType: string | null;
	fileCategory: string;
	isStarred: boolean;
	isSystem: boolean;
	systemKind?: string;
	createdAt: string;
	updatedAt: string;
	hashAlgo?: string;
	fileHash?: string;
}
