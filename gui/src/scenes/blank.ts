import { BaseScene } from './base';
import { createMist } from './mist';

export function blankScene(canvas?: HTMLCanvasElement): () => void {
	const base = new BaseScene(canvas);

	// Add mist to the scene.
	createMist(base.scene);

	// Run the render loop.
	base.start();

	return () => base.stop();
}
