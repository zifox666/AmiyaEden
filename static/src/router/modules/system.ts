import { AppRouteRecord } from '@/types/router'

export const systemRoutes: AppRouteRecord = {
  path: '/system',
  name: 'System',
  component: '/index/index',
  meta: {
    title: 'menus.system.title',
    icon: 'ri:user-3-line',
    roles: ['super_admin', 'admin']
  },
  children: [
    {
      path: 'user',
      name: 'User',
      component: '/system/user',
      meta: {
        title: 'menus.system.user',
        keepAlive: true,
        roles: ['super_admin', 'admin'],
        authList: [
          { title: '删除用户', authMark: 'delete_user' },
          { title: '分配角色', authMark: 'assign_role' }
        ]
      }
    },
    {
      path: 'esi-refresh',
      name: 'ESIRefresh',
      component: '/system/esi-refresh',
      meta: {
        title: 'menus.system.esiRefresh',
        keepAlive: true,
        roles: ['super_admin', 'admin'],
        authList: [{ title: '执行任务', authMark: 'execute_task' }]
      }
    },
    {
      path: 'wallet',
      name: 'SystemWallet',
      component: '/system/wallet',
      meta: {
        title: 'menus.system.wallet',
        keepAlive: true,
        roles: ['super_admin', 'admin'],
        authList: [
          { title: '调整余额', authMark: 'adjust_balance' },
          { title: '查看日志', authMark: 'view_log' }
        ]
      }
    },
    {
      path: 'pap-exchange',
      name: 'PAPExchange',
      component: '/system/pap-exchange',
      meta: {
        title: 'menus.system.papExchange',
        keepAlive: true,
        roles: ['super_admin', 'admin'],
        authList: [{ title: '编辑兑换率', authMark: 'edit_exchange_rate' }]
      }
    },
    {
      path: 'newbro-settings',
      name: 'NewbroSettings',
      component: '/system/newbro-settings',
      meta: {
        title: 'menus.system.newbroSettings',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'pap',
      name: 'AlliancePAP',
      component: '/system/pap',
      meta: {
        title: 'menus.system.alliancePap',
        keepAlive: true,
        roles: ['super_admin', 'admin'],
        authList: [{ title: '手动拉取', authMark: 'manual_fetch' }]
      }
    },
    {
      path: 'npc-kills',
      name: 'CorpNpcKillReport',
      component: '/system/npc-kills',
      meta: {
        title: 'menus.system.npcKills',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'auto-role',
      name: 'AutoRole',
      component: '/system/auto-role',
      meta: {
        title: 'menus.system.autoRole',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'user-center',
      name: 'UserCenter',
      component: '/system/user-center',
      meta: {
        title: 'menus.system.userCenter',
        isHide: true,
        keepAlive: true,
        isHideTab: true
      }
    },
    {
      path: 'webhook',
      name: 'WebhookSettings',
      component: '/system/webhook',
      meta: {
        title: 'menus.system.webhook',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'basic-config',
      name: 'BasicConfig',
      component: '/system/basic-config',
      meta: {
        title: 'menus.system.basicConfig',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
  ]
}
