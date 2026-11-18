import { useCallback, useEffect, useRef, useState } from "react";
import { Button, Collapse, Table } from "react-bootstrap";
import { useNavigate, useParams } from "react-router-dom";
import { useQuizState } from "../helpers/useQuizState";
import {
    axiosQuizAdjustScore,
    axiosQuizCloseQuestion,
    axiosQuizDeleteButton,
    axiosQuizGetQuestions,
    axiosQuizGetTeams,
    axiosQuizGrantTurn,
    axiosQuizListButtons,
    axiosQuizPauseTimer,
    axiosQuizRegisterButton,
    axiosQuizResetQuestion,
    axiosQuizResumeTimer,
    axiosQuizRevealAnswer,
    axiosQuizSetOrder,
    axiosQuizSetupTeams,
    axiosQuizShowLeaderboard,
    axiosQuizStartQuestion,
    axiosQuizStartRound,
} from "../../requests/quizRequests";
import { BuzzStatus, type QuizButtonAdmin, type QuizQuestion, type QuizTeamAdmin } from "../../types/quizTypes";
import "../../assets/Quiz.css";

const ORDER_OPTIONS: [number, string][] = [
    [0, "Не выбран"],
    [1, "Красный"],
    [2, "Жёлтый"],
    [3, "Зелёный"],
];

function formatTime(totalSeconds: number): string {
    const m = Math.floor(totalSeconds / 60);
    const s = totalSeconds % 60;
    return `${m}:${s.toString().padStart(2, "0")}`;
}

function teamTotal(t: QuizTeamAdmin): number {
    return t.round1_score + t.round2_score + t.round3_score;
}

/** Editable round score cell: shows the effective value, OK applies the delta. */
function AdjustCell({
    value,
    adjust,
    onApply,
}: {
    value: number;
    adjust: number;
    onApply: (delta: number) => void;
}) {
    const [draft, setDraft] = useState(String(value));

    useEffect(() => setDraft(String(value)), [value]);

    const parsed = Number(draft);
    const valid = draft.trim() !== "" && !Number.isNaN(parsed);
    const changed = valid && parsed !== value;

    return (
        <div>
            <div className="quiz-admin-adjust">
                <input
                    type="number"
                    className="form-control form-control-sm quiz-admin-adjust-input"
                    value={draft}
                    onChange={(e) => setDraft(e.target.value)}
                />
                <Button
                    size="sm"
                    variant="outline-primary"
                    disabled={!changed}
                    onClick={() => onApply(parsed - value)}
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

/** Round-3 order selector; applies immediately on change (admin override). */
function OrderCell({ value, onChange }: { value: number; onChange: (order: string) => void }) {
    return (
        <select
            className="form-select form-select-sm quiz-admin-order-select"
            value={value}
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

    const [teams, setTeams] = useState<QuizTeamAdmin[]>([]);
    const [questions, setQuestions] = useState<QuizQuestion[]>([]);
    const [buttons, setButtons] = useState<QuizButtonAdmin[]>([]);
    const [buttonDraft, setButtonDraft] = useState({ team_id: "", button_id: "" });
    const [showTeams, setShowTeams] = useState(false);
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");

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

    // track a local deadline so the timer ticks smoothly between broadcasts;
    // re-sync only on real discontinuities: start/resume or a new duration
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
            if (!runningRef.current) return; // stay frozen while paused
            setRemainingMs(Math.max(deadlineRef.current - Date.now(), 0));
        }, 250);
        return () => clearInterval(tick);
    }, []);

    const run = async (fn: () => Promise<void>, successMsg: string) => {
        setError("");
        setNotice("");
        try {
            await fn();
            setNotice(successMsg);
            loadTeams();
            loadButtons();
        } catch (e: any) {
            setError(e.message);
        }
    };

    if (!eventId) return null;

    const roundCfg = state?.rounds.find((r) => r.round === state.current_round);
    const nextQuestionNumber = (state?.current_question ?? 0) + 1;
    const nextQuestionExists = roundCfg ? nextQuestionNumber <= roundCfg.question_count : false;
    const currentQuestionData =
        state && roundCfg
            ? questions.find(
                  (q) => q.round === state.current_round && q.question_number === state.current_question,
              )
            : undefined;

    const timerRunning = !!state?.timer.running;
    const timerState = !roundCfg ? "idle" : !timerRunning ? "paused" : remainingMs <= 30_000 ? "warning" : "running";
    const timerLabel =
        timerState === "running"
            ? "Время идёт"
            : timerState === "warning"
              ? "Осталось мало времени"
              : timerState === "paused"
                ? "Пауза"
                : "Таймер не запущен";
    const teamsTotal = teams.reduce((sum, t) => sum + teamTotal(t), 0);

    return (
        <div className="quiz-admin-root">
            <div className="quiz-admin-head">
                <h2 className="mb-0">Панель ведущего — квиз-шоу</h2>
                <span className="quiz-admin-status">
                    {roundCfg
                        ? `Раунд ${roundCfg.round} — «${roundCfg.name}» · Вопрос ${state?.current_question ?? 0} из ${roundCfg.question_count}`
                        : "Раунд не начат"}
                    {state?.answer_revealed ? " · ответ показан" : ""}
                </span>
            </div>

            {error && <div className="quiz-msg error mb-3">{error}</div>}
            {notice && <div className="quiz-msg ok mb-3">{notice}</div>}

            {/* Timer */}
            <div className={`quiz-admin-timer ${timerState}`}>
                <span className="quiz-admin-timer-value">{formatTime(Math.ceil(remainingMs / 1000))}</span>
                <span className="quiz-admin-timer-state">{timerLabel}</span>
            </div>

            {/* Game flow */}
            <div className="quiz-admin-section liquid-glass-bg">
                <h4>Управление игрой</h4>
                <div className="quiz-admin-row mb-3">
                    {[1, 2, 3].map((r) => (
                        <Button
                            key={r}
                            variant={state?.current_round === r ? "success" : "outline-secondary"}
                            onClick={() =>
                                run(() => axiosQuizStartRound(eventId, r), `Раунд ${r} начат`)
                            }
                        >
                            Начать раунд {r}
                        </Button>
                    ))}
                    <Button
                        variant={state?.leaderboard_shown ? "success" : "outline-secondary"}
                        onClick={() =>
                            run(
                                () => axiosQuizShowLeaderboard(eventId, !state?.leaderboard_shown),
                                "Таблица лидеров переключена",
                            )
                        }
                    >
                        {state?.leaderboard_shown ? "Скрыть таблицу" : "Показать таблицу"}
                    </Button>
                </div>

                <div className="quiz-admin-row mb-3">
                    <Button
                        variant="primary"
                        disabled={!roundCfg || !nextQuestionExists}
                        onClick={() => run(() => axiosQuizStartQuestion(eventId), `Вопрос ${nextQuestionNumber} запущен`)}
                    >
                        {`Запустить вопрос ${nextQuestionNumber}`}
                    </Button>
                    <Button
                        variant="outline-secondary"
                        onClick={() => run(() => axiosQuizPauseTimer(eventId), "Таймер приостановлен")}
                    >
                        Пауза
                    </Button>
                    <Button
                        variant="outline-secondary"
                        onClick={() => run(() => axiosQuizResumeTimer(eventId), "Таймер продолжен")}
                    >
                        Продолжить
                    </Button>
                    <Button
                        variant="outline-secondary"
                        onClick={() => run(() => axiosQuizCloseQuestion(eventId), "Вопрос закрыт")}
                    >
                        Закрыть вопрос
                    </Button>
                    <Button
                        variant="success"
                        onClick={() => run(() => axiosQuizRevealAnswer(eventId), "Ответ показан")}
                    >
                        Показать ответ
                    </Button>
                    <Button
                        variant="warning"
                        onClick={() => run(() => axiosQuizResetQuestion(eventId), "Очередь сброшена")}
                    >
                        Сбросить очередь
                    </Button>
                </div>

                {state?.current_round === 1 && state.current_question > 0 && (
                    <div className="mb-2">
                        <div className="text-body-secondary mb-2">
                            Окно нажатий (попыток использовано: {state.attempts_used} /{" "}
                            {state.attempts_max})
                        </div>
                        {state.buzz_queue.length === 0 && (
                            <span className="text-body-secondary">Нажатий пока не было.</span>
                        )}
                        <div className="quiz-admin-row">
                            {state.buzz_queue.map((b, i) => (
                                <div
                                    key={b.team_id}
                                    className={`quiz-queue-chip${
                                        b.is_active ? " active" : ""
                                    }${b.status === BuzzStatus.Wrong ? " wrong" : ""}${
                                        b.status === BuzzStatus.Correct ? " correct" : ""
                                    }${b.status === BuzzStatus.Outpaced ? " outpaced" : ""}`}
                                >
                                    {b.team_name}
                                    {i > 0 && ` +${b.offset_ms} мс`}
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {state && roundCfg && state.current_question > 0 && (
                    <div className="mb-1" style={{ textAlign: "left" }}>
                        <div className="text-body-secondary mb-2">
                            Текущий вопрос для ведущего:
                        </div>
                        {currentQuestionData ? (
                            <div className="liquid-glass-bg rounded-3 p-3">
                                <div>{currentQuestionData.text}</div>
                                <div className="text-success mt-2">
                                    Ответ: {currentQuestionData.answer}
                                </div>
                            </div>
                        ) : (
                            <div className="quiz-msg error">
                                Для текущего вопроса нет данных в банке — проверьте загрузку вопросов.
                            </div>
                        )}
                    </div>
                )}
            </div>

            {/* Teams & access codes — collapsible */}
            <div className="quiz-admin-section liquid-glass-bg">
                <div className="quiz-admin-row quiz-admin-teams-head">
                    <h4 className="mb-0">Команды и коды доступа</h4>
                    <Button
                        size="sm"
                        variant="outline-secondary"
                        onClick={() => setShowTeams(!showTeams)}
                        aria-expanded={showTeams}
                    >
                        {showTeams ? "Скрыть таблицу" : "Показать таблицу"}
                    </Button>
                    {!showTeams && teams.length > 0 && (
                        <span className="text-body-secondary">
                            Команд: {teams.length} · Сумма очков: {teamsTotal}
                        </span>
                    )}
                </div>
                <div className="quiz-admin-row mb-3">
                    <Button onClick={() => run(() => axiosQuizSetupTeams(eventId), "Команды подготовлены")}>
                        Подготовить / добавить команды
                    </Button>
                    <span className="text-body-secondary">
                        Коды выдаются герольдам для входа на экраны оценки. Очки по раундам можно
                        исправить вручную — коррекции не стираются при пересчётах.
                    </span>
                </div>
                <Collapse in={showTeams}>
                    <div>
                        {teams.length > 0 ? (
                            <div className="table-responsive">
                                <Table striped bordered hover size="sm" className="quiz-admin-table">
                                    <thead>
                                        <tr>
                                            <th>Команда</th>
                                            <th>Код герольда</th>
                                            <th>Раунд 1</th>
                                            <th>Раунд 2</th>
                                            <th>Раунд 3</th>
                                            <th>Итог</th>
                                            <th>Орден</th>
                                            <th></th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {teams.map((t) => (
                                            <tr key={t.id}>
                                                <td>
                                                    {t.name}
                                                    {t.round3_eliminated && (
                                                        <span className="quiz-admin-eliminated"> (выбыла)</span>
                                                    )}
                                                </td>
                                                <td>
                                                    <strong>{t.access_code}</strong>
                                                </td>
                                                <td>
                                                    <AdjustCell
                                                        value={t.round1_score}
                                                        adjust={t.round1_adjust}
                                                        onApply={(delta) =>
                                                            run(
                                                                () =>
                                                                    axiosQuizAdjustScore(
                                                                        eventId,
                                                                        t.team_id,
                                                                        1,
                                                                        delta,
                                                                    ),
                                                                `Раунд 1, ${t.name}: ${delta > 0 ? "+" : ""}${delta}`,
                                                            )
                                                        }
                                                    />
                                                </td>
                                                <td>
                                                    <AdjustCell
                                                        value={t.round2_score}
                                                        adjust={t.round2_adjust}
                                                        onApply={(delta) =>
                                                            run(
                                                                () =>
                                                                    axiosQuizAdjustScore(
                                                                        eventId,
                                                                        t.team_id,
                                                                        2,
                                                                        delta,
                                                                    ),
                                                                `Раунд 2, ${t.name}: ${delta > 0 ? "+" : ""}${delta}`,
                                                            )
                                                        }
                                                    />
                                                </td>
                                                <td>
                                                    <AdjustCell
                                                        value={t.round3_score}
                                                        adjust={t.round3_adjust}
                                                        onApply={(delta) =>
                                                            run(
                                                                () =>
                                                                    axiosQuizAdjustScore(
                                                                        eventId,
                                                                        t.team_id,
                                                                        3,
                                                                        delta,
                                                                    ),
                                                                `Раунд 3, ${t.name}: ${delta > 0 ? "+" : ""}${delta}`,
                                                            )
                                                        }
                                                    />
                                                </td>
                                                <td>
                                                    <strong>{teamTotal(t)}</strong>
                                                </td>
                                                <td>
                                                    <OrderCell
                                                        value={t.round3_order}
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
                                                <td>
                                                    <Button
                                                        size="sm"
                                                        variant="warning"
                                                        onClick={() =>
                                                            run(
                                                                () => axiosQuizGrantTurn(eventId, t.team_id),
                                                                `Право ответа передано: ${t.name}`,
                                                            )
                                                        }
                                                    >
                                                        Дать право ответа
                                                    </Button>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </Table>
                            </div>
                        ) : (
                            <span className="text-body-secondary">Команды ещё не подготовлены.</span>
                        )}
                    </div>
                </Collapse>

                {/* ESP32 buttons */}
                <div className="mt-3">
                    <div className="text-body-secondary mb-2">
                        Кнопки команд (ESP32): id кнопки прошивается в устройство один раз вместе с id
                        события, привязка к команде — здесь.
                    </div>
                    <div className="quiz-admin-row mb-2">
                        <select
                            className="form-select form-select-sm quiz-admin-order-select"
                            value={buttonDraft.team_id}
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
                            placeholder="ID кнопки"
                            value={buttonDraft.button_id}
                            onChange={(e) =>
                                setButtonDraft((d) => ({ ...d, button_id: e.target.value }))
                            }
                        />
                        <Button
                            size="sm"
                            variant="primary"
                            disabled={!buttonDraft.team_id || !buttonDraft.button_id.trim()}
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
                    {buttons.length > 0 && (
                        <div className="quiz-admin-row">
                            {buttons.map((b) => (
                                <span key={b.button_id} className="quiz-button-chip">
                                    <strong>{b.button_id}</strong> — {b.team_name}
                                    <Button
                                        size="sm"
                                        variant="outline-danger"
                                        onClick={() =>
                                            run(
                                                () => axiosQuizDeleteButton(eventId, b.button_id),
                                                "Кнопка удалена",
                                            )
                                        }
                                    >
                                        Удалить
                                    </Button>
                                </span>
                            ))}
                        </div>
                    )}
                </div>
            </div>

            {/* Question bank (read-only: questions are preloaded via migrations) */}
            <div className="quiz-admin-section liquid-glass-bg">
                <div className="quiz-admin-row">
                    <h4 className="mb-0">Банк вопросов</h4>
                    <Button
                        variant="outline-secondary"
                        onClick={() => navigate(`/quiz/admin/${eventId}/questions`)}
                    >
                        Открыть банк вопросов
                    </Button>
                    <span className="text-body-secondary">
                        Вопросы загружаются в базу заранее (миграцией). Загружено: {questions.length}.
                    </span>
                </div>
            </div>
        </div>
    );
}
