/**
 * Projector rules slides content («Королевская битва» regulations).
 * Pure data — rendering lives in QuizDisplay. Bullet strings support
 * **bold** markers, rendered by the slide component.
 */

export interface QuizRuleScore {
    tone: "good" | "mid" | "bad";
    label: string;
    points: string;
}

export interface QuizOrderRules {
    tone: "red" | "yellow" | "green";
    name: string;
    points: string;
    mistakes: string;
    note: string;
}

export interface QuizRoundRules {
    round: number;
    name: string;
    format: string;
    params: string[];
    bullets: string[];
    /** Main scoring feature, shown in the "overview" columns. */
    highlight: string;
    /** Round 2 grading scale badges. */
    scores?: QuizRuleScore[];
    /** Round 3 important note above the order cards. */
    note?: string;
    /** Round 3 order (Орден) cards. */
    orders?: QuizOrderRules[];
}

export const QUIZ_ROUND_RULES: QuizRoundRules[] = [
    {
        round: 1,
        name: "Стрелы логики",
        format: "Командный «Брейн-ринг»",
        params: ["До 10 вопросов", "60 секунд", "+2 балла за ответ"],
        bullets: [
            "После старта таймера команда, готовая ответить, нажимает сигнальную кнопку.",
            "Верный ответ приносит **2 балла** — вопрос закрывается.",
            "При ошибке команда теряет право ответа на этот вопрос, а ход переходит следующей команде (максимум **3 попытки** на вопрос).",
        ],
        highlight: "+2 балла за верный ответ",
    },
    {
        round: 2,
        name: "Вектор удара",
        format: "«Что? Где? Когда?» — ответы на бланках",
        params: ["7 вопросов", "60 секунд", "До 4 баллов за ответ"],
        bullets: [
            "Все команды играют одновременно и записывают ответы в бумажные бланки.",
            "По сигналу таймера на экран выводится правильный ответ, герольды проверяют бланки.",
        ],
        highlight: "+4 / +2 / 0 — по оценке герольда",
        scores: [
            { tone: "good", label: "Полностью верный", points: "4 балла" },
            { tone: "mid", label: "Частично верный", points: "2 балла" },
            { tone: "bad", label: "Неверный / нет ответа", points: "0 баллов" },
        ],
    },
    {
        round: 3,
        name: "Формула клинка",
        format: "Конкурс капитанов («Умники и умницы»)",
        params: ["5 вопросов", "120 секунд (2 мин)", "Только капитаны"],
        bullets: [],
        highlight: "5 / 3 / 1 балла — по выбранному Ордену",
        note: "Перед 1-м вопросом капитан выбирает Орден — изменить выбор по ходу тура нельзя. Все баллы, набранные до выбывания, сохраняются в общем зачёте.",
        orders: [
            {
                tone: "red",
                name: "Красный орден",
                points: "5 баллов за верный ответ",
                mistakes: "0 прав на ошибку",
                note: "выбывание после 1-й ошибки",
            },
            {
                tone: "yellow",
                name: "Жёлтый орден",
                points: "3 балла за верный ответ",
                mistakes: "1 право на ошибку",
                note: "выбывание после 2-й ошибки",
            },
            {
                tone: "green",
                name: "Зелёный орден",
                points: "1 балл за верный ответ",
                mistakes: "2 права на ошибку",
                note: "выбывание после 3-й ошибки",
            },
        ],
    },
];

export const QUIZ_TIEBREAK_RULES = {
    title: "Тай-брейк",
    subtitle: "Борьба за призовые места (1 – 4-е место)",
    params: ["60 секунд", "Брейн-ринг"],
    bullets: [
        "Кнопки активны только у команд с равным счётом за призовое место.",
        "Очки тай-брейка определяют порядок только внутри этой группы.",
    ],
};

export const QUIZ_RULES_FOOTER =
    "Баллы за все три тура суммируются. Побеждает команда с наибольшей суммой баллов.";
