<template>
  <router-view />
</template>

<script setup lang="ts">
// Root component. Layout is handled per-route via AppLayout for authenticated
// pages and bare views for login/register.
</script>

<style>
html,
body,
#app {
  height: 100%;
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}
</style>
