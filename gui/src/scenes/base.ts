import { ArcRotateCamera, Color3, Color4, DefaultRenderingPipeline, DirectionalLight, Engine, GlowLayer, HemisphericLight, ImportMeshAsync, Light, Matrix, Scene, ShadowGenerator, Vector2, Vector3 } from '@babylonjs/core';
import type { AbstractEngine, EngineView, ISceneLoaderAsyncResult, Observer } from '@babylonjs/core';
import { Optimizer } from './optimizer';

/**
 * Whilst one engine can render multiple scenes at the same time,\
 * each scene size/resolution needs to be handled by a different engine\
 * so that the size of each engine's rendering canvas remains consistent.
 * 
 * Because of this, {@link GlobalEngine} has multiple layers - each to contain a different general scene size.
 * 
 * See {@link https://doc.babylonjs.com/features/featuresDeepDive/scene/multiCanvas#performance | Using Multiple Canvases: Performance} for more information.
 */
type SceneLayer = 'background' | 'foreground';

interface View {
	canvas: HTMLCanvasElement;
	layer: SceneLayer;
	scene: BaseScene;
	view: EngineView;
}

class GlobalEngine {
	private static readonly canvas = document.createElement('canvas');
	public static readonly engine: Record<SceneLayer, Engine> = {
		background: new Engine(GlobalEngine.canvas),
		foreground: new Engine(GlobalEngine.canvas)
	};
	private static views = new Array<View>();

	/**
	 * Add a scene to the global rendering engine.
	 * 
	 * @param scene The scene to add.
	 * @param canvas The associated canvas.
	 * @param layer _(Optional, default: `'background'`)_ The layer on which to render this scene. This determines which engine will be used.
	 */
	public static addScene(scene: BaseScene, canvas: HTMLCanvasElement, layer: SceneLayer = 'background'): void {
		if (GlobalEngine.views.find(view => view.scene === scene)) return;
		const engine = GlobalEngine.engine[layer];
		const view: View = {
			canvas,
			layer,
			scene,
			view: engine.registerView(canvas, scene.camera, true)
		};
		GlobalEngine.views.push(view);
	}

	/**
	 * Remove a scene from the global rendering engine.
	 * 
	 * @param scene The scene to remove.
	 */
	public static removeScene(scene: BaseScene): void {
		const index = GlobalEngine.views.findIndex(view => view.scene === scene);
		const view = GlobalEngine.views[index];
		GlobalEngine.engine[view.layer].unRegisterView(view.canvas);
		if (index >= 0) GlobalEngine.views.splice(index, 1);
	}

	static {
		// Start the render loops.
		GlobalEngine.engine['background'].runRenderLoop(() => {
			for (const { layer, scene } of GlobalEngine.views) {
				if (layer === 'background') scene.render();
			}
		});
		GlobalEngine.engine['foreground'].runRenderLoop(() => {
			for (const { layer, scene } of GlobalEngine.views) {
				if (layer === 'foreground') scene.render();
			}
		});
	}
}

export class BaseScene {
	public camera: ArcRotateCamera;
	public canvas: HTMLCanvasElement | null;
	public engine: Engine;
	public glow: GlowLayer;
	public layer: SceneLayer;
	public lights = new Array<Light>();
	public optimizer: Optimizer;
	public pipeline: DefaultRenderingPipeline;
	public resizeObserver?: Observer<AbstractEngine>;
	public scene: Scene;
	public shadows: ShadowGenerator;

	constructor(canvas?: HTMLCanvasElement, layer: SceneLayer = 'background') {
		this.engine = GlobalEngine.engine[layer];

		// Get the canvas element.
		this.canvas = canvas || document.querySelector('canvas');

		// Create a new scene.
		this.layer = layer;
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
		const light2 = new DirectionalLight('directionalLight', new Vector3(0, -10, -1), this.scene);
		light2.position = new Vector3(0, 2, 0);
		light2.intensity = 15;
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
		this.shadows = new ShadowGenerator(1024, light2);
		this.shadows.darkness = 0.00;
		this.shadows.filter = ShadowGenerator.FILTER_POISSONSAMPLING;

		// Glow.
		this.glow = new GlowLayer('glow', this.scene);
		this.glow.intensity = 2.0;
		this.glow.blurKernelSize = 64;

		// Optimizer.
		this.optimizer = new Optimizer(this.engine, this.scene, this.pipeline);
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

	/** Convert screen coordinates to world coordinates. */
	public screenToWorld(vector: Vector2, depth = 0.99935): Vector3 {
		const viewMatrix = this.scene.getViewMatrix();
		const projectionMatrix = this.scene.getProjectionMatrix();
		if (!viewMatrix || !projectionMatrix) {
			return new Vector3();
		}
		return Vector3.Unproject(
			new Vector3(vector.x, vector.y, depth),
			this.engine.getRenderWidth(),
			this.engine.getRenderHeight(),
			Matrix.Identity(),
			viewMatrix,
			projectionMatrix
		);
	}

	/** Convert world coordinates to screen coordinates. */
	public worldToScreen(vector: Vector3): Vector2 {
		const world = Vector3.Project(
			vector,
			Matrix.Identity(),
			this.scene.getTransformMatrix(),
			this.camera.viewport.toGlobal(
				this.engine.getRenderWidth(),
				this.engine.getRenderHeight()
			)
		);
		return new Vector2(world.x, world.y);
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
		if (!this.canvas) return;
		GlobalEngine.addScene(this, this.canvas, this.layer);
		this.resizeObserver?.remove();
		this.resizeObserver = this.engine.onResizeObservable.add(this.onResize.bind(this));
		this.optimizer.start();
	}

	/** Stop the render loop. */
	public stop(): void {
		GlobalEngine.removeScene(this);
		this.resizeObserver?.remove();
	}
}
