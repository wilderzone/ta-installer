<script setup lang="ts">
</script>

<template>
	<div class="logo">
		<img src="../assets/tribes_ascend.png" alt="Tribes: Ascend" draggable="false">
		<img src="../assets/tribes_ascend.png" alt="Tribes: Ascend" draggable="false">
	</div>
</template>

<style scoped>
.logo {
	display: grid;
	place-items: center;
	height: 100%;
	min-height: 0;

	& img {
		position: absolute;
		height: 100%;
		min-height: 0;
		margin: auto;
	}

	& img:nth-of-type(1) {
		filter: blur(50px);
	}
}
</style>

<style>
:root.linux .logo img:nth-of-type(1) {
	display: none;
}
</style>
