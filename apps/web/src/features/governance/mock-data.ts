import type { CaseSummary, ProtocolDraft } from './types';

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
];

export const protocols: Record<string, ProtocolDraft> = {
  'c-101': {
    caseId: 'c-101',
    caseTitle: 'Гідроізоляція покрівлі, під’їзд 2',
    number: '17/2026',
    date: '2026-10-21',
    board: ['О. Коваленко (голова)', 'І. Мельник', 'Т. Шевчук'],
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
};
