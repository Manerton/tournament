import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
    Badge,
    Button,
    Form,
    Modal,
    Nav,
    ProgressBar,
    Spinner,
    Table,
    Toast,
    ToastContainer,
} from "react-bootstrap";
import { useNavigate, useParams } from "react-router-dom";
import { useQuizState } from "../helpers/useQuizState";
import {
    axiosQuizAdjustScore,
    axiosQuizCloseQuestion,
    axiosQuizDeleteButton,
    axiosQuizEndTiebreak,
    axiosQuizGetQuestions,
    axiosQuizGetTeams,
    axiosQuizGrantTurn,
    axiosQuizJuryDecision,
    axiosQuizListButtons,
    axiosQuizPauseTimer,
    axiosQuizRegisterButton,
    axiosQuizResetQuestion,
    axiosQuizResetTournament,
    axiosQuizResumeTimer,
    axiosQuizRevealAnswer,
    axiosQuizSetOrder,
    axiosQuizSetupTeams,
    axiosQuizShowLeaderboard,
    axiosQuizStartQuestion,
    axiosQuizStartRound,
    axiosQuizStartTiebreak,
    axiosQuizStartTiebreakQuestion,
} from "../../requests/quizRequests";
import { BuzzStatus, type QuizButtonAdmin, type QuizQuestion, type QuizTeamAdmin } from "../../types/quizTypes";
import "../../assets/Quiz.css";

type AdminTab = "live" | "teams" | "setup";

interface ConfirmConfig {
    title: string;
    body: string;
    confirmLabel: string;
    variant: "primary" | "success" | "warning" | "danger";
    onConfirm: () => void;
}

const ROUND_META: Record<number, { defaultName: string; mechanic: string; defaultCount: number }> = {
    1: { defaultName: "Стрелы логики", mechanic: "Брейн-ринг · 60 с", defaultCount: 10 },
    2: { defaultName: "Вектор удара", mechanic: "ЧГК · 60 с", defaultCount: 7 },
    3: { defaultName: "Формула клинка", mechanic: "Капитаны · 120 с", defaultCount: 5 },
};

const ORDER_OPTIONS: [number, string][] = [
    [0, "Не выбран"],
    [1, "🔴 Красный (5 б. / 0 ош.)"],
    [2, "🟡 Жёлтый (3 б. / 1 ош.)"],
    [3, "🟢 Зелёный (1 б. / 2 ош.)"],
];

function formatTime(totalSeconds: number): string {
    const m = Math.floor(totalSeconds / 60);
    const s = totalSeconds % 60;
    return `${m}:${s.toString().padStart(2, "0")}`;
}

function teamTotal(t: QuizTeamAdmin): number {
    return t.round1_score + t.round2_score + t.round3_score;
}

/** Editable round score cell: Enter applies, Escape resets. */
function AdjustCell({
    value,
    adjust,
    disabled,
    onApply,
}: {
    value: number;
    adjust: number;
    disabled?: boolean;
    onApply: (delta: number) => void;
}) {
    const [draft, setDraft] = useState(String(value));

    useEffect(() => setDraft(String(value)), [value]);

    const parsed = Number(draft);
    const valid = draft.trim() !== "" && !Number.isNaN(parsed);
    const changed = valid && parsed !== value;

    const commit = () => {
        if (changed && !disabled) {
            onApply(parsed - value);
        }
    };

    return (
        <div>
            <div className="quiz-admin-adjust">
                <input
                    type="number"
                    className="form-control form-control-sm quiz-admin-adjust-input"
                    value={draft}
                    disabled={disabled}
                    onChange={(e) => setDraft(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === "Enter") commit();
                        if (e.key === "Escape") setDraft(String(value));
                    }}
                />
                <Button
                    size="sm"
                    variant={changed ? "primary" : "outline-secondary"}
                    disabled={!changed || disabled}
                    onClick={commit}
                >
                    OK
                </Button>
            </div>
            {adjust !== 0 && (
                <div className="quiz-admin-adjust-note">
                    коррекция {adjust > 0 ? "+" : ""}
                    {adjust}
                </div>
            )}
        </div>
    );
}

/** Jury decision cell: sets the absolute tiebreak score of a team (place rank). */
function JuryCell({
    value,
    disabled,
    onApply,
}: {
    value: number;
    disabled?: boolean;
    onApply: (score: number) => void;
}) {
    const [draft, setDraft] = useState(String(value));

    useEffect(() => setDraft(String(value)), [value]);

    const parsed = Number(draft);
    const valid = draft.trim() !== "" && !Number.isNaN(parsed) && parsed >= 0;
    const changed = valid && parsed !== value;

    const commit = () => {
        if (changed && !disabled) {
            onApply(parsed);
        }
    };

    return (
        <div className="quiz-admin-adjust">
            <input
                type="number"
                min={0}
                className="form-control form-control-sm quiz-admin-adjust-input"
                value={draft}
                disabled={disabled}
                onChange={(e) => setDraft(e.target.value)}
                onKeyDown={(e) => {
                    if (e.key === "Enter") commit();
                    if (e.key === "Escape") setDraft(String(value));
                }}
            />
            <Button
                size="sm"
                variant={changed ? "primary" : "outline-secondary"}
                disabled={!changed || disabled}
                onClick={commit}
            >
                OK
            </Button>
        </div>
    );
}

/** Round-3 order selector; applies immediately on change (admin override). */
function OrderCell({
    value,
    disabled,
    onChange,
}: {
    value: number;
    disabled?: boolean;
    onChange: (order: string) => void;
}) {
    return (
        <select
            className={`form-select form-select-sm quiz-admin-order-select order-${value}`}
            value={value}
            disabled={disabled}
            onChange={(e) => onChange(e.target.value)}
        >
            {ORDER_OPTIONS.map(([v, label]) => (
                <option key={v} value={v}>
                    {label}
                </option>
            ))}
        </select>
    );
}

export default function QuizAdmin() {
    const { eventId } = useParams<{ eventId: string }>();
    const navigate = useNavigate();
    const { state } = useQuizState(eventId);

    const [activeTab, setActiveTab] = useState<AdminTab>("live");
    const [teams, setTeams] = useState<QuizTeamAdmin[]>([]);
    const [questions, setQuestions] = useState<QuizQuestion[]>([]);
    const [buttons, setButtons] = useState<QuizButtonAdmin[]>([]);
    const [buttonDraft, setButtonDraft] = useState({ team_id: "", button_id: "" });
    const [teamSearch, setTeamSearch] = useState("");
    const [showAllColumns, setShowAllColumns] = useState(false);
    const [tiebreakSelection, setTiebreakSelection] = useState<string[]>([]);

    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");
    const [confirmModal, setConfirmModal] = useState<ConfirmConfig | null>(null);

    const [remainingMs, setRemainingMs] = useState(0);
    const deadlineRef = useRef(0);
    const runningRef = useRef(false);

    const loadTeams = useCallback(() => {
        if (!eventId) return;
        axiosQuizGetTeams(eventId).then(setTeams).catch(() => setTeams([]));
    }, [eventId]);

    const loadButtons = useCallback(() => {
        if (!eventId) return;
        axiosQuizListButtons(eventId).then(setButtons).catch(() => setButtons([]));
    }, [eventId]);

    const loadQuestions = useCallback(() => {
        if (!eventId) return;
        axiosQuizGetQuestions(eventId).then(setQuestions).catch(() => setQuestions([]));
    }, [eventId]);

    useEffect(() => {
        loadTeams();
        loadQuestions();
        loadButtons();
    }, [loadTeams, loadQuestions, loadButtons]);

    // Refresh teams when switching to Round 3 or Teams tab so order/score statuses stay fresh
    useEffect(() => {
        if (activeTab === "teams" || state?.current_round === 3) {
            loadTeams();
        }
    }, [activeTab, state?.current_round, state?.current_question, loadTeams]);

    // Smooth local timer between server broadcasts
    useEffect(() => {
        if (!state) return;
        const t = state.timer;
        if (!t.running) {
            runningRef.current = false;
            deadlineRef.current = Date.now() + t.remaining_ms;
            setRemainingMs(t.remaining_ms);
            return;
        }
        const localRem = Math.max(deadlineRef.current - Date.now(), 0);
        if (!runningRef.current || Math.abs(t.remaining_ms - localRem) > 1000) {
            deadlineRef.current = Date.now() + t.remaining_ms;
            setRemainingMs(t.remaining_ms);
        }
        runningRef.current = true;
    }, [state]);

    useEffect(() => {
        const tick = setInterval(() => {
            if (!runningRef.current) return;
            setRemainingMs(Math.max(deadlineRef.current - Date.now(), 0));
        }, 250);
        return () => clearInterval(tick);
    }, []);

    const run = async (fn: () => Promise<void>, successMsg: string) => {
        if (busy) return;
        setBusy(true);
        setError("");
        try {
            await fn();
            setNotice(successMsg);
            loadTeams();
            loadButtons();
        } catch (e: unknown) {
            setError(e instanceof Error ? e.message : "Ошибка выполнения запроса");
        } finally {
            setBusy(false);
        }
    };

    const filteredTeams = useMemo(() => {
        const q = teamSearch.trim().toLowerCase();
        if (!q) return teams;
        return teams.filter(
            (t) => t.name.toLowerCase().includes(q) || t.access_code.toLowerCase().includes(q),
        );
    }, [teams, teamSearch]);

    if (!eventId) return null;

    const currentRound = state?.current_round ?? 0;
    const currentQuestion = state?.current_question ?? 0;
    const roundCfg = state?.rounds.find((r) => r.round === currentRound);
    const hasActiveQuestion = !!roundCfg && currentQuestion > 0;
    const nextQuestionNumber = currentQuestion + 1;
    const nextQuestionExists = roundCfg ? nextQuestionNumber <= roundCfg.question_count : false;
    const isRoundFinished = !!roundCfg && currentQuestion >= roundCfg.question_count;

    const currentQuestionData =
        state && roundCfg && currentQuestion > 0
            ? questions.find((q) => q.round === currentRound && q.question_number === currentQuestion)
            : undefined;

    const timerRunning = !!state?.timer.running;
    const timerStarted = !!state?.timer.started;
    const timerState = !roundCfg
        ? "idle"
        : !timerRunning
          ? "paused"
          : remainingMs <= 15_000
            ? "warning"
            : "running";
    const timerLabel =
        timerState === "running"
            ? "Время идёт"
            : timerState === "warning"
              ? "Осталось мало времени"
              : timerState === "paused"
                ? hasActiveQuestion
                    ? timerStarted
                        ? "Пауза"
                        : "Вопрос на экране — таймер не запущен"
                    : "Готов к старту"
                : "Тур не начат";

    const teamsTotal = teams.reduce((sum, t) => sum + teamTotal(t), 0);
    const activeBuzz = state?.buzz_queue.find((b) => b.is_active);

    // Tie-break (переигровка)
    const tiebreakActive = !!state?.tiebreak_active;
    const tiebreakQuestionNumber = state?.tiebreak_question ?? 0;
    const tiebreakTeams = state?.teams.filter((t) => t.in_tiebreak) ?? [];
    const reserveQuestions = questions.filter((q) => q.round === 4);
    const usedReserveCount = reserveQuestions.filter((q) => q.used).length;
    const currentReserveQuestion =
        tiebreakActive && tiebreakQuestionNumber > 0
            ? questions.find((q) => q.round === 4 && q.question_number === tiebreakQuestionNumber)
            : undefined;
    const teamsByTotal = [...teams].sort((a, b) => teamTotal(b) - teamTotal(a));
    const toggleTiebreakTeam = (teamId: string) =>
        setTiebreakSelection((prev) =>
            prev.includes(teamId) ? prev.filter((id) => id !== teamId) : [...prev, teamId],
        );

    // Round 3 readiness stats
    const round3ReadyCount = teams.filter((t) => t.round3_order !== 0).length;
    const round3PendingTeams = teams.filter((t) => t.round3_order === 0);
    const round3EliminatedCount = teams.filter((t) => t.round3_eliminated).length;
    const round3OrderCounts = {
        red: teams.filter((t) => t.round3_order === 1).length,
        yellow: teams.filter((t) => t.round3_order === 2).length,
        green: teams.filter((t) => t.round3_order === 3).length,
    };

    // Screen status indicator for the projector
    const screenStatus = state?.tiebreak_active
        ? {
              label:
                  tiebreakQuestionNumber > 0
                      ? `Переигровка · резервный вопрос ${tiebreakQuestionNumber}`
                      : "Переигровка · ожидание вопроса",
              tone: "warning" as const,
          }
        : state?.leaderboard_shown
        ? { label: "Таблица лидеров", tone: "info" as const }
        : !roundCfg
          ? { label: "Заставка / ожидание", tone: "secondary" as const }
          : !hasActiveQuestion
            ? { label: `Заголовок Тура ${roundCfg.round}`, tone: "secondary" as const }
            : state?.answer_revealed
              ? { label: `Вопрос ${currentQuestion} · Ответ показан`, tone: "success" as const }
              : { label: `Вопрос ${currentQuestion} · Ответ скрыт`, tone: "primary" as const };

    const requestStartRound = (targetRound: number) => {
        const targetInfo = state?.rounds.find((r) => r.round === targetRound);
        const targetName = targetInfo?.name ?? ROUND_META[targetRound].defaultName;

        if (currentRound === targetRound) {
            setConfirmModal({
                title: `Перезапустить Тур ${targetRound}?`,
                body: `Тур ${targetRound} «${targetName}» уже идёт (вопрос ${currentQuestion} из ${roundCfg?.question_count ?? "?"}). Повторный запуск сбросит текущий прогресс тура.`,
                confirmLabel: "Да, перезапустить",
                variant: "warning",
                onConfirm: () =>
                    run(() => axiosQuizStartRound(eventId, targetRound), `Тур ${targetRound} перезапущен`),
            });
            return;
        }

        const unfinishedWarning =
            roundCfg && !isRoundFinished
                ? ` В текущем Туре ${currentRound} сыграно только ${currentQuestion} из ${roundCfg.question_count} вопросов.`
                : "";

        setConfirmModal({
            title: `Начать Тур ${targetRound} — «${targetName}»?`,
            body: `Экраны герольдов и проектор переключатся на механику «${ROUND_META[targetRound].mechanic}».${unfinishedWarning}`,
            confirmLabel: `Начать Тур ${targetRound}`,
            variant: "primary",
            onConfirm: () => {
                setActiveTab("live");
                run(() => axiosQuizStartRound(eventId, targetRound), `Тур ${targetRound} начат`);
            },
        });
    };

    const handleStartNextQuestion = () => {
        // Warn before starting Round 3 Q1 if some teams haven't selected an order
        if (currentRound === 3 && currentQuestion === 0 && round3PendingTeams.length > 0) {
            setConfirmModal({
                title: "Не все команды выбрали орден",
                body: `Орден выбрали ${round3ReadyCount} из ${teams.length} команд. Ещё не определились: ${round3PendingTeams
                    .slice(0, 5)
                    .map((t) => t.name)
                    .join(", ")}${round3PendingTeams.length > 5 ? "…" : ""}. Запустить вопрос 1 всё равно?`,
                confirmLabel: "Запустить вопрос 1",
                variant: "warning",
                onConfirm: () =>
                    run(() => axiosQuizStartQuestion(eventId), `Вопрос ${nextQuestionNumber} запущен`),
            });
            return;
        }

        run(() => axiosQuizStartQuestion(eventId), `Вопрос ${nextQuestionNumber} запущен`);
    };

    const copyAccessCode = (code: string, teamName: string) => {
        navigator.clipboard?.writeText(code).then(
            () => setNotice(`Код команды «${teamName}» скопирован: ${code}`),
            () => {},
        );
    };

    const requestResetTournament = () => {
        setConfirmModal({
            title: "Сбросить турнир полностью?",
            body: "Будут удалены ВСЕ очки команд (включая ручные поправки), все ответы и нажатия кнопок, данные переигровки, а прогресс туров вернётся к началу. Резервные вопросы снова станут доступны. Команды, их коды доступа и вопросы сохранятся. Действие необратимо!",
            confirmLabel: "Да, сбросить всё",
            variant: "danger",
            onConfirm: () =>
                run(() => axiosQuizResetTournament(eventId), "Турнир сброшен: все очки и прогресс удалены"),
        });
    };

    const showOrderColumn = currentRound === 3 || showAllColumns;
    const showGrantTurnColumn = currentRound === 1 || showAllColumns;

    return (
        <div className="quiz-admin-root quiz-admin-v2">
            {/* Fixed Toast notifications (no layout jump) */}
            <ToastContainer position="top-end" className="p-3 quiz-admin-toasts">
                <Toast
                    show={!!notice}
                    onClose={() => setNotice("")}
                    delay={3200}
                    autohide
                    bg="success"
                    className="text-white"
                >
                    <Toast.Body className="d-flex justify-content-between align-items-center gap-2">
                        <span>{notice}</span>
                        <button
                            type="button"
                            className="btn-close btn-close-white btn-sm"
                            aria-label="Закрыть"
                            onClick={() => setNotice("")}
                        />
                    </Toast.Body>
                </Toast>
                <Toast show={!!error} onClose={() => setError("")} bg="danger" className="text-white">
                    <Toast.Body className="d-flex justify-content-between align-items-center gap-2">
                        <span>{error}</span>
                        <button
                            type="button"
                            className="btn-close btn-close-white btn-sm"
                            aria-label="Закрыть"
                            onClick={() => setError("")}
                        />
                    </Toast.Body>
                </Toast>
            </ToastContainer>

            {/* Top Header + Projector Status */}
            <header className="quiz-admin-topbar">
                <div>
                    <h2 className="mb-1">Панель ведущего — квиз-шоу</h2>
                    <div className="quiz-admin-screen-pill">
                        <span className={`quiz-status-dot dot-${screenStatus.tone}`} />
                        <span className="text-body-secondary">На проекторе:</span>
                        <strong>{screenStatus.label}</strong>
                    </div>
                </div>
                <div className="quiz-admin-topbar-actions">
                    <Button
                        variant={state?.leaderboard_shown ? "info" : "outline-dark"}
                        disabled={busy}
                        onClick={() =>
                            run(
                                () => axiosQuizShowLeaderboard(eventId, !state?.leaderboard_shown),
                                state?.leaderboard_shown
                                    ? "Таблица лидеров скрыта с проектора"
                                    : "Таблица лидеров выведена на проектор",
                            )
                        }
                    >
                        {state?.leaderboard_shown
                            ? "Скрыть табло с проектора"
                            : "Вывести табло на проектор"}
                    </Button>
                    <Button
                        variant="outline-danger"
                        disabled={busy}
                        onClick={requestResetTournament}
                    >
                        Сбросить турнир
                    </Button>
                </div>
            </header>

            {/* Round Stepper (Timeline) */}
            <section className="quiz-round-stepper" aria-label="Этапы турнира">
                {[1, 2, 3].map((rNum) => {
                    const cfg = state?.rounds.find((r) => r.round === rNum);
                    const meta = ROUND_META[rNum];
                    const name = cfg?.name ?? meta.defaultName;
                    const totalQ = cfg?.question_count ?? meta.defaultCount;
                    const isActive = currentRound === rNum;
                    const isPassed = currentRound > rNum;

                    return (
                        <button
                            key={rNum}
                            type="button"
                            disabled={busy}
                            className={`quiz-round-step${isActive ? " active" : ""}${isPassed ? " passed" : ""}`}
                            onClick={() => requestStartRound(rNum)}
                        >
                            <div className="quiz-round-step-head">
                                <span className="quiz-round-step-num">Тур {rNum}</span>
                                {isActive && <Badge bg="success">Идёт сейчас</Badge>}
                                {isPassed && <Badge bg="secondary">Завершён</Badge>}
                                {!isActive && !isPassed && (
                                    <Badge bg="light" text="dark">
                                        Ожидает
                                    </Badge>
                                )}
                            </div>
                            <div className="quiz-round-step-title">«{name}»</div>
                            <div className="quiz-round-step-meta">
                                <span>{meta.mechanic}</span>
                                <strong>
                                    {isActive ? `${currentQuestion} / ${totalQ}` : `${totalQ} вопр.`}
                                </strong>
                            </div>
                        </button>
                    );
                })}
            </section>

            {/* Mode Tabs */}
            <Nav
                variant="tabs"
                activeKey={activeTab}
                onSelect={(k) => k && setActiveTab(k as AdminTab)}
                className="quiz-admin-tabs"
            >
                <Nav.Item>
                    <Nav.Link eventKey="live">Пульт эфира</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                    <Nav.Link eventKey="teams">
                        Команды и баллы{" "}
                        <Badge bg="secondary" pill>
                            {teams.length}
                        </Badge>
                    </Nav.Link>
                </Nav.Item>
                <Nav.Item>
                    <Nav.Link eventKey="setup">
                        Настройка (ESP32 и банк)
                        {(teams.length === 0 || questions.length === 0) && (
                            <Badge bg="warning" text="dark" pill className="ms-1">
                                !
                            </Badge>
                        )}
                    </Nav.Link>
                </Nav.Item>
            </Nav>

            {/* ================= TAB 1: LIVE CONTROL ================= */}
            {activeTab === "live" && (
                <div className="quiz-admin-tab-body">
                    {!roundCfg ? (
                        <div className="quiz-admin-section liquid-glass-bg quiz-empty-stage">
                            <h4>Турнир ещё не запущен</h4>
                            <p className="text-body-secondary mb-3">
                                Проверьте готовность команд и кнопок во вкладке «Настройка», затем запустите первый тур.
                            </p>
                            <Button
                                size="lg"
                                variant="primary"
                                disabled={busy}
                                onClick={() => requestStartRound(1)}
                            >
                                Начать Тур 1 — «{ROUND_META[1].defaultName}»
                            </Button>
                        </div>
                    ) : (
                        <>
                            {/* Unified Timer & Pause Bar */}
                            <div className={`quiz-admin-timer-bar ${timerState}`}>
                                <div className="quiz-admin-timer-left">
                                    <span className="quiz-admin-timer-value">
                                        {formatTime(Math.ceil(remainingMs / 1000))}
                                    </span>
                                    <div className="quiz-admin-timer-meta">
                                        <span className="quiz-admin-timer-state">{timerLabel}</span>
                                        <span className="text-body-secondary small">
                                            Тур {roundCfg.round} «{roundCfg.name}» · Вопрос {currentQuestion} из{" "}
                                            {roundCfg.question_count}
                                        </span>
                                    </div>
                                </div>

                                <div className="quiz-admin-timer-controls">
                                    <Button
                                        variant={timerRunning ? "warning" : "success"}
                                        disabled={busy || !hasActiveQuestion}
                                        onClick={() =>
                                            run(
                                                () =>
                                                    timerRunning
                                                        ? axiosQuizPauseTimer(eventId)
                                                        : axiosQuizResumeTimer(eventId),
                                                timerRunning
                                                    ? "Таймер на паузе"
                                                    : timerStarted
                                                      ? "Таймер продолжен"
                                                      : "Таймер запущен",
                                            )
                                        }
                                    >
                                        {timerRunning
                                            ? "⏸ Пауза"
                                            : timerStarted
                                              ? "▶ Продолжить"
                                              : "▶ Старт таймера"}
                                    </Button>
                                    <Button
                                        variant="outline-secondary"
                                        disabled={busy || !hasActiveQuestion}
                                        onClick={() =>
                                            run(() => axiosQuizCloseQuestion(eventId), "Вопрос закрыт")
                                        }
                                    >
                                        ⏹ Закрыть вопрос
                                    </Button>
                                </div>
                            </div>

                            {/* Main 2-column Workspace */}
                            <div className="quiz-live-grid">
                                {/* Left: Teleprompter (Question + Answer) */}
                                <div className="quiz-admin-section liquid-glass-bg quiz-teleprompter">
                                    <div className="quiz-section-header">
                                        <h5 className="mb-0">
                                            {hasActiveQuestion
                                                ? `Вопрос ${currentQuestion} из ${roundCfg.question_count}`
                                                : `Тур ${roundCfg.round} готов к запуску`}
                                        </h5>
                                        {hasActiveQuestion && (
                                            <Badge bg={state?.answer_revealed ? "success" : "secondary"}>
                                                {state?.answer_revealed
                                                    ? "Ответ показан залу"
                                                    : "Ответ скрыт от зала"}
                                            </Badge>
                                        )}
                                    </div>

                                    {!hasActiveQuestion ? (
                                        <div className="quiz-teleprompter-empty">
                                            Нажмите <strong>«Запустить вопрос 1»</strong> ниже, когда зал и герольды
                                            будут готовы.
                                        </div>
                                    ) : currentQuestionData ? (
                                        <>
                                            <div className="quiz-question-text">{currentQuestionData.text}</div>
                                            <div
                                                className={`quiz-answer-box${
                                                    state?.answer_revealed ? " revealed" : ""
                                                }`}
                                            >
                                                <div className="quiz-answer-label">
                                                    Правильный ответ{" "}
                                                    {state?.answer_revealed
                                                        ? "(выведен на проектор)"
                                                        : "(видите только вы)"}
                                                </div>
                                                <div className="quiz-answer-value">
                                                    {currentQuestionData.answer}
                                                </div>
                                                {currentQuestionData.explanation && (
                                                    <div className="quiz-explanation-value">
                                                        {currentQuestionData.explanation}
                                                    </div>
                                                )}
                                                {currentQuestionData.criteria && (
                                                    <div className="quiz-criteria-value">
                                                        <span className="quiz-criteria-label">Критерии жюри</span>
                                                        {currentQuestionData.criteria}
                                                    </div>
                                                )}
                                            </div>
                                        </>
                                    ) : (
                                        <div className="quiz-msg error">
                                            Для вопроса {currentQuestion} нет текста в банке — проверьте загрузку
                                            вопросов.
                                        </div>
                                    )}
                                </div>

                                {/* Right: Round-Specific Live Context */}
                                <div className="quiz-admin-section liquid-glass-bg quiz-live-side">
                                    {currentRound === 1 && (
                                        <>
                                            <div className="quiz-section-header">
                                                <h5 className="mb-0">Очередь нажатий</h5>
                                                {hasActiveQuestion && (
                                                    <div className="quiz-attempts-dots" title="Использовано попыток">
                                                        {Array.from({ length: state?.attempts_max ?? 3 }).map(
                                                            (_, idx) => (
                                                                <span
                                                                    key={idx}
                                                                    className={`quiz-attempt-dot${
                                                                        idx < (state?.attempts_used ?? 0)
                                                                            ? " used"
                                                                            : ""
                                                                    }`}
                                                                />
                                                            ),
                                                        )}
                                                        <span className="small text-body-secondary ms-1">
                                                            {state?.attempts_used ?? 0}/{state?.attempts_max ?? 3}
                                                        </span>
                                                    </div>
                                                )}
                                            </div>

                                            {activeBuzz ? (
                                                <div className="quiz-active-responder">
                                                    <span className="quiz-active-responder-label">
                                                        Сейчас отвечает
                                                    </span>
                                                    <strong className="quiz-active-responder-name">
                                                        {activeBuzz.team_name}
                                                    </strong>
                                                </div>
                                            ) : (
                                                <div className="quiz-side-placeholder">
                                                    {hasActiveQuestion
                                                        ? "Ожидание нажатия кнопки командой…"
                                                        : "Кнопки активируются после запуска вопроса."}
                                                </div>
                                            )}

                                            {state && state.buzz_queue.length > 0 && (
                                                <div className="quiz-queue-list">
                                                    {state.buzz_queue.map((b, i) => (
                                                        <div
                                                            key={b.team_id}
                                                            className={`quiz-queue-chip${
                                                                b.is_active ? " active" : ""
                                                            }${b.status === BuzzStatus.Wrong ? " wrong" : ""}${
                                                                b.status === BuzzStatus.Correct ? " correct" : ""
                                                            }${b.status === BuzzStatus.Outpaced ? " outpaced" : ""}`}
                                                        >
                                                            <span>
                                                                <strong>{i + 1}.</strong> {b.team_name}
                                                            </span>
                                                            {i > 0 && (
                                                                <span className="small opacity-75">
                                                                    +{b.offset_ms} мс
                                                                </span>
                                                            )}
                                                        </div>
                                                    ))}
                                                </div>
                                            )}

                                            <div className="quiz-side-actions">
                                                <Button
                                                    size="sm"
                                                    variant="outline-warning"
                                                    disabled={busy || !hasActiveQuestion}
                                                    onClick={() =>
                                                        run(
                                                            () => axiosQuizResetQuestion(eventId),
                                                            "Очередь нажатий сброшена",
                                                        )
                                                    }
                                                >
                                                    Сбросить очередь
                                                </Button>
                                                <Button
                                                    size="sm"
                                                    variant="link"
                                                    className="px-1"
                                                    onClick={() => setActiveTab("teams")}
                                                >
                                                    Дать право вручную →
                                                </Button>
                                            </div>
                                        </>
                                    )}

                                    {currentRound === 2 && (
                                        <>
                                            <div className="quiz-section-header">
                                                <h5 className="mb-0">Регламент Тура 2</h5>
                                                <Badge bg="light" text="dark">
                                                    ЧГК
                                                </Badge>
                                            </div>
                                            <ul className="quiz-rules-checklist">
                                                <li>Все команды пишут ответы на бумажных бланках одновременно.</li>
                                                <li>После сигнала выведите ответ на проектор.</li>
                                                <li>Герольды выставляют оценку: 4 / 2 / 0 баллов.</li>
                                            </ul>
                                            <div className="quiz-mini-stats">
                                                <div>
                                                    <span className="text-body-secondary small">Команд</span>
                                                    <strong>{teams.length}</strong>
                                                </div>
                                                <div>
                                                    <span className="text-body-secondary small">Всего очков</span>
                                                    <strong>{teamsTotal}</strong>
                                                </div>
                                            </div>
                                        </>
                                    )}

                                    {currentRound === 3 && (
                                        <>
                                            <div className="quiz-section-header">
                                                <h5 className="mb-0">Статус капитанов</h5>
                                                <Button
                                                    size="sm"
                                                    variant="outline-secondary"
                                                    onClick={loadTeams}
                                                    disabled={busy}
                                                >
                                                    Обновить
                                                </Button>
                                            </div>

                                            <div className="mb-2">
                                                <div className="d-flex justify-content-between small mb-1">
                                                    <span>Выбрали орден</span>
                                                    <strong>
                                                        {round3ReadyCount} из {teams.length}
                                                    </strong>
                                                </div>
                                                <ProgressBar
                                                    now={teams.length ? (round3ReadyCount / teams.length) * 100 : 0}
                                                    variant={
                                                        round3ReadyCount === teams.length && teams.length > 0
                                                            ? "success"
                                                            : "primary"
                                                    }
                                                    style={{ height: 8 }}
                                                />
                                            </div>

                                            <div className="quiz-order-summary">
                                                <span>🔴 {round3OrderCounts.red}</span>
                                                <span>🟡 {round3OrderCounts.yellow}</span>
                                                <span>🟢 {round3OrderCounts.green}</span>
                                                <span className="text-danger">
                                                    Выбыли: {round3EliminatedCount}
                                                </span>
                                            </div>

                                            {round3PendingTeams.length > 0 ? (
                                                <div className="quiz-pending-orders">
                                                    <div className="small text-body-secondary mb-1">
                                                        Ещё не выбрали орден:
                                                    </div>
                                                    <div className="small fw-medium">
                                                        {round3PendingTeams.map((t) => t.name).join(", ")}
                                                    </div>
                                                </div>
                                            ) : (
                                                teams.length > 0 && (
                                                    <div className="text-success small fw-medium mt-2">
                                                        ✓ Все капитаны зафиксировали свой орден.
                                                    </div>
                                                )
                                            )}
                                        </>
                                    )}
                                </div>
                            </div>

                            {/* Primary Action Bar (Next Best Action) */}
                            <div className="quiz-action-bar liquid-glass-bg">
                                <div className="quiz-action-bar-group">
                                    <Button
                                        size="lg"
                                        variant={
                                            hasActiveQuestion && !state?.answer_revealed
                                                ? "success"
                                                : "outline-success"
                                        }
                                        disabled={busy || !hasActiveQuestion || !!state?.answer_revealed}
                                        onClick={() =>
                                            run(
                                                () => axiosQuizRevealAnswer(eventId),
                                                "Правильный ответ выведен на проектор",
                                            )
                                        }
                                    >
                                        {state?.answer_revealed
                                            ? "✓ Ответ показан"
                                            : "Показать ответ на экране"}
                                    </Button>
                                </div>

                                <div className="quiz-action-bar-group">
                                    {nextQuestionExists ? (
                                        <Button
                                            size="lg"
                                            variant={
                                                !hasActiveQuestion || state?.answer_revealed
                                                    ? "primary"
                                                    : "outline-primary"
                                            }
                                            disabled={busy}
                                            onClick={handleStartNextQuestion}
                                        >
                                            {busy && <Spinner size="sm" className="me-2" />}
                                            Запустить вопрос {nextQuestionNumber} →
                                        </Button>
                                    ) : (
                                        <>
                                            {!state?.leaderboard_shown && (
                                                <Button
                                                    size="lg"
                                                    variant="info"
                                                    disabled={busy}
                                                    onClick={() =>
                                                        run(
                                                            () => axiosQuizShowLeaderboard(eventId, true),
                                                            "Итоги тура выведены на проектор",
                                                        )
                                                    }
                                                >
                                                    Показать итоги Тура {currentRound}
                                                </Button>
                                            )}
                                            {currentRound < 3 && (
                                                <Button
                                                    size="lg"
                                                    variant="primary"
                                                    disabled={busy}
                                                    onClick={() => requestStartRound(currentRound + 1)}
                                                >
                                                    Перейти к Туру {currentRound + 1} →
                                                </Button>
                                            )}
                                        </>
                                    )}
                                </div>
                            </div>

                            {/* Tie-break (переигровка) */}
                            <div className="quiz-admin-section liquid-glass-bg mt-3">
                                <div className="quiz-section-header">
                                    <h5 className="mb-0">Переигровка за призовые места</h5>
                                    <Badge bg={tiebreakActive ? "warning" : "secondary"}>
                                        {tiebreakActive ? "Идёт" : "Не активна"}
                                    </Badge>
                                </div>

                                {!tiebreakActive ? (
                                    <>
                                        <div className="small text-body-secondary mb-2">
                                            Выберите команды, делящие одно призовое место (минимум две).
                                            Раундовые баллы при переигровке не изменяются — результат идёт
                                            только в дополнительный критерий сортировки итоговой таблицы.
                                        </div>
                                        <div className="quiz-tiebreak-team-picker">
                                            {teamsByTotal.map((t) => (
                                                <Form.Check
                                                    key={t.team_id}
                                                    id={`tb-${t.team_id}`}
                                                    label={`${t.name} — ${teamTotal(t)} б.`}
                                                    checked={tiebreakSelection.includes(t.team_id)}
                                                    onChange={() => toggleTiebreakTeam(t.team_id)}
                                                    disabled={busy}
                                                />
                                            ))}
                                        </div>
                                        <Button
                                            variant="warning"
                                            disabled={busy || tiebreakSelection.length < 2}
                                            onClick={() =>
                                                run(
                                                    () =>
                                                        axiosQuizStartTiebreak(
                                                            eventId,
                                                            tiebreakSelection,
                                                        ),
                                                    "Переигровка начата",
                                                ).then(() => setTiebreakSelection([]))
                                            }
                                        >
                                            Начать переигровку ({tiebreakSelection.length})
                                        </Button>
                                    </>
                                ) : (
                                    <div className="row g-3">
                                        <div className="col-lg-6">
                                            <div className="small text-body-secondary mb-1">
                                                Участники (ещё не заняли место):
                                            </div>
                                            {tiebreakTeams.length > 0 ? (
                                                <div className="fw-medium mb-2">
                                                    {tiebreakTeams.map((t) => t.name).join(", ")}
                                                </div>
                                            ) : (
                                                <div className="text-success small fw-medium mb-2">
                                                    ✓ Все места в группе разыграны — можно завершать
                                                    переигровку.
                                                </div>
                                            )}

                                            <div className="small text-body-secondary mb-1">
                                                Резервные вопросы: использовано {usedReserveCount} из{" "}
                                                {reserveQuestions.length}
                                            </div>

                                            {currentReserveQuestion ? (
                                                <div className="quiz-answer-box revealed mb-2">
                                                    <div className="quiz-answer-label">
                                                        Резервный вопрос №
                                                        {currentReserveQuestion.question_number}
                                                    </div>
                                                    <div className="quiz-question-text">
                                                        {currentReserveQuestion.text}
                                                    </div>
                                                    <div className="quiz-answer-value">
                                                        {currentReserveQuestion.answer}
                                                    </div>
                                                    {currentReserveQuestion.explanation && (
                                                        <div className="quiz-explanation-value">
                                                            {currentReserveQuestion.explanation}
                                                        </div>
                                                    )}
                                                </div>
                                            ) : (
                                                <div className="quiz-side-placeholder mb-2">
                                                    Резервный вопрос ещё не запущен.
                                                </div>
                                            )}

                                            <div className="d-flex gap-2 flex-wrap">
                                                <Button
                                                    variant="primary"
                                                    disabled={busy || tiebreakTeams.length === 0}
                                                    onClick={() =>
                                                        run(
                                                            () =>
                                                                axiosQuizStartTiebreakQuestion(eventId),
                                                            "Резервный вопрос запущен",
                                                        )
                                                    }
                                                >
                                                    {tiebreakQuestionNumber > 0
                                                        ? "Следующий резервный вопрос →"
                                                        : "Запустить резервный вопрос"}
                                                </Button>
                                                <Button
                                                    variant="outline-danger"
                                                    disabled={busy}
                                                    onClick={() =>
                                                        setConfirmModal({
                                                            title: "Завершить переигровку?",
                                                            body: "Режим переигровки будет закрыт, группа участников очищена. Начисленные tiebreak-баллы сохранятся.",
                                                            confirmLabel: "Завершить",
                                                            variant: "danger",
                                                            onConfirm: () =>
                                                                run(
                                                                    () => axiosQuizEndTiebreak(eventId),
                                                                    "Переигровка завершена",
                                                                ),
                                                        })
                                                    }
                                                >
                                                    Завершить переигровку
                                                </Button>
                                            </div>
                                        </div>

                                        <div className="col-lg-6">
                                            <div className="small text-body-secondary mb-1">
                                                Очередь нажатий:
                                            </div>
                                            {state && state.buzz_queue.length > 0 ? (
                                                <div className="quiz-queue-list mb-2">
                                                    {state.buzz_queue.map((b, i) => (
                                                        <div
                                                            key={b.team_id}
                                                            className={`quiz-queue-chip${
                                                                b.is_active ? " active" : ""
                                                            }${
                                                                b.status === BuzzStatus.Wrong
                                                                    ? " wrong"
                                                                    : ""
                                                            }${
                                                                b.status === BuzzStatus.Correct
                                                                    ? " correct"
                                                                    : ""
                                                            }${
                                                                b.status === BuzzStatus.Outpaced
                                                                    ? " outpaced"
                                                                    : ""
                                                            }`}
                                                        >
                                                            <span>
                                                                <strong>{i + 1}.</strong> {b.team_name}
                                                            </span>
                                                            {i > 0 && (
                                                                <span className="small opacity-75">
                                                                    +{b.offset_ms} мс
                                                                </span>
                                                            )}
                                                        </div>
                                                    ))}
                                                </div>
                                            ) : (
                                                <div className="quiz-side-placeholder mb-2">
                                                    Ожидание нажатия…
                                                </div>
                                            )}

                                            <div className="small text-body-secondary mb-1">
                                                Решение жюри (tiebreak-баллы; больше = выше место):
                                            </div>
                                            {(state?.teams ?? [])
                                                .filter((t) => t.in_tiebreak || t.tiebreak_score > 0)
                                                .map((t) => (
                                                    <div
                                                        key={t.team_id}
                                                        className="d-flex align-items-center gap-2 mb-1"
                                                    >
                                                        <div className="flex-grow-1">
                                                            {t.name}{" "}
                                                            <span className="text-body-secondary small">
                                                                ({t.total_score} б.)
                                                            </span>
                                                        </div>
                                                        <JuryCell
                                                            value={t.tiebreak_score}
                                                            disabled={busy}
                                                            onApply={(score) =>
                                                                run(
                                                                    () =>
                                                                        axiosQuizJuryDecision(
                                                                            eventId,
                                                                            t.team_id,
                                                                            score,
                                                                        ),
                                                                    `Решение жюри применено: ${t.name}`,
                                                                )
                                                            }
                                                        />
                                                    </div>
                                                ))}
                                        </div>
                                    </div>
                                )}
                            </div>
                        </>
                    )}
                </div>
            )}

            {/* ================= TAB 2: TEAMS & SCORES ================= */}
            {activeTab === "teams" && (
                <div className="quiz-admin-section liquid-glass-bg">
                    <div className="quiz-teams-toolbar">
                        <div className="d-flex align-items-center gap-3 flex-wrap">
                            <h4 className="mb-0">Команды и баллы</h4>
                            <span className="text-body-secondary small">
                                Команд: <strong>{teams.length}</strong> · Сумма очков:{" "}
                                <strong>{teamsTotal}</strong>
                            </span>
                        </div>
                        <div className="d-flex align-items-center gap-3 flex-wrap">
                            <Form.Check
                                type="switch"
                                id="show-all-cols"
                                label="Все спецколонки (Орден / Право ответа)"
                                checked={showAllColumns}
                                onChange={(e) => setShowAllColumns(e.target.checked)}
                                className="small mb-0"
                            />
                            <Form.Control
                                size="sm"
                                type="search"
                                placeholder="Поиск команды или кода…"
                                value={teamSearch}
                                onChange={(e) => setTeamSearch(e.target.value)}
                                style={{ width: 220 }}
                            />
                        </div>
                    </div>

                    <p className="text-body-secondary small mb-3">
                        Очки по раундам можно править вручную (Enter — применить). Коррекции сохраняются при
                        пересчётах. Нажмите на код герольда, чтобы скопировать его.
                    </p>

                    {filteredTeams.length > 0 ? (
                        <div className="table-responsive">
                            <Table striped bordered hover size="sm" className="quiz-admin-table align-middle">
                                <thead>
                                    <tr>
                                        <th>Команда</th>
                                        <th>Код герольда</th>
                                        <th>Тур 1</th>
                                        <th>Тур 2</th>
                                        <th>Тур 3</th>
                                        <th>Итог</th>
                                        {showOrderColumn && <th>Орден (Тур 3)</th>}
                                        {showGrantTurnColumn && <th>Брейн-ринг</th>}
                                    </tr>
                                </thead>
                                <tbody>
                                    {filteredTeams.map((t) => (
                                        <tr key={t.id}>
                                            <td>
                                                <strong>{t.name}</strong>
                                                {t.round3_eliminated && (
                                                    <Badge bg="danger" className="ms-2">
                                                        выбыла
                                                    </Badge>
                                                )}
                                            </td>
                                            <td>
                                                <button
                                                    type="button"
                                                    className="quiz-code-copy"
                                                    title="Нажмите, чтобы скопировать"
                                                    onClick={() => copyAccessCode(t.access_code, t.name)}
                                                >
                                                    {t.access_code}
                                                </button>
                                            </td>
                                            <td>
                                                <AdjustCell
                                                    value={t.round1_score}
                                                    adjust={t.round1_adjust}
                                                    disabled={busy}
                                                    onApply={(delta) =>
                                                        run(
                                                            () =>
                                                                axiosQuizAdjustScore(
                                                                    eventId,
                                                                    t.team_id,
                                                                    1,
                                                                    delta,
                                                                ),
                                                            `Тур 1, ${t.name}: ${delta > 0 ? "+" : ""}${delta}`,
                                                        )
                                                    }
                                                />
                                            </td>
                                            <td>
                                                <AdjustCell
                                                    value={t.round2_score}
                                                    adjust={t.round2_adjust}
                                                    disabled={busy}
                                                    onApply={(delta) =>
                                                        run(
                                                            () =>
                                                                axiosQuizAdjustScore(
                                                                    eventId,
                                                                    t.team_id,
                                                                    2,
                                                                    delta,
                                                                ),
                                                            `Тур 2, ${t.name}: ${delta > 0 ? "+" : ""}${delta}`,
                                                        )
                                                    }
                                                />
                                            </td>
                                            <td>
                                                <AdjustCell
                                                    value={t.round3_score}
                                                    adjust={t.round3_adjust}
                                                    disabled={busy}
                                                    onApply={(delta) =>
                                                        run(
                                                            () =>
                                                                axiosQuizAdjustScore(
                                                                    eventId,
                                                                    t.team_id,
                                                                    3,
                                                                    delta,
                                                                ),
                                                            `Тур 3, ${t.name}: ${delta > 0 ? "+" : ""}${delta}`,
                                                        )
                                                    }
                                                />
                                            </td>
                                            <td>
                                                <strong>{teamTotal(t)}</strong>
                                            </td>
                                            {showOrderColumn && (
                                                <td>
                                                    <OrderCell
                                                        value={t.round3_order}
                                                        disabled={busy}
                                                        onChange={(order) =>
                                                            run(
                                                                () =>
                                                                    axiosQuizSetOrder(
                                                                        eventId,
                                                                        t.team_id,
                                                                        order,
                                                                    ),
                                                                `Орден обновлён: ${t.name}`,
                                                            )
                                                        }
                                                    />
                                                </td>
                                            )}
                                            {showGrantTurnColumn && (
                                                <td>
                                                    <Button
                                                        size="sm"
                                                        variant="outline-warning"
                                                        disabled={busy}
                                                        onClick={() =>
                                                            run(
                                                                () => axiosQuizGrantTurn(eventId, t.team_id),
                                                                `Право ответа передано: ${t.name}`,
                                                            )
                                                        }
                                                    >
                                                        Дать право
                                                    </Button>
                                                </td>
                                            )}
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        </div>
                    ) : (
                        <div className="text-body-secondary">
                            {teams.length === 0
                                ? "Команды ещё не подготовлены — перейдите во вкладку «Настройка»."
                                : "По вашему запросу команды не найдены."}
                        </div>
                    )}
                </div>
            )}

            {/* ================= TAB 3: SETUP ================= */}
            {activeTab === "setup" && (
                <div className="quiz-setup-stack">
                    {/* Teams init */}
                    <div className="quiz-admin-section liquid-glass-bg">
                        <h4 className="mb-2">1. Подготовка команд</h4>
                        <p className="text-body-secondary small mb-3">
                            Создаёт записи команд для турнира и генерирует коды доступа для экранов герольдов.
                            Сейчас в турнире команд: <strong>{teams.length}</strong>.
                        </p>
                        <Button
                            variant={teams.length === 0 ? "primary" : "outline-primary"}
                            disabled={busy}
                            onClick={() =>
                                run(() => axiosQuizSetupTeams(eventId), "Команды подготовлены")
                            }
                        >
                            Подготовить / синхронизировать команды
                        </Button>
                    </div>

                    {/* ESP32 buttons */}
                    <div className="quiz-admin-section liquid-glass-bg">
                        <h4 className="mb-2">2. Кнопки команд (ESP32)</h4>
                        <p className="text-body-secondary small mb-3">
                            ID кнопки прошивается в устройство вместе с ID события. Здесь задаётся соответствие
                            «физическая кнопка → команда». Привязано: <strong>{buttons.length}</strong>.
                        </p>
                        <div className="quiz-admin-row mb-3">
                            <select
                                className="form-select form-select-sm quiz-admin-order-select"
                                value={buttonDraft.team_id}
                                disabled={busy}
                                onChange={(e) =>
                                    setButtonDraft((d) => ({ ...d, team_id: e.target.value }))
                                }
                            >
                                <option value="">Выберите команду…</option>
                                {teams.map((t) => (
                                    <option key={t.team_id} value={t.team_id}>
                                        {t.name}
                                    </option>
                                ))}
                            </select>
                            <input
                                type="text"
                                className="form-control form-control-sm quiz-admin-button-input"
                                placeholder="ID кнопки (например, 1)"
                                value={buttonDraft.button_id}
                                disabled={busy}
                                onChange={(e) =>
                                    setButtonDraft((d) => ({ ...d, button_id: e.target.value }))
                                }
                            />
                            <Button
                                size="sm"
                                variant="primary"
                                disabled={busy || !buttonDraft.team_id || !buttonDraft.button_id.trim()}
                                onClick={() =>
                                    run(
                                        () =>
                                            axiosQuizRegisterButton(
                                                eventId,
                                                buttonDraft.button_id.trim(),
                                                buttonDraft.team_id,
                                            ).then(() => setButtonDraft({ team_id: "", button_id: "" })),
                                        "Кнопка привязана",
                                    )
                                }
                            >
                                Привязать
                            </Button>
                        </div>

                        {buttons.length > 0 ? (
                            <div className="quiz-admin-row">
                                {buttons.map((b) => (
                                    <span key={b.button_id} className="quiz-button-chip">
                                        <strong>#{b.button_id}</strong> — {b.team_name}
                                        <Button
                                            size="sm"
                                            variant="outline-danger"
                                            disabled={busy}
                                            onClick={() =>
                                                run(
                                                    () => axiosQuizDeleteButton(eventId, b.button_id),
                                                    `Кнопка #${b.button_id} удалена`,
                                                )
                                            }
                                        >
                                            Удалить
                                        </Button>
                                    </span>
                                ))}
                            </div>
                        ) : (
                            <span className="text-body-secondary small">Привязанных кнопок пока нет.</span>
                        )}
                    </div>

                    {/* Question bank */}
                    <div className="quiz-admin-section liquid-glass-bg">
                        <div className="d-flex justify-content-between align-items-center flex-wrap gap-3">
                            <div>
                                <h4 className="mb-1">3. Банк вопросов</h4>
                                <span className="text-body-secondary small">
                                    Вопросы загружаются в базу заранее (миграцией). Загружено:{" "}
                                    <strong>{questions.length}</strong>.
                                </span>
                            </div>
                            <Button
                                variant="outline-secondary"
                                onClick={() => navigate(`/quiz/admin/${eventId}/questions`)}
                            >
                                Открыть банк вопросов →
                            </Button>
                        </div>
                    </div>
                </div>
            )}

            {/* Confirmation Modal */}
            <Modal show={!!confirmModal} onHide={() => setConfirmModal(null)} centered>
                <Modal.Header closeButton>
                    <Modal.Title>{confirmModal?.title}</Modal.Title>
                </Modal.Header>
                <Modal.Body>{confirmModal?.body}</Modal.Body>
                <Modal.Footer>
                    <Button variant="outline-secondary" onClick={() => setConfirmModal(null)}>
                        Отмена
                    </Button>
                    <Button
                        variant={confirmModal?.variant ?? "primary"}
                        disabled={busy}
                        onClick={() => {
                            const action = confirmModal?.onConfirm;
                            setConfirmModal(null);
                            action?.();
                        }}
                    >
                        {confirmModal?.confirmLabel}
                    </Button>
                </Modal.Footer>
            </Modal>
        </div>
    );
}