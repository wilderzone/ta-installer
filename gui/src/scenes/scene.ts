import type { AbstractMesh, ISceneLoaderAsyncResult, TransformNode } from '@babylonjs/core';
import { Color3, PBRMaterial, Space, StandardMaterial, Texture, Vector3 } from '@babylonjs/core';
import { watch } from 'vue';
import { store } from '../store';
import { AsyncDebounce } from '../utilities';
import { BaseScene } from './base';
import { getMaterials, getMeshes, getNode } from './gltf';
import { items, Model } from './item';
import { createMist } from './mist';

const meshFilter = /(?:SKL_VEH_GravCycle|SKL_VEH_Shrike|SKL_VEH_Beowulf_primitive0|SKL_VEH_Beowulf_primitive1)/;
const materialFilter = /(?:MAT_VEH_GravCycle|MAT_Shrike|MAT_Beowulf|MIC_Chaingun_1P)/;

export function homeScene(canvas: HTMLCanvasElement): () => void {
	const base = new BaseScene(canvas);
	base.camera.setTarget(new Vector3(0, -0.05, 0));
	base.camera.setPosition(new Vector3(0, 0.35, -1.6));
	base.lights[0].intensity = 0.5;

	// Create a GLB model.
	let glbModel: ISceneLoaderAsyncResult | undefined;
	let vehicleModel: TransformNode | undefined;
	const glowMeshes = new Array<AbstractMesh>();
	const glowMaterial = new StandardMaterial('glow', base.scene);
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

	async function loadModel(model: Model): Promise<void> {
		if (glbModel) {
			base.unloadModel(glbModel);
		}
		for (const mesh of glowMeshes) {
			mesh.dispose(true, false);
		}
		glowMeshes.length = 0;

		const result = await base.loadModel(model.directory.href, model.file);
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

		const shadowRenderList = base.shadows.getShadowMap()?.renderList;
		for (const mesh of getMeshes(glbModel, meshFilter)) {
			mesh.receiveShadows = true;
			mesh.doNotSyncBoundingInfo = true;
			mesh.alwaysSelectAsActiveMesh = true;
			shadowRenderList?.push(mesh);
		}

		const materials = getMaterials(glbModel, materialFilter);
		for (const material of materials) {
			if (material instanceof PBRMaterial) {
				material.albedoColor = new Color3(1, 1, 1);
				material.invertNormalMapY = true;
				material.maxSimultaneousLights = 16;
				material.clearCoat.isEnabled = true;
				material.clearCoat.intensity = 0.3;
				material.clearCoat.roughness = 0.1;
				material.iridescence.isEnabled = true;
				material.iridescence.intensity = 0.25;
				material.iridescence.indexOfRefraction = 1.6;
			}
		}

		for (const mesh of getMeshes(glbModel, meshFilter)) {
			const clone = mesh.clone(`${mesh.name}-glow`, null, true);
			if (clone) {
				clone.material = glowMaterial;
				glowMeshes.push(clone);
			}
		}
	}

	// Reload the model whenever it changes.
	const load = new AsyncDebounce(loadModel);
	watch(
		() => store.model,
		value => {
			const model = items.get(value)?.model;
			if (!model) return;
			load.call(model);
		},
		{ immediate: true }
	);

	// Add mist to the scene.
	createMist(base.scene);

	// Slowly rotate the model clockwise.
	base.scene.beforeRender = () => {
		const deltaTime = base.scene.deltaTime || 0;
		const rotation = 0.00005 * deltaTime;
		vehicleModel?.rotate(new Vector3(0, 1, 0), rotation, Space.WORLD);

		// Animate the glow texture.
		if (glbModel) {
			const distance = 0.00003 * deltaTime;
			// @ts-ignore
			glowMaterial.diffuseTexture.vOffset -= distance;
			// @ts-ignore
			glowMaterial.diffuseTexture.vOffset -= distance;
		}
	};

	// Run the render loop.
	base.start();

	return () => base.stop();
}
