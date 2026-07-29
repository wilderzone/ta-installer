<script setup lang="ts">
import Inputs from './Inputs.vue';
import StepNavigation from './StepNavigation.vue';
import Steps from './Steps.vue';
</script>

<template>
	<form class="configurator" action="" @submit.prevent>
		<Steps v-grid:steps />
		<Inputs v-grid:inputs />
		<StepNavigation v-grid:navigation />
	</form>
</template>

<style scoped>
.configurator {
	display: grid;
	grid-template:
		'steps     ' auto
		'inputs    ' 1fr
		'navigation' auto
		/ 1fr;
	gap: var(--gap);
	height: 100%;
}
</style>
