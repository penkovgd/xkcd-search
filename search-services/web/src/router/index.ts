import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/HomePage.vue'
import Login from '@/views/LoginPage.vue'
import SearchResults from '@/views/SearchResultsPage.vue'
import AdminPanelPage from '@/views/AdminPanelPage.vue'
import CategoryComicsPage from '@/views/CategoryComicsPage.vue'
import { isTokenValid } from '@/utils/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home,
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    },
    {
      path: '/admin',
      name: 'admin',
      component: AdminPanelPage,
      meta: { requiresAuth: true },
    },
    {
      path: '/search',
      name: 'search',
      component: SearchResults,
    },
    {
      path: '/category/:category?',
      name: 'category',
      component: CategoryComicsPage,
      props: true,
    },
  ],
   scrollBehavior() {
    return { top: 0 }
  }
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth) {
    const isAuthenticated = isTokenValid()

    if (!isAuthenticated) {
      return next({ name: 'login', query: { redirect: to.fullPath } })
    }
  }
  next()
})

export default router
