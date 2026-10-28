import type { AbstractMesh, ISceneLoaderAsyncResult, Material, TransformNode } from '@babylonjs/core';

export function getMaterials(model: ISceneLoaderAsyncResult, name: string | RegExp): Array<Material> {
	if (typeof name === 'string') {
		return model.meshes
			.map(mesh => mesh.material)
			.filter(material => name === material?.name) as Array<Material>;
	} else {
		return model.meshes
			.map(mesh => mesh.material)
			.filter(material => material && name.test(material!.name)) as Array<Material>;
	}
}

export function getMeshes(model: ISceneLoaderAsyncResult, name: string | RegExp): Array<AbstractMesh> {
	if (typeof name === 'string') {
		return model.meshes
			.filter(mesh => name === mesh.name);
	} else {
		return model.meshes
			.filter(mesh => name.test(mesh.name));
	}
}

export function getNode(model: ISceneLoaderAsyncResult, name: string | RegExp): TransformNode | undefined {
	if (typeof name === 'string') {
		return model.transformNodes
			.find(node => name === node.name);
	} else {
		return model.transformNodes
			.find(node => name.test(node.name));
	}
}
