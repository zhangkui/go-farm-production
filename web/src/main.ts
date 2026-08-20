import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import App from './App.vue'
import { router } from './router'
import { permissionDirective } from './directives/permission'
import './styles/theme.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

// Globally register every Element Plus icon so templates can use
// <IconName /> without per-file imports (used heavily in the layout menu).
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component as any)
}

// Register the v-permission directive for declarative RBAC in templates.
app.directive('permission', permissionDirective)
app.mount('#app')
