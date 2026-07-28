import { Color3, Material, PBRMaterial, Space, StandardMaterial, Texture, Vector3 } from '@babylonjs/core';
import { ISceneLoaderAsyncResult, TransformNode } from '@babylonjs/core';
import { items } from './item';
import { store } from '../store';
import { BaseScene } from './base';
import { getMaterial, getNode } from './gltf';
import { createMist } from './mist';

export function homeScene(canvas?: HTMLCanvasElement): () => void {
	const base = new BaseScene(canvas);
	base.camera.setTarget(new Vector3(0, -0.05, 0));
	base.camera.setPosition(new Vector3(0, 0.35, -1.6));
	base.lights[0].intensity = 0.5;

	// Create a GLB model.
	let glbModel: ISceneLoaderAsyncResult | undefined;
	let vehicleModel: TransformNode | undefined;
	let vehicleMaterial: Material | undefined;
	const glowMaterial = new StandardMaterial('glow', base.scene);
	const model = items.get(store.model)?.model;
	if (model) {
		base.loadModel(model.directory.href, model.file)
			.then(result => {
				glbModel = result;
				// console.log(result);
				vehicleModel = getNode(glbModel, model.root);
				const offset = new Vector3(0.1, 0.25, 0.6);
				vehicleModel?.translate(new Vector3(1, 0, 0), offset.x + model.position.x, Space.WORLD);
				vehicleModel?.translate(new Vector3(0, 1, 0), offset.y + model.position.y, Space.WORLD);
				vehicleModel?.translate(new Vector3(0, 0, -1), offset.z + model.position.z, Space.WORLD);
				vehicleModel?.rotate(new Vector3(1, 0, 0), model.rotation.x);
				vehicleModel?.rotate(new Vector3(0, 1, 0), model.rotation.y);
				vehicleModel?.rotate(new Vector3(0, 0, 1), model.rotation.z);

				if (vehicleModel) {
					const scale = model.scale;
					vehicleModel.scaling = new Vector3(scale, scale, scale);
				}

				glbModel.meshes[1].receiveShadows = true;
				base.shadows.getShadowMap()?.renderList?.push(glbModel.meshes[1]);

				vehicleMaterial = getMaterial(glbModel, /MAT_VEH_GravCycle/);
				if (vehicleMaterial instanceof PBRMaterial) {
					vehicleMaterial.albedoColor = new Color3(1, 1, 1);
					vehicleMaterial.alphaMode = 7;
					vehicleMaterial.maxSimultaneousLights = 16;
					vehicleMaterial.needDepthPrePass = true;
					vehicleMaterial.transparencyMode = 3;
					vehicleMaterial.iridescence.isEnabled = true;
					vehicleMaterial.iridescence.intensity = 0.25;
					vehicleMaterial.iridescence.indexOfRefraction = 1.6;
				}

				const glowTexture = new Texture('/T_PFX_Clouds_Difference01.png', base.scene);
				glowMaterial.diffuseTexture = glowTexture;
				glowMaterial.diffuseTexture.hasAlpha = true;
				glowMaterial.diffuseColor = new Color3(0, 0, 1);
				glowMaterial.alphaMode = 6; // One one.
				glowMaterial.useAlphaFromDiffuseTexture = true;
				glowMaterial.emissiveTexture = glowTexture;
				glowMaterial.emissiveColor = new Color3(0.1, 0.2, 1.0);
				glowMaterial.emissiveTexture.level = 0.08;
				glowMaterial.specularPower = 128;
				glowMaterial.disableLighting = true;
				const clone = glbModel.meshes[1].clone('vehicleGlow', null);
				if (clone) {
					clone.material = glowMaterial;
				}
			});
	}

	// Add mist to the scene.
	createMist(base.scene);

	// Slowly rotate the model clockwise.
	base.scene.beforeRender = () => {
		vehicleModel?.rotate(new Vector3(0, 1, 0), 0.0005, Space.WORLD);

		// Animate the glow texture.
		if (glbModel) {
			// @ts-ignore
			glowMaterial.diffuseTexture.vOffset -= 0.0002;
			// @ts-ignore
			glowMaterial.diffuseTexture.vOffset -= 0.0002;
		}
	};

	// Run the render loop.
	base.start();

	return () => base.stop();
}
