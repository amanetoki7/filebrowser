<template>
  <img :src="folderIcon" class="windows-folder-icon" alt="" aria-hidden="true" draggable="false" />
</template>
<script setup>
import folderIcon from '@/assets/windows11-folder.svg';
</script>
<style>
.windows-folder-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  object-fit: contain;
  vertical-align: middle;
  flex-shrink: 0;
}
.image-preview > .windows-folder-icon { width: 100%; height: 100%; object-fit: contain; }
.action > .windows-folder-icon { font-size: 1.5em; }
</style>
