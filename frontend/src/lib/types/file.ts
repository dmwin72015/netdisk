export interface NormalizedFile {
	id: string;
	name: string;
	isDir: boolean;
	size: number;
	mimeType: string | null;
	fileCategory: string;
	isStarred: boolean;
	updatedAt: string;
}
