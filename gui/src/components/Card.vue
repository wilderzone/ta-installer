<script setup lang="ts">
defineProps<{
	name: string;
}>();
</script>

<template>
	<section class="card" :name>
		<slot></slot>
	</section>
</template>

<style scoped>
.card {
	width: 100%;
	height: 100%;
	padding: var(--gap);
	background: var(--color-card);
	border-radius: var(--border-radius);
	box-shadow: 0 5px 12px -4px #0008;
	backdrop-filter: blur(8px) brightness(0.3);
}
</style>
