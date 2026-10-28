import { formatUah } from '@/i18n/uk';

import type { CaseDocument, CaseSummary, ProtocolDraft, ProtocolSlot } from './types';

const roofDocs = [
  { id: 'd-1', kind: 'offer' as const, title: 'Комерційна пропозиція ТОВ «Дах-Сервіс»' },
  { id: 'd-2', kind: 'contract' as const, title: 'Договір № 14/10' },
];

export const cases: CaseSummary[] = [
  { id: 'c-101', title: 'Гідроізоляція покрівлі, під’їзд 2', createdAt: '2026-10-02', status: 'review', amountUah: 184500, documents: roofDocs },
  { id: 'c-102', title: 'Щорічний огляд ліфта', createdAt: '2026-10-09', status: 'draft', amountUah: null, documents: [{ id: 'd-3', kind: 'invoice', title: 'Рахунок № 2291' }] },
  {
    id: 'c-103',
    title: 'Заміна освітлення в підвалі',
    createdAt: '2026-09-18',
    status: 'ready',
    amountUah: 23760,
    documents: [
      { id: 'd-4', kind: 'invoice', title: 'Рахунок № 118' },
      { id: 'd-5', kind: 'contract', title: 'Договір № 9/09' },
    ],
  },
  {
    id: 'c-104',
    title: 'Ремонт дитячого майданчика',
    createdAt: '2026-08-27',
    status: 'printed',
    amountUah: 61200,
    documents: [
      { id: 'd-6', kind: 'contract', title: 'Договір № 3/08' },
      { id: 'd-7', kind: 'act', title: 'Акт виконаних робіт' },
    ],
  },
  {
    id: 'c-105',
    title: 'Прибирання прибудинкової території',
    createdAt: '2026-07-14',
    status: 'printed',
    amountUah: 38400,
    documents: [{ id: 'd-8', kind: 'contract', title: 'Договір № 5/07' }],
  },
  {
    id: 'c-106',
    title: 'Встановлення шлагбаума у дворі',
    createdAt: '2026-08-05',
    status: 'draft',
    amountUah: 96000,
    documents: [{ id: 'd-9', kind: 'offer', title: 'Комерційна пропозиція ТОВ «Брама-Плюс»' }],
  },
];

const board = ['О. Коваленко (голова)', 'І. Мельник', 'Т. Шевчук', 'Н. Бондар'];

// Every slot of a frozen protocol is confirmed: no AI-proposed value is left unconfirmed (AC-06).
function confirmedSlots(o: { agenda: string; amount: number; inWords: string; contractor: string; funding: string; decision: string; contract: CaseDocument }): ProtocolSlot[] {
  const ref = (clause: string, fragment: string) => ({ documentId: o.contract.id, documentTitle: `${o.contract.title}, ${clause}`, fragment });
  return [
    { id: 's-agenda', label: 'Питання порядку денного', origin: 'ai', value: o.agenda, status: 'verified', editable: true, sourceRef: ref('п. 1.1', o.agenda) },
    {
      id: 's-amount',
      label: 'Сума договору',
      origin: 'ai',
      value: formatUah(o.amount),
      status: 'verified',
      editable: false,
      sourceRef: ref('п. 3.1', 'Загальна вартість робіт'),
      amount: { uah: o.amount, inWords: o.inWords },
    },
    { id: 's-contractor', label: 'Підрядник', origin: 'human', value: o.contractor, status: 'verified', editable: true, sourceRef: ref('преамбула', o.contractor) },
    { id: 's-funding', label: 'Джерело коштів', origin: 'human', value: o.funding, status: 'verified', editable: true, sourceRef: null },
    { id: 's-decision', label: 'Рішення', origin: 'human', value: o.decision, status: 'verified', editable: true, sourceRef: null },
  ];
}

export const protocols: Record<string, ProtocolDraft> = {
  'c-101': {
    caseId: 'c-101',
    state: 'draft',
    caseTitle: 'Гідроізоляція покрівлі, під’їзд 2',
    number: '17/2026',
    date: '2026-10-21',
    board,
    documents: roofDocs,
    slots: [
      {
        id: 's-agenda',
        label: 'Питання порядку денного',
        origin: 'ai',
        value: 'Затвердження договору на гідроізоляцію покрівлі над під’їздом № 2',
        status: 'verified',
        editable: true,
        sourceRef: { documentId: 'd-2', documentTitle: 'Договір № 14/10, п. 1.1', fragment: 'Предмет договору: гідроізоляція покрівлі над під’їздом № 2' },
      },
      {
        id: 's-amount',
        label: 'Сума договору',
        origin: 'ai',
        value: '184 500,00 грн',
        status: 'verified',
        editable: false,
        sourceRef: { documentId: 'd-2', documentTitle: 'Договір № 14/10, п. 3.1', fragment: 'Загальна вартість: 184 500,00 грн з ПДВ' },
        amount: { uah: 184500, inWords: 'сто вісімдесят чотири тисячі п’ятсот гривень 00 копійок' },
      },
      {
        id: 's-contractor',
        label: 'Підрядник',
        origin: 'ai',
        value: 'Дах-Сервіс',
        status: 'needs_review',
        editable: true,
        sourceRef: { documentId: 'd-1', documentTitle: 'Комерційна пропозиція ТОВ «Дах-Сервіс»', fragment: 'Товариство з обмеженою відповідальністю «Дах-Сервіс»' },
        expected: { value: 'ТОВ «Дах-Сервіс»', reason: 'У договорі повна назва — ТОВ «Дах-Сервіс». Протокол має з нею збігатися.' },
      },
      {
        id: 's-funding',
        label: 'Джерело коштів',
        origin: 'ai',
        value: '',
        status: 'not_found',
        editable: true,
        sourceRef: null,
        examples: ['Фонд ремонту', 'Внески співвласників', 'Залишок 2025 року'],
      },
      {
        id: 's-decision',
        label: 'Рішення',
        origin: 'human',
        value: 'Затвердити договір і доручити голові правління його підписати.',
        status: 'verified',
        editable: true,
        sourceRef: null,
      },
    ],
    warnings: [
      {
        id: 'w-1',
        tier: 3,
        title: 'Рішення за загальними зборами',
        message: 'Ліміт правління — 150 000 грн (Статут, п. 9.4). Винесіть питання на збори або зменште суму.',
        slotId: 's-amount',
        ruleRef: { title: 'Статут, п. 9.4', fragment: 'Правління ухвалює рішення щодо договорів до 150 000 грн' },
      },
    ],
  },
  'c-103': {
    caseId: 'c-103',
    state: 'ready_for_vote',
    caseTitle: 'Заміна освітлення в підвалі',
    number: '16/2026',
    date: '2026-10-14',
    board,
    documents: cases[2]!.documents,
    slots: confirmedSlots({
      agenda: 'Заміна світильників у підвалі на світлодіодні',
      amount: 23760,
      inWords: 'двадцять три тисячі сімсот шістдесят гривень 00 копійок',
      contractor: 'ФОП Савчук О. П.',
      funding: 'Фонд ремонту',
      decision: 'Замовити заміну світильників у ФОП Савчук О. П. за рахунком № 118.',
      contract: cases[2]!.documents[1]!,
    }),
    warnings: [],
    poll: {
      round: 2,
      voted: 3,
      boardSize: board.length,
      items: [
        { id: 'p-1', title: 'Заміна освітлення в підвалі', for: 3, against: 0, abstain: 0 },
        { id: 'p-2', title: 'Джерело коштів: фонд ремонту', for: 2, against: 1, abstain: 0 },
      ],
    },
  },
  'c-104': {
    caseId: 'c-104',
    state: 'ready_for_signing',
    caseTitle: 'Ремонт дитячого майданчика',
    number: '15/2026',
    date: '2026-09-30',
    board,
    documents: cases[3]!.documents,
    slots: confirmedSlots({
      agenda: 'Ремонт дитячого майданчика у дворі будинку',
      amount: 61200,
      inWords: 'шістдесят одна тисяча двісті гривень 00 копійок',
      contractor: 'ТОВ «Ігросвіт»',
      funding: 'Внески співвласників',
      decision: 'Затвердити договір № 3/08 і прийняти роботи за актом.',
      contract: cases[3]!.documents[0]!,
    }),
    warnings: [],
  },
  'c-105': {
    caseId: 'c-105',
    state: 'signed',
    caseTitle: 'Прибирання прибудинкової території',
    number: '12/2026',
    date: '2026-07-28',
    board,
    documents: cases[4]!.documents,
    slots: confirmedSlots({
      agenda: 'Договір на прибирання прибудинкової території на 2026 рік',
      amount: 38400,
      inWords: 'тридцять вісім тисяч чотириста гривень 00 копійок',
      contractor: 'ФОП Гнатюк Л. М.',
      funding: 'Внески співвласників',
      decision: 'Укласти договір № 5/07 на рік.',
      contract: cases[4]!.documents[0]!,
    }),
    warnings: [],
  },
  'c-106': {
    caseId: 'c-106',
    state: 'withdrawn',
    caseTitle: 'Встановлення шлагбаума у дворі',
    number: '13/2026',
    date: '2026-08-19',
    board,
    documents: cases[5]!.documents,
    slots: confirmedSlots({
      agenda: 'Встановлення шлагбаума на в’їзді у двір',
      amount: 96000,
      inWords: 'дев’яносто шість тисяч гривень 00 копійок',
      contractor: 'ТОВ «Брама-Плюс»',
      funding: 'Фонд ремонту',
      decision: 'Замовити встановлення шлагбаума у ТОВ «Брама-Плюс».',
      contract: cases[5]!.documents[0]!,
    }),
    warnings: [],
    poll: {
      round: 1,
      voted: 4,
      boardSize: board.length,
      items: [{ id: 'p-1', title: 'Встановлення шлагбаума у дворі', for: 1, against: 3, abstain: 0 }],
    },
    withdrawal: { reason: 'Пропозиція не набрала підтримки на попередньому голосуванні.', date: '2026-08-26' },
  },
};
