import { DefaultRenderingPipeline, Engine, Scene } from '@babylonjs/core';
import type { AbstractEngine, Observer } from '@babylonjs/core';

export class Optimizer {
	private engine: Engine;
	private scene: Scene;
	private pipeline: DefaultRenderingPipeline;
	private targetFrameRate: number;
	private targetFrameTime: number;
	private history = new Array<number>();
	private buffer = new Array<number>();
	private bufferFull = false;
	private currentLevel = 0;
	private enabled = false;
	private optimizationWindow = 2000;
	private observable?: Observer<AbstractEngine>;

	constructor(engine: Engine, scene: Scene, pipeline: DefaultRenderingPipeline, targetFrameRate = 60) {
		this.engine = engine;
		this.scene = scene;
		this.pipeline = pipeline;
		this.targetFrameRate = targetFrameRate;
		this.targetFrameTime = Math.floor(1000 / this.targetFrameRate);
	}

	private reset(): void {
		this.history = [performance.now()];
		this.buffer = [];
		this.bufferFull = false;
	}

	/** Remove old entries from `buffer` & `history`. */
	private trim(): void {
		let i = this.buffer.length - 1;
		let total = 0;
		while (total < this.optimizationWindow && i > 0) {
			total += this.buffer[i];
			i--;
		}
		if (i > 0) {
			this.buffer = this.buffer.slice(i);
			this.history = this.history.slice(i - 1);
			this.bufferFull = true;
		} else {
			this.bufferFull = false;
		}
	}

	/** Record a frame-time measurement based on the provided timestamp. */
	private recordMeasurement(timestamp: number): void {
		this.history.push(timestamp);
		if (this.history.length >= 2) {
			const frameTime = timestamp - this.history.at(-2)!;
			this.buffer.push(frameTime);
		}
		this.trim();
	}

	private optimize(): void {
		const total = this.buffer.reduce((sum, value) => sum + value, 0);
		const averageFrameTime = total / this.buffer.length;
		if (this.bufferFull && averageFrameTime > this.targetFrameTime) {
			this.currentLevel++;
			this.setFeatures(this.currentLevel);
			this.reset();
			if (this.currentLevel >= 3) {
				this.enabled = false;
				this.observable?.remove();
			}
		}
	}

	private setFeatures(level: number): void {
		if (level === 0) {
			const glow = this.scene.getGlowLayerByName('glow');
			if (glow) {
				glow.isEnabled = true;
			}
			this.pipeline.bloomEnabled = true;
			this.scene.shadowsEnabled = true;
			this.engine.setHardwareScalingLevel(1);
		}

		if (level === 1) {
			const glow = this.scene.getGlowLayerByName('glow');
			if (glow) {
				glow.isEnabled = false;
			}
			this.pipeline.bloomEnabled = false;
			this.scene.shadowsEnabled = true;
			this.engine.setHardwareScalingLevel(1);
		}

		if (level === 2) {
			const glow = this.scene.getGlowLayerByName('glow');
			if (glow) {
				glow.isEnabled = false;
			}
			this.pipeline.bloomEnabled = false;
			this.scene.shadowsEnabled = false;
			this.engine.setHardwareScalingLevel(1);
		}

		if (level === 3) {
			const glow = this.scene.getGlowLayerByName('glow');
			if (glow) {
				glow.isEnabled = false;
			}
			this.pipeline.bloomEnabled = false;
			this.scene.shadowsEnabled = false;
			this.engine.setHardwareScalingLevel(2);
		}
	}

	public start(): void {
		this.enabled = true;
		this.reset();
		this.observable?.remove();
		this.observable = this.engine.onEndFrameObservable.add(() => {
			if (this.enabled) {
				this.recordMeasurement(performance.now());
				this.optimize();
			}
		});
	}
}
