import { ArcRotateCamera, Color3, Color4, CubeTexture, DefaultRenderingPipeline, DirectionalLight, Engine, GlowLayer, HemisphericLight, ImageProcessingConfiguration, ImportMeshAsync, Light, Scene, ShadowGenerator, Vector3 } from '@babylonjs/core';
import type { AbstractEngine, ISceneLoaderAsyncResult, Observer } from '@babylonjs/core';

export class BaseScene {
	public camera: ArcRotateCamera;
	public canvas: HTMLCanvasElement;
	public engine: Engine;
	public glow: GlowLayer;
	public lights = new Array<Light>();
	public pipeline: DefaultRenderingPipeline;
	public resizeObserver?: Observer<AbstractEngine>;
	public scene: Scene;
	public shadows: ShadowGenerator;

	constructor(canvas: HTMLCanvasElement) {
		// Create the engine.
		this.canvas = canvas;
		this.engine = new Engine(canvas);
		this.engine.maxFPS = 60;

		// Create a new scene.
		this.scene = new Scene(this.engine);
		this.scene.clearColor = new Color4(0, 0, 0, 0);
		this.scene.skipPointerMovePicking = true;

		// Camera.
		this.camera = new ArcRotateCamera('Camera', 0, 0, 10, new Vector3(0, 0, 0), this.scene);
		this.camera.setPosition(new Vector3(0, -1.5, -11));
		this.camera.minZ = 0.01;

		// Create the rendering pipeline.
		// At least one other effect needs to be enabled for anti-aliasing to actually get applied.
		this.pipeline = new DefaultRenderingPipeline('default', false, this.scene, [this.camera]);
		this.pipeline.imageProcessing.exposure = 1.2;
		this.pipeline.imageProcessing.toneMappingEnabled = true;
		this.pipeline.imageProcessing.toneMappingType = ImageProcessingConfiguration.TONEMAPPING_ACES;
		this.pipeline.samples = 4;
		this.pipeline.bloomEnabled = true;
		this.pipeline.bloomScale = 0.5;
		this.pipeline.bloomWeight = 0.4;

		// Fog.
		this.scene.fogMode = Scene.FOGMODE_LINEAR;
		this.scene.fogColor = new Color3(0, 0, 0);
		this.scene.fogDensity = 0.05;
		this.scene.fogStart = 20.0;
		this.scene.fogEnd = 45.0;

		// Lights.
		const light = new HemisphericLight('light', new Vector3(0, 1, 0), this.scene);
		light.intensity = 0.25;
		const light2 = new DirectionalLight('directionalLight', new Vector3(1, -3, -0.5), this.scene);
		light2.position = new Vector3(-0.6, 2, 0);
		light2.intensity = 12;
		light2.autoUpdateExtends = false;
		light2.orthoLeft = -1;
		light2.orthoRight = 1;
		light2.orthoTop = 1;
		light2.orthoBottom = -1;
		light2.shadowMinZ = 1;
		light2.shadowMaxZ = 4;
		this.lights.push(light);
		this.lights.push(light2);

		// Enable shadows.
		this.shadows = new ShadowGenerator(2048, light2);
		this.shadows.bias = 0.0005;
		this.shadows.darkness = 0.3;
		this.shadows.filter = ShadowGenerator.FILTER_PCF;
		this.shadows.normalBias = 0.002;

		// Glow.
		this.glow = new GlowLayer('glow', this.scene);
		this.glow.intensity = 1.5;
		this.glow.blurKernelSize = 32;

		this.scene.environmentTexture = CubeTexture.CreateFromPrefilteredData('/Studio_Softbox_2Umbrellas_cube_specular.env', this.scene);
		this.scene.environmentIntensity = 0.3;
	}

	/** Handle the canvas being resized. */
	private onResize(): void {
		const width = this.engine.getRenderWidth();
		const height = this.engine.getRenderHeight();
		const ratio = width / height;
		if (ratio >= 1.0) {
			this.camera.fovMode = ArcRotateCamera.FOVMODE_VERTICAL_FIXED;
		} else {
			this.camera.fovMode = ArcRotateCamera.FOVMODE_HORIZONTAL_FIXED;
		}
	}

	/** Load a GLB/GLTF model into the scene. */
	public loadModel(path: string, file: string): Promise<ISceneLoaderAsyncResult> {
		return ImportMeshAsync(file, this.scene, { rootUrl: path });
	}

	/** Remove a GLB/GLTF model from the scene. */
	public unloadModel(model: ISceneLoaderAsyncResult): void {
		for (const item of model.animationGroups) item.dispose();
		for (const item of model.geometries) item.dispose();
		for (const item of model.lights) item.dispose();
		for (const item of model.meshes) item.dispose();
		for (const item of model.particleSystems) item.dispose();
		for (const item of model.skeletons) item.dispose();
		for (const item of model.spriteManagers) item.dispose();
		for (const item of model.transformNodes) item.dispose();
	}

	/** Render the scene. */
	public render(): void {
		this.scene.render();
	}

	/** Start the render loop. */
	public start(): void {
		this.engine.runRenderLoop(() => this.scene.render());
		this.resizeObserver?.remove();
		this.resizeObserver = this.engine.onResizeObservable.add(this.onResize.bind(this));
	}

	/** Stop the render loop. */
	public stop(): void {
		this.engine.stopRenderLoop();
		this.resizeObserver?.remove();
	}
}
