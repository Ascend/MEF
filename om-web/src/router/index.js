/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { createRouter, createWebHistory } from 'vue-router'
import networkImg from '../assets/img/manager/network.svg'
import timeImg from '../assets/img/manager/time.svg';
import diskImg from '../assets/img/manager/disk.svg';
import alarmImg from '../assets/img/maintenance/alarm.svg';
import updateImg from '../assets/img/maintenance/update.svg';
import reloadImg from '../assets/img/maintenance/reload.svg';
import journalImg from '../assets/img/maintenance/journal.svg';
import registrationImg from '../assets/img/maintenance/registration.svg';
import informationImg from '../assets/img/maintenance/information.svg';
import safetyImg from '../assets/img/setting/safety.svg';
import resetImg from '../assets/img/setting/reset.svg';
import restoreImg from '../assets/img/setting/restore.svg';
import extendModuleImg from '../assets/img/manager/train.svg';
import { isA500, isA200 } from '@/utils/commonMethods';
import { queryAllModules } from '@/api/extendModule';
import { generateA200Routes } from '@/utils/commonMethods';

const addA500DynamicRoute = () => {
  const parentRoute = router.options.routes.find(el => el.name === 'setting');
  if (parentRoute.children[parentRoute.children.length - 1].path !== 'restore') {
    let resetItem = {
      name: 'reset.menuName',
      path: 'reset',
      img: resetImg,
      component: () => import('@/views/setting/reset/Index.vue'),
    }
    let restoreItem = {
      name: 'restore.menuName',
      path: 'restore',
      img: restoreImg,
      component: () => import('@/views/setting/restore/Index.vue'),
    }
    router.addRoute('setting', resetItem)
    router.addRoute('setting', restoreItem)
    parentRoute.children.push(resetItem);
    parentRoute.children.push(restoreItem);
  }
}

const addA200DynamicRoute = async () => {
  // 如果能从 sessionStorage 获取 hasExtendModule，并且值为 true，说明已经添加过动态路由，无须再请求接口
  if (JSON.parse(sessionStorage.getItem('hasExtendModule'))) {
    return;
  }

  try {
    let { data } = await queryAllModules(false, true);
    let hasExtendModule = data && data?.Members && data?.Members.length !== 0;
    sessionStorage.setItem('hasExtendModule', hasExtendModule);

    // 没有扩展模组列表则不添加动态路由
    if (!hasExtendModule) {
      return
    }

    addExtendModuleRoute();
  } catch (e) {
    // 防止在A500上请求扩展模组的接口失败导致页面不可访问
    return
  }
}

const addExtendModuleRoute = () => {
  const parentRoute = router.options.routes.find(el => el.name === 'manager');
  let extendedItem = {
    name: 'extendModule.menuName',
    path: 'extendModule',
    img: extendModuleImg,
    component: () => import('@/views/manager/extendModule/Index.vue'),
  }
  if (parentRoute.children[parentRoute.children.length - 1].path !== 'extendModule') {
    router.addRoute('manager', extendedItem);
    parentRoute.children.push(extendedItem);
  }
}

const handleBeforeEnter = (to, from, next, route) => {
  const firstChild = to.matched[0].children[0];
  if (firstChild && to.fullPath === `/${route}`) {
    // 如果存在第一个子路由，则重定向到该路由
    next({ path: `/${route}/${firstChild.path}` });
  } else {
    // 否则不进行重定向
    next();
  }
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
    },
    {
      path: '/changePassword',
      name: 'changePassword',
      component: () => import('@/views/ChangePassword.vue'),
    },
    {
      name: 'manager',
      path: '/manager',
      component: () => import('@/views/manager/Index.vue'),
      beforeEnter: (to, from, next) => {
        handleBeforeEnter(to, from, next, 'manager');
      },
      children: [
        {
          name: 'network.menuName',
          path: 'network',
          img: networkImg,
          component: () => import('@/views/manager/network/Index.vue'),
        },
        {
          name: 'time.menuName',
          path: 'time',
          img: timeImg,
          component: () => import('@/views/manager/time/Index.vue'),
        },
        {
          name: 'registration.menuName',
          path: 'registration',
          img: registrationImg,
          component: () => import('@/views/manager/registration/Index.vue'),
        },
        {
          name: 'disk.menuName',
          path: 'disk',
          img: diskImg,
          component: () => import('@/views/manager/disk/Index.vue'),
        },
        {
          name: 'alarm.menuName',
          path: 'alarm',
          img: alarmImg,
          component: () => import('@/views/manager/alarm/Index.vue'),
        },
        {
          name: 'journal.menuName',
          path: 'journal',
          img: journalImg,
          component: () => import('@/views/manager/journal/Index.vue'),
        },

        {
          name: 'update.menuName',
          path: 'update',
          img: updateImg,
          component: () => import('@/views/manager/update/Index.vue'),
        },
        {
          name: 'reload.menuName',
          path: 'reload',
          img: reloadImg,
          component: () => import('@/views/manager/reload/Index.vue'),
        },
        {
          name: 'information.menuName',
          path: 'information',
          img: informationImg,
          component: () => import('@/views/manager/information/Index.vue'),
        },
      ],
    },
    {
      name: 'setting',
      path: '/setting',
      component: () => import('@/views/setting/Index.vue'),
      beforeEnter: (to, from, next) => {
        handleBeforeEnter(to, from, next, 'setting');
      },
      children: [
        {
          name: 'safety.menuName',
          path: 'safety',
          img: safetyImg,
          component: () => import('@/views/setting/safety/Index.vue'),
        },
      ],
    },
  ],
})

// 防止页面刷新后，router.beforeEach 不会执行，导致动态路由丢失
if (isA500()) {
  addA500DynamicRoute();
}

// 防止页面刷新后动态路由丢失，在设备为 A200 且 sessionStorage 的 hasExtendModule 值为 true 时添加扩展模组路由
if (isA200() && JSON.parse(sessionStorage.getItem('hasExtendModule'))) {
  addExtendModuleRoute();
}

await generateA200Routes(router);
export default router;

router.beforeEach(async (to, from, next) => {
  if (to.name !== 'login' && !sessionStorage.getItem('token')) {
    next({
      name: 'login',
    })
  } else {
    // 用户登录后才能动态添加路由
    if (isA500()) {
      addA500DynamicRoute();
    }
    if (isA200() && to.path.indexOf('manager') !== -1) {
      await addA200DynamicRoute()
    }
    next()
  }
})
