import { ISceneLoaderAsyncResult, Material, TransformNode } from '@babylonjs/core';

export function getMaterial(model: ISceneLoaderAsyncResult, name: string | RegExp): Material | undefined {
	if (typeof name === 'string') {
		return model.meshes
			.map(mesh => mesh.material)
			.find(material => material && name === material!.name)!;
	} else {
		return model.meshes
			.map(mesh => mesh.material)
			.find(material => material && name.test(material!.name))!;
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
