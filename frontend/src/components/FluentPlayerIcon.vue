<template><span class="fluent-player-icon" aria-hidden="true" v-html="markup"></span></template>
<script setup>
import { computed } from 'vue';
const props = defineProps({ name: { type: String, required: true } });
const icons = import.meta.glob('../assets/fluent-player/*.svg', { query: '?raw', import: 'default', eager: true });
const markup = computed(() => String(icons[`../assets/fluent-player/${props.name}.svg`] || '').replaceAll('#212121', 'currentColor'));
</script>
<style>
.fluent-player-icon { display: inline-flex; width: 24px; height: 24px; flex: 0 0 24px; color: inherit; }
.fluent-player-icon svg { display: block; width: 100%; height: 100%; fill: currentColor; }
</style>
