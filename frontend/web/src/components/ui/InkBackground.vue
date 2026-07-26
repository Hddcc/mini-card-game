<script setup lang="ts">
/** 固定背景层：径向金/玉光晕 + 墨色渐变 + 5% 噪点（DESIGN.md: Elevation Base Layer） */
</script>

<template>
  <div class="pointer-events-none fixed inset-0 z-0" aria-hidden="true">
    <div
      class="absolute inset-0 opacity-70 mix-blend-screen bg-[radial-gradient(circle_at_30%_16%,rgba(255,225,109,.24),transparent_24%),radial-gradient(circle_at_76%_60%,rgba(111,242,174,.14),transparent_24%),linear-gradient(135deg,#2d2d2d_0%,#131313_62%,#0e0e0e_100%)]"
    ></div>
    <div class="absolute inset-0 bg-gradient-to-t from-background via-transparent to-background/50"></div>
    <div class="ink-overlay absolute inset-0"></div>
  </div>
</template>
