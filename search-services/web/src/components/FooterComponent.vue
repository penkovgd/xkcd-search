<script setup lang="ts">
const year = new Date().getFullYear()
</script>

<template>
  <footer class="site-footer">
    <div class="footer-content">
      <span>{{ year }} All comics sourced from </span>
      <a href="https://xkcd.com" target="_blank" rel="noopener noreferrer">xkcd.com</a>
      <span class="separator">|</span>
      <a href="https://github.com/penkovgd" target="_blank" rel="noopener noreferrer"> GitHub </a>
    </div>
  </footer>
</template>

<style scoped>
.site-footer {
  text-align: center;
  padding: 12px 0;
  margin-top: 40px;
  border-top: 1px solid var(--divider-color);
  font-size: 14px;
  color: white;
}
.footer-content {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.separator {
  margin: 0 6px;
}
a {
  color: inherit;
  text-decoration: underline;
}
</style>
