import { AppRouteRecord } from '@/types/router'

export const voiceRoutes: AppRouteRecord = {
  path: '/voice',
  name: 'VoiceCenter',
  component: '/index/index',
  meta: {
    title: 'menus.voice.title',
    icon: 'ri:mic-line'
  },
  children: [
    {
      path: 'mumble',
      name: 'MumbleCenter',
      component: '/voice/mumble',
      meta: {
        title: 'menus.voice.mumble',
        keepAlive: true
      }
    }
  ]
}
