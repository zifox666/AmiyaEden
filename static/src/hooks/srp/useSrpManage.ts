import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTable } from '@/hooks/core/useTable'
import { useNameResolver } from '@/hooks'
import { useEnterSearch } from '@/hooks/core/useEnterSearch'
import { useUserStore } from '@/store/modules/user'
import { fetchFleetList } from '@/api/fleet'
import { fetchApplicationList } from '@/api/srp'
import { formatIskSmart, formatTime } from '@utils/common'
import { ElTag, ElTooltip, ElLink } from 'element-plus'
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import ArtCopyButton from '@/components/core/forms/art-copy-button/index.vue'

type SrpApp = Api.Srp.Application
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

export function useSrpManage(callbacks: {
  openReviewDialog: (row: SrpApp, action: 'approve' | 'reject') => void
  handlePayoutAction: (row: SrpApp) => void
  openKmPreview: (row: SrpApp) => void
}) {
  const { t } = useI18n()
  const { getName, resolve: resolveNames } = useNameResolver()
  const { createEnterSearchHandler } = useEnterSearch()
  const userStore = useUserStore()

  const canPayout = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((r) => ['super_admin', 'admin', 'srp'].includes(r))
  })

  // ─── Fleets ───
  const fleets = ref<Api.Fleet.FleetItem[]>([])
  const fleetMap = computed(() => new Map(fleets.value.map((f) => [f.id, f])))
  const loadFleets = async () => {
    try {
      const res = await fetchFleetList({ size: 200 } as any)
      fleets.value = res?.list ?? []
    } catch {
      fleets.value = []
    }
  }

  // ─── Filters & Tabs ───
  const activeTab = ref('pending')
  const payoutMode = ref<Api.Srp.PayoutMode>('fuxi_coin')
  const filter = reactive({ review_status: '', fleet_id: '', keyword: '' })

  // ─── Formatters ───
  const reviewStatusType = (s: string): TagType =>
    (({ pending: 'info', approved: 'success', rejected: 'danger' }) as Record<string, TagType>)[
      s
    ] ?? 'info'

  const reviewStatusLabel = (s: string) =>
    ({
      submitted: t('srp.status.submitted'),
      approved: t('srp.status.approved'),
      rejected: t('srp.status.rejected')
    })[s as 'submitted' | 'approved' | 'rejected'] ?? s

  const payoutStatusType = (s: string): TagType => (s === 'paid' ? 'success' : 'warning')

  const formatFleetLabel = (f: Api.Fleet.FleetItem) =>
    `${f.fc_character_name}: ${f.title} (${f.pap_count}PAP) @ ${formatTime(f.start_at)} ~ ${formatTime(f.end_at)}`

  // ─── Table ───
  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    getData,
    searchParams
  } = useTable({
    core: {
      apiFn: fetchApplicationList,
      apiParams: { current: 1, size: 200, tab: 'pending' },
      columnsFactory: () => [
        { type: 'index', width: 40, label: '#' },
        {
          prop: 'review_status',
          label: t('srp.manage.columns.review'),
          width: 80,
          formatter: (row: SrpApp) => {
            const tag = h(ElTag, { type: reviewStatusType(row.review_status), size: 'small' }, () =>
              reviewStatusLabel(row.review_status)
            )
            if (row.review_note) {
              return h(ElTooltip, { content: row.review_note, placement: 'top' }, () => tag)
            }
            return tag
          }
        },
        {
          prop: 'payout_status',
          label: t('srp.manage.columns.payout'),
          width: 80,
          formatter: (row: SrpApp) =>
            h(ElTag, { type: payoutStatusType(row.payout_status), size: 'small' }, () =>
              row.payout_status === 'paid' ? t('srp.status.paid') : t('srp.status.notpaid')
            )
        },
        {
          prop: 'last_actor_nickname',
          label: t('srp.manage.columns.lastActor'),
          width: 130,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h(
              'span',
              { class: row.last_actor_nickname ? '' : 'text-gray-400' },
              row.last_actor_nickname || '-'
            )
        },
        {
          prop: 'nickname',
          label: t('srp.manage.columns.nickname'),
          width: 120,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', { class: row.nickname ? '' : 'text-gray-400' }, row.nickname || '-')
        },
        {
          prop: 'character_name',
          label: t('srp.manage.columns.character'),
          width: 140,
          formatter: (row: SrpApp) =>
            h('div', { class: 'flex items-center gap-1 min-w-0' }, [
              h('span', { class: 'truncate' }, row.character_name || '-'),
              h(ArtCopyButton, { text: row.character_name })
            ])
        },
        {
          prop: 'ship_type_id',
          label: t('srp.manage.columns.ship'),
          width: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', {}, getName(row.ship_type_id, `TypeID: ${row.ship_type_id}`, 'type'))
        },
        {
          prop: 'recommended_amount',
          label: t('srp.manage.columns.recommendedAmount'),
          width: 90,
          formatter: (row: SrpApp) => h('span', {}, formatIskSmart(row.recommended_amount))
        },
        {
          prop: 'final_amount',
          label: t('srp.manage.columns.finalAmount'),
          width: 90,
          formatter: (row: SrpApp) =>
            h('span', { class: 'font-semibold text-blue-600' }, formatIskSmart(row.final_amount))
        },
        {
          prop: 'fleet_title',
          label: t('srp.manage.columns.fleet'),
          width: 150,
          formatter: (row: SrpApp) => {
            if (!row.fleet_id) return h('span', { class: 'text-gray-400' }, '-')
            const fleet = fleetMap.value.get(row.fleet_id)
            const tooltipContent = fleet
              ? formatFleetLabel(fleet)
              : row.fleet_fc_name
                ? `${row.fleet_fc_name}: ${row.fleet_title || row.fleet_id}`
                : row.fleet_title || row.fleet_id
            return h(ElTooltip, { content: tooltipContent, placement: 'top' }, () =>
              h('span', { class: 'cursor-default' }, row.fleet_title || row.fleet_id || '')
            )
          }
        },
        {
          prop: 'solar_system_id',
          label: t('srp.manage.columns.system'),
          width: 128,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', {}, getName(row.solar_system_id, String(row.solar_system_id), 'solar_system'))
        },
        {
          prop: 'killmail_id',
          label: t('srp.manage.columns.killId'),
          width: 96,
          formatter: (row: SrpApp) =>
            h(
              ElLink,
              {
                href: `https://zkillboard.com/kill/${row.killmail_id}/`,
                target: '_blank',
                type: 'primary'
              },
              () => String(row.killmail_id)
            )
        },
        {
          prop: 'killmail_time',
          label: t('srp.manage.columns.kmTime'),
          width: 160,
          formatter: (row: SrpApp) => h('span', {}, formatTime(row.killmail_time))
        },
        {
          prop: 'corporation_id',
          label: t('srp.manage.columns.corporation'),
          width: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h(
              'span',
              {},
              getName(
                row.corporation_id,
                row.corporation_id ? `ID: ${row.corporation_id}` : '-',
                'esi'
              )
            )
        },
        {
          prop: 'alliance_id',
          label: t('srp.manage.columns.alliance'),
          width: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h(
              'span',
              {},
              getName(row.alliance_id, row.alliance_id ? `ID: ${row.alliance_id}` : '-', 'esi')
            )
        },
        {
          prop: 'note',
          label: t('srp.manage.columns.note'),
          minWidth: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', { class: row.note ? '' : 'text-gray-400' }, row.note || '-')
        },
        {
          prop: 'review_note',
          label: t('srp.manage.columns.reviewNote'),
          minWidth: 170,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', { class: row.review_note ? '' : 'text-gray-400' }, row.review_note || '-')
        },
        {
          prop: 'actions',
          label: t('srp.manage.columns.action'),
          width: 220,
          fixed: 'right',
          formatter: (row: SrpApp) => {
            const btns: ReturnType<typeof h>[] = [
              h(ArtButtonTable, { type: 'view', onClick: () => callbacks.openKmPreview(row) })
            ]
            if (row.review_status === 'submitted') {
              btns.push(
                h(ArtButtonTable, {
                  label: t('srp.manage.approveBtn'),
                  elType: 'success',
                  onClick: () => callbacks.openReviewDialog(row, 'approve')
                }),
                h(ArtButtonTable, {
                  label: t('srp.manage.rejectBtn'),
                  elType: 'danger',
                  onClick: () => callbacks.openReviewDialog(row, 'reject')
                })
              )
            } else if (row.review_status === 'approved' && row.payout_status === 'notpaid') {
              if (canPayout.value) {
                btns.push(
                  h(ArtButtonTable, {
                    label: t('srp.manage.payoutBtn'),
                    elType: 'primary',
                    onClick: () => callbacks.handlePayoutAction(row)
                  })
                )
              }
              btns.push(
                h(ArtButtonTable, {
                  label: t('srp.manage.editBtn'),
                  elType: 'warning',
                  onClick: () => callbacks.openReviewDialog(row, 'approve')
                }),
                h(ArtButtonTable, {
                  label: t('srp.manage.reRejectBtn'),
                  elType: 'danger',
                  onClick: () => callbacks.openReviewDialog(row, 'reject')
                })
              )
            } else if (row.review_status === 'rejected') {
              btns.push(
                h(ArtButtonTable, {
                  label: t('srp.manage.reApproveBtn'),
                  elType: 'success',
                  onClick: () => callbacks.openReviewDialog(row, 'approve')
                })
              )
            }
            return h('div', { class: 'flex items-center gap-1' }, btns)
          }
        }
      ]
    }
  })

  // ─── Name Resolution ───
  watch(data, async (list) => {
    if (list.length) await resolveManageNames(list)
  })

  const resolveManageNames = async (list: SrpApp[]) => {
    const typeIds = new Set<number>()
    const solarIds = new Set<number>()
    const esiIds = new Set<number>()
    for (const app of list) {
      if (app.ship_type_id) typeIds.add(app.ship_type_id)
      if (app.solar_system_id) solarIds.add(app.solar_system_id)
      if (app.corporation_id) esiIds.add(app.corporation_id)
      if (app.alliance_id) esiIds.add(app.alliance_id)
    }
    await resolveNames({
      ids: {
        ...(typeIds.size ? { type: [...typeIds] } : {}),
        ...(solarIds.size ? { solar_system: [...solarIds] } : {})
      },
      esi: esiIds.size ? [...esiIds] : undefined
    })
  }

  // ─── Search & Filter ───
  const handleSearch = () => {
    Object.assign(searchParams, {
      current: 1,
      tab: activeTab.value,
      review_status: filter.review_status || undefined,
      fleet_id: filter.fleet_id || undefined,
      keyword: filter.keyword.trim() || undefined
    })
    getData()
  }

  const handleKeywordSearchKeyup = createEnterSearchHandler(handleSearch)

  const resetFilter = () => {
    filter.review_status = ''
    filter.fleet_id = ''
    filter.keyword = ''
    Object.assign(searchParams, {
      current: 1,
      tab: activeTab.value,
      review_status: undefined,
      fleet_id: undefined,
      keyword: undefined
    })
    getData()
  }

  const handleTabChange = () => {
    filter.review_status = ''
    filter.fleet_id = ''
    filter.keyword = ''
    Object.assign(searchParams, {
      current: 1,
      tab: activeTab.value,
      review_status: undefined,
      fleet_id: undefined,
      keyword: undefined
    })
    getData()
  }

  // ─── Export ───
  const manageExportHeaders = {
    character_name: '人物',
    ship_name: '舰船',
    solar_system: '星系',
    killmail_id: 'KillID',
    killmail_time: 'KM时间',
    corporation: '军团',
    alliance: '联盟',
    fleet_title: '关联舰队',
    fleet_fc_name: 'FC',
    note: '备注',
    recommended_amount: '推荐金额',
    final_amount: '最终金额',
    review_status: '审批状态',
    review_note: '审批备注',
    payout_status: '发放状态',
    last_actor_nickname: '最后处理人'
  }

  const exportManageData = computed(() =>
    data.value.map((app) => ({
      character_name: app.character_name,
      ship_name: getName(app.ship_type_id, `TypeID: ${app.ship_type_id}`, 'type'),
      solar_system: getName(app.solar_system_id, String(app.solar_system_id), 'solar_system'),
      killmail_id: app.killmail_id,
      killmail_time: formatTime(app.killmail_time),
      corporation: getName(
        app.corporation_id,
        app.corporation_id ? `ID: ${app.corporation_id}` : '-',
        'esi'
      ),
      alliance: getName(app.alliance_id, app.alliance_id ? `ID: ${app.alliance_id}` : '-', 'esi'),
      fleet_title: app.fleet_title || '-',
      fleet_fc_name: app.fleet_fc_name || '-',
      note: app.note || '-',
      recommended_amount: app.recommended_amount,
      final_amount: app.final_amount,
      review_status: reviewStatusLabel(app.review_status),
      review_note: app.review_note || '-',
      payout_status: app.payout_status === 'paid' ? t('srp.status.paid') : t('srp.status.notpaid'),
      last_actor_nickname: app.last_actor_nickname || '-'
    }))
  )

  // ─── KM Preview ───
  const kmPreviewVisible = ref(false)
  const previewKillmailId = ref(0)
  const openKmPreview = (row: SrpApp) => {
    previewKillmailId.value = row.killmail_id
    kmPreviewVisible.value = true
  }

  onMounted(() => {
    loadFleets()
  })

  return {
    // permissions
    canPayout,
    // fleets
    fleets,
    fleetMap,
    formatFleetLabel,
    // filters
    activeTab,
    payoutMode,
    filter,
    handleSearch,
    handleKeywordSearchKeyup,
    resetFilter,
    handleTabChange,
    // table
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    // formatters
    formatISK: formatIskSmart,
    // export
    manageExportHeaders,
    exportManageData,
    // km preview
    kmPreviewVisible,
    previewKillmailId,
    openKmPreview
  }
}
