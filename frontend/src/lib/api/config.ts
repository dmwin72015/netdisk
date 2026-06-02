import { api } from './client';

export const DeviceType = { Web: 'web', PC: 'pc', Mobile: 'mobile' } as const;
export type DeviceType = (typeof DeviceType)[keyof typeof DeviceType];

export type ClientConfig = {
	device: DeviceType;
	configs: Record<string, unknown>;
};

export async function getClientConfig(device: DeviceType = DeviceType.Web): Promise<ClientConfig> {
	return api<ClientConfig>(`/api/v1/config?device=${device}`);
}
