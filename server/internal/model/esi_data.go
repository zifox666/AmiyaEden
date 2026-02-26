package model

// ─────────────────────────────────────────────
//  ESI 数据模型别名 — 实际定义在 internal/model/esi/
// ─────────────────────────────────────────────

import esimodel "amiya-eden/internal/model/esi"

type EveCharacterAsset = esimodel.EveCharacterAsset

type EveCharacterNotification = esimodel.EveCharacterNotification

type EveCharacterTitle = esimodel.EveCharacterTitle

type EveCharacterClone = esimodel.EveCharacterClone

type EveKillmailList = esimodel.EveKillmailList
type EveKillmailItem = esimodel.EveKillmailItem
type EveCharacterKillmail = esimodel.EveCharacterKillmail

type EveCharacterContract = esimodel.EveCharacterContract

type EVECharacterWallet = esimodel.EVECharacterWallet
type EVECharacterWalletJournal = esimodel.EVECharacterWalletJournal
type EVECharacterWalletTransaction = esimodel.EVECharacterWalletTransaction

type EveCharacterSkill = esimodel.EveCharacterSkill
type EveCharacterSkills = esimodel.EveCharacterSkills
type EveCharacterSkillQueue = esimodel.EveCharacterSkillQueue
