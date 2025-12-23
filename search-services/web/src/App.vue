<script setup lang="ts">
import { RouterView } from 'vue-router'
import Footer from '@/components/FooterComponent.vue'
import { NConfigProvider, NMessageProvider, type GlobalThemeOverrides } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#6e7b91',
    primaryColorHover: '#5a6476',
    primaryColorPressed: '#5a6476',
    primaryColorSuppl: '#6e7b91',
  },
  Input: {
    border: '1.5px solid var(--surface-border)',
  },
  Tabs: {
    tabFontWeightActive: 600,
  },
  Pagination: {
    // itemColor: 'var(--color-primary)', // Цвет неактивных кнопок
    // itemBorder: '1px solid #d9d9d9', // Обводка
    // itemBorderHover: '1px solid #1890ff', // Обводка при наведении
    // itemBorderActive: '2px solid #1890ff', // Обводка активной страницы
    // itemBorderRadius: '6px', // Радиус скругления
    // itemSizeMedium: '36px', // Размер кнопки (medium)
    // itemTextColor: '#ffffff', // Цвет текста
    // itemTextColorHover: '#fff', // Цвет текста при наведении
    // itemTextColorPressed: '#096dd9', // Цвет текста при нажатии
    // itemTextColorActive: '#ffffff', // Цвет текста активной страницы
    // itemColorActive: '#1890ff', // Фон активной страницы (primary)
    // Для других состояний:
    // itemColorActiveHover: '#5a6476',
    // itemColorActivePressed: '#096dd9',
    // Цвет кнопок "вперед/назад" в простом (simple) режиме
    buttonColor: 'var(--color-primary)',
    buttonColorHover: '#5a6476',
    buttonColorPressed: '#5a6476',
    buttonIconColor: '#fff',
    buttonIconColorHover: '#fff',
    buttonIconColorPressed: '#fff',
  },
}
</script>

<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-message-provider :duration="4000">
      <div id="app">
        <main class="page-content">
          <RouterView />
        </main>
        <Footer />
      </div>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
html,
body,
#app {
  height: 100vh;
  margin: 0;
}

#app {
  display: flex;
  flex-direction: column;
}

/* главный контент между header и footer */
.page-content {
  flex: 1; /* растягивает всё пространство между header и footer */
  display: flex;
  flex-direction: column;
}
</style>
