export type UploadPhase = 'hashing' | 'pending' | 'verifying' | 'uploading' | 'paused' | 'importing' | 'completed' | 'failed' | 'interrupted';

export type UploadItem = {
	uid: string;
	file: File;
	fileName: string;
	fileSize: number;
	fileHash: string;
	preHash: string;
	phase: UploadPhase;
	progress: number;
	hashProgress: number;
	uploadedBytes: number;
	speed: number;
	uploadSlug: string | null;
	sessionId: string | null;
	abortCtrl: AbortController | null;
	errorMsg: string | null;
};
