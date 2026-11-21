<script setup lang="ts">
</script>

<template>
	<h1 class="logo">
		<img src="../assets/tribes_ascend.png" alt="" draggable="false">
		<img src="../assets/tribes_ascend.png" alt="Tribes: Ascend Installer" draggable="false">
	</h1>
</template>

<style scoped>
.logo {
	display: grid;
	place-items: center;
	height: 100%;
	min-height: 0;
	margin: 0;

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
