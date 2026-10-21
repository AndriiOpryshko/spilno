// Ukrainian UI copy. The product is Ukrainian-first; add a sibling locale file with the
// same shape when another language is needed.

const plural = new Intl.PluralRules('uk-UA');
const pick = (n: number, forms: { one: string; few: string; many: string }) => {
  const rule = plural.select(n);
  return rule === 'one' ? forms.one : rule === 'few' ? forms.few : forms.many;
};

export const uk = {
  app: { name: 'Спільно', osbb: 'ОСББ «Січових Стрільців, 12»', osbbShort: 'ОСББ 12', search: 'Пошук і команди' },
  nav: { cases: 'Справи', protocols: 'Протоколи', meetings: 'Збори', documents: 'Документи', soon: 'Незабаром' },
  caseStatus: { draft: 'Чернетка', review: 'На перевірці', ready: 'Готово до друку', printed: 'Надруковано' },
  documentKind: { offer: 'Пропозиція', invoice: 'Рахунок', contract: 'Договір', act: 'Акт' },
  origin: { system: 'Система', ai: 'ШІ', human: 'Людина', editing: 'Редагується' },
  cases: {
    eyebrow: 'Реєстр справ ОСББ',
    title: 'Справи',
    summary: (total: number, attention: number) =>
      `${total} ${pick(total, { one: 'справа', few: 'справи', many: 'справ' })} · ${attention} ${pick(attention, { one: 'потребує', few: 'потребують', many: 'потребують' })} вашої уваги`,
    search: 'Пошук за назвою',
    all: 'Усі',
    columns: { case: 'Справа', documents: 'Документи', amount: 'Сума', status: 'Статус' },
    open: 'Відкрити',
    created: (date: string) => `Створено ${date}`,
    amountNotFound: 'Суму не знайдено',
    loadError: 'Не вдалося завантажити справи',
    loadErrorHint: 'Перевірте з’єднання та спробуйте ще раз. Фільтри збережено.',
    retry: 'Спробувати ще раз',
    empty: 'Немає справ за цим фільтром',
    emptyHint: 'Очистьте пошук або оберіть інший статус.',
    clear: 'Скинути фільтри',
  },
  protocol: {
    eyebrow: 'Протокол засідання правління',
    title: (number: string, caseTitle: string) => `№ ${number} · ${caseTitle}`,
    breadcrumb: (number: string) => `Протокол № ${number}`,
    intro: (n: number) =>
      n === 0
        ? 'Усі поля перевірено — протокол можна друкувати.'
        : `Перевірте ${n} ${pick(n, { one: 'поле', few: 'поля', many: 'полів' })} нижче — решту заповнено з документів справи.`,
    saved: 'Збережено',
    savedJustNow: 'Збережено щойно',
    readOnly: 'лише читання',
    manual: 'Заповнено вручну',
    notFoundInDocs: 'Не знайдено в документах справи',
    placeholder: 'Впишіть значення',
    fix: 'Підставити',
    save: 'Зберегти',
    saving: 'Збереження…',
    cancel: 'Скасувати',
    keySave: 'зберегти',
    keyCancel: 'скасувати',
    readiness: 'Готовність до друку',
    print: 'Надрукувати протокол',
    allVerified: 'Усі поля перевірено',
    needsMeeting: (label: string) => `${label}: потрібне рішення зборів`,
    needsReview: (label: string) => `${label}: уточніть значення`,
    missing: (label: string) => `${label}: не вказано`,
    props: { status: 'Статус', date: 'Дата', board: 'Правління', documents: 'Документи' },
    unavailable: 'Чернетки протоколу ще немає',
    unavailableHint: (caseId: string) => `Документи справи ${caseId} ще обробляються. Чернетка з’явиться тут автоматично.`,
    backToCases: 'До справ',
  },
};

export const formatDate = (iso: string) => new Date(iso).toLocaleDateString('uk-UA');

const uahParts = new Intl.NumberFormat('uk-UA', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

export const formatUah = (amount: number) => `${uahParts.format(amount)} грн`;

// "184 500" and ",00" separately, so the integer part can be set large.
export const splitUah = (amount: number) => {
  const [int = '', frac = '00'] = uahParts.format(amount).split(',');
  return { int, frac: `,${frac} грн` };
};
