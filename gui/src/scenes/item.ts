import { Vector3 } from '@babylonjs/core';
import { cdn } from '../config';
import itemData from './items.json';

export enum ItemType {
	Vehicle = "Vehicle"
}

export class Item {
	public readonly id: number;
	public readonly type: ItemType;
	public readonly category: string;
	public readonly subcategory?: string;
	public readonly internalName: string;
	public readonly displayName: string;
	public readonly description: string;
	public readonly model: Model;
	public readonly tags: Array<string>;

	constructor(
		id: number,
		type: ItemType,
		category: string,
		internalName: string,
		displayName: string,
		description: string,
		model: Model,
		tags?: Array<string>,
		subcategory?: string
	) {
		this.id = id;
		this.type = type;
		this.category = category;
		this.subcategory = subcategory;
		this.internalName = internalName;
		this.displayName = displayName;
		this.description = description;
		this.model = model;
		this.tags = tags ?? [];
	}

	public static fromJSON(data: any): Item {
		return new Item(
			data.id,
			data.type,
			data.category,
			data.internalName,
			data.displayName,
			data.description,
			data.model && Model.fromJSON(data.model),
			data.tags,
			data.subcategory
		);
	}
}

export class Model {
	public readonly internalName: string;
	public readonly root: string;
	public readonly scale: number;
	public readonly rotation: Vector3;
	public readonly position: Vector3;

	constructor(
		internalName: string,
		root: string,
		scale: number,
		rotation: Vector3,
		position: Vector3
	) {
		this.internalName = internalName;
		this.root = root || 'Root';
		this.scale = scale || 1;
		this.rotation = rotation || new Vector3(0, 0, 0);
		this.position = position || new Vector3(0, 0, 0);
	}

	public get directory(): URL {
		return cdn(`tar/streaming/models/${this.internalName}/`);
	}

	public get file(): string {
		return 'model.gltf';
	}

	public get model(): URL {
		return cdn(`tar/streaming/models/${this.internalName}/${this.file}`);
	}

	public static fromJSON(data: any): Model {
		return new Model(
			data.internalName,
			data.root,
			data.scale,
			data.rotation && new Vector3(data.rotation.x, data.rotation.y, data.rotation.z),
			data.position && new Vector3(data.position.x, data.position.y, data.position.z)
		);
	}
}

export const items = {
	/** All items. */
	all: itemData.map(Item.fromJSON),
	/** Get an {@link Item} by its ID or internal name. */
	get: (nameOrId: number | string): Item | undefined => {
		return items.all.find(item => {
			if (typeof nameOrId === 'number') return item.id === nameOrId;
			if (typeof nameOrId === 'string') return item.internalName === nameOrId;
		});
	}
};
