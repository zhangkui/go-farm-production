# UI 美化方案：登录页 + 主布局菜单

## 问题诊断（含一个真实 bug）
1. **菜单图标不显示（bug）**：`AppLayout.vue` 里用 `<DataLine/>`、`<OfficeBuilding/>` 等组件形式，但从未导入、也未全局注册 → 图标全部解析失败，菜单光秃秃只剩文字。`LoginView` 的 `:prefix-icon` 是导入后传入的所以正常。
2. 登录页是单卡片，视觉层次单薄，无品牌感。
3. 侧边栏 17 个菜单项平铺，过长且无分组。
4. 未统一品牌主色（Element Plus 默认蓝色），与绿色品牌不搭。

## 改动清单
1. **`main.ts`**：全局注册 `@element-plus/icons-vue` 所有图标；引入全局主题样式。
2. **新增 `src/styles/theme.css`**：覆盖 Element Plus CSS 变量（`--el-color-primary` = 绿 `#2e7d32` 及其色阶），统一圆角/阴影/滚动条/背景。
3. **`LoginView.vue` 重设计**：左右分屏——左侧品牌面板（绿色渐变 + SVG logo + 系统名 + 标语 + 装饰），右侧表单卡片；窄屏自动堆叠；逻辑不变。
4. **`RegisterView.vue`**：同步登录页风格。
5. **`AppLayout.vue` 重设计**：
   - 顶部 logo（SVG mark + 文字，收起时只显示 mark）
   - 菜单按业务域分组成 `el-sub-menu`：仪表盘置顶，其余分 资源管理 / 生产管理 / 投入品 / 库存与成本 / 系统
   - 顶栏：折叠按钮 + 当前页面包屑 + 用户下拉（首字母头像 + 姓名 + 角色）
   - 激活/hover/收起态样式打磨，修复图标渲染
   - 侧边栏改为更柔和的深绿渐变 + 分组标题

## 不改动
- 业务逻辑、API、路由表结构、权限指令、各业务列表页（本次聚焦登录与主框架）

## 验证
- `npm run type-check` + `npm run build` 通过
- 浏览器查看登录页与登录后菜单视觉
