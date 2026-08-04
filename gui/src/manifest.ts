type AssetType = 'core' | 'dlc' | 'prerequisite';

export interface Asset {
	name: string;
	type: AssetType;
	source: string;
	destination: string;
	dependencies: Array<string>;
	packedSize: number;
	unpackedSize: number;
	automatic: boolean;
	files: Array<File>;
}

interface File {
	checksum: string;
	file: string;
}

export interface Manifest {
	assets: Array<Asset>;
	version: number;
}
