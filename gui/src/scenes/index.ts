import "@babylonjs/loaders/glTF";
import './debug.css';
import { blankScene } from './blank';
import { homeScene } from './scene';

export type Scene = (canvas: HTMLCanvasElement) => StopScene;
export type StopScene = (() => void) | undefined;

export const scenes = {
	blank: <Scene> blankScene,
	home: <Scene> homeScene,
} as const;
