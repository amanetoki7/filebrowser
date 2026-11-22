<template>
  <img :src="musicIcon" class="music-file-icon" alt="" aria-hidden="true" draggable="false" />
</template>
<script setup>
import musicIcon from '@/assets/music-file.svg';
</script>
<style>
.image-preview.music-preview { background: transparent !important; border: none !important; box-shadow: none !important; }
.image-preview > .music-file-icon { width: 100%; height: 100%; object-fit: contain; display: block; }
</style>
