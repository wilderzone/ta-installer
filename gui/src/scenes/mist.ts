import { Color4, GPUParticleSystem, MeshBuilder, ParticleSystem, Texture, Vector3 } from '@babylonjs/core';
import type { Scene } from '@babylonjs/core';

export function createMist(scene: Scene) {
	// Leave the mist out entirely for reduced motion.
	if (matchMedia('(prefers-reduced-motion: reduce)').matches) return;

	// Setup environment.
	const fountain = MeshBuilder.CreateBox('mistFountain', { size: 0.01 }, scene);
	fountain.isVisible = false;

	// Create a particle system.
	let particleSystem: ParticleSystem | GPUParticleSystem | undefined = undefined;

	// const fogTexture = new Texture('https://raw.githubusercontent.com/aWeirdo/Babylon.js/master/smoke_15.png', scene);
	const fogTexture = new Texture('/ThumbnailTexture.png', scene);

	function createNewSystem(): void {
		if (particleSystem) {
			particleSystem.dispose();
		}

		const emitBox = {
			from: new Vector3(-30, -15, -10),
			to: new Vector3(30, 10, 40)
		};

		if (GPUParticleSystem.IsSupported) {
			particleSystem = new GPUParticleSystem('particles', { capacity: 200 }, scene);
			particleSystem.activeParticleCount = 100;
			particleSystem.manualEmitCount = particleSystem.activeParticleCount;
			particleSystem.minEmitBox = emitBox.from;
			particleSystem.maxEmitBox = emitBox.to;
		} else {
			particleSystem = new ParticleSystem('particles', 2500 , scene);
			particleSystem.manualEmitCount = particleSystem.getCapacity();
			particleSystem.minEmitBox = emitBox.from;
			particleSystem.maxEmitBox = emitBox.to;
		}

		particleSystem.particleTexture = fogTexture.clone();
		particleSystem.emitter = fountain;

		particleSystem.color1 = new Color4(0.8, 0.8, 0.8, 0.03);
		particleSystem.color2 = new Color4(.95, .95, .95, 0.08);
		particleSystem.colorDead = new Color4(0.9, 0.9, 0.9, 0.03);
		particleSystem.minSize = 4.5;
		particleSystem.maxSize = 8.0;
		particleSystem.minLifeTime = Number.MAX_SAFE_INTEGER;
		particleSystem.emitRate = 50000;
		particleSystem.blendMode = ParticleSystem.BLENDMODE_STANDARD;
		particleSystem.gravity = new Vector3(0, 0, 0);
		particleSystem.direction1 = new Vector3(0, 0, 0);
		particleSystem.direction2 = new Vector3(0, 0, 0);
		particleSystem.minAngularSpeed = -1.5;
		particleSystem.maxAngularSpeed = 1.5;
		particleSystem.minEmitPower = .5;
		particleSystem.maxEmitPower = 1;
		particleSystem.updateSpeed = 0.005;
	
		particleSystem.start();
	}

	createNewSystem();
}
