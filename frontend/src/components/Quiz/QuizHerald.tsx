import { type FormEvent, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { useQuizState } from "../helpers/useQuizState";
import {
    axiosHeraldChooseOrder,
    axiosHeraldGrade,
    axiosHeraldLogin,
    axiosQuizBuzz,
} from "../../requests/quizRequests";
import { BuzzStatus, Round3OrderNames, type HeraldLoginResponse } from "../../types/quizTypes";
import BackgroundParticles from "../common/BackgroundParticles";
import "../../assets/Quiz.css";

interface OrderOption {
    key: "red" | "yellow" | "green";
    value: number;
    title: string;
    shortTitle: string;
    hint: string;
    points: number;
    maxMistakes: number;
}

const ORDER_OPTIONS: OrderOption[] = [
    {
        key: "red",
        value: 1,
        title: "1. Красный орден",
        shortTitle: "Красный орден",
        hint: "0 ошибок · 5 баллов за ответ",
        points: 5,
        maxMistakes: 0,
    },
    {
        key: "yellow",
        value: 2,
        title: "2. Жёлтый орден",
        shortTitle: "Жёлтый орден",
        hint: "1 ошибка · 3 балла за ответ",
        points: 3,
        maxMistakes: 1,
    },
    {
        key: "green",
        value: 3,
        title: "3. Зелёный орден",
        shortTitle: "Зелёный орден",
        hint: "2 ошибки · 1 балл за ответ",
        points: 1,
        maxMistakes: 2,
    },
];

function getStoredTeam(eventId: string): (HeraldLoginResponse & { code: string }) | null {
    try {
        const raw = localStorage.getItem(`quiz_herald_${eventId}`);
        return raw ? JSON.parse(raw) : null;
    } catch {
        return null;
    }
}

export default function QuizHerald() {
    const { eventId } = useParams<{ eventId: string }>();
    const { state } = useQuizState(eventId);

    const [team, setTeam] = useState<(HeraldLoginResponse & { code: string }) | null>(null);
    const [codeInput, setCodeInput] = useState("");
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");
    const [busy, setBusy] = useState(false);
    const [lastGrade, setLastGrade] = useState("");
    const [orderPending, setOrderPending] = useState<OrderOption | null>(null);
    const [logoutConfirm, setLogoutConfirm] = useState(false);

    useEffect(() => {
        if (eventId) setTeam(getStoredTeam(eventId));
    }, [eventId]);

    // Reset local status when moving to a new round or question
    useEffect(() => {
        setLastGrade("");
        setNotice("");
        setError("");
    }, [state?.current_round, state?.current_question, state?.tiebreak_active, state?.tiebreak_question]);

    // Keep phone screen awake during the tournament
    useEffect(() => {
        if (!team) return;
        let wakeLock: { release: () => Promise<void> } | null = null;
        const nav = navigator as Navigator & {
            wakeLock?: { request: (type: "screen") => Promise<{ release: () => Promise<void> }> };
        };

        const acquire = async () => {
            try {
                if (nav.wakeLock && document.visibilityState === "visible") {
                    wakeLock = await nav.wakeLock.request("screen");
                }
            } catch {
                /* wakeLock not supported or rejected */
            }
        };

        acquire();
        const onVisibility = () => {
            if (document.visibilityState === "visible") acquire();
        };
        document.addEventListener("visibilitychange", onVisibility);

        return () => {
            document.removeEventListener("visibilitychange", onVisibility);
            wakeLock?.release().catch(() => {});
        };
    }, [team]);

    if (!eventId) return null;

    const myTeamState = state?.teams.find((t) => t.team_id === team?.team_id);

    const handleLogin = async (e?: FormEvent) => {
        e?.preventDefault();
        const trimmed = codeInput.trim();
        if (!trimmed || busy) return;
        setError("");
        setBusy(true);
        try {
            const res = await axiosHeraldLogin(eventId, trimmed);
            const stored = { ...res, code: trimmed };
            localStorage.setItem(`quiz_herald_${eventId}`, JSON.stringify(stored));
            setTeam(stored);
            setCodeInput("");
        } catch (err: any) {
            setError(err?.message || "Неверный код команды");
        } finally {
            setBusy(false);
        }
    };

    const handleLogout = () => {
        localStorage.removeItem(`quiz_herald_${eventId}`);
        setTeam(null);
        setLogoutConfirm(false);
        setError("");
        setNotice("");
        setLastGrade("");
    };

    const handleBuzz = async () => {
        if (!team || busy) return;
        setError("");
        setNotice("");
        setBusy(true);
        try {
            await axiosQuizBuzz(eventId, team.code);
            setNotice("Нажатие зафиксировано!");
        } catch (err: any) {
            setError(err?.message || "Не удалось зафиксировать нажатие");
        } finally {
            setBusy(false);
        }
    };

    const handleGrade = async (grade: string) => {
        if (!team || busy) return;
        setError("");
        setNotice("");
        setBusy(true);
        try {
            const res = await axiosHeraldGrade(eventId, team.code, grade);
            setLastGrade(grade);
            if (res.eliminated) {
                setNotice("Команда завершила участие в туре.");
            } else if (res.points_awarded > 0) {
                setNotice(`Начислено баллов: +${res.points_awarded}`);
            } else {
                setNotice("Оценка сохранена (0 баллов).");
            }
        } catch (err: any) {
            setError(err?.message || "Ошибка отправки оценки");
        } finally {
            setBusy(false);
        }
    };

    const handleOrderConfirm = async () => {
        if (!team || !orderPending || busy) return;
        setError("");
        setBusy(true);
        try {
            await axiosHeraldChooseOrder(eventId, team.code, orderPending.key);
            setOrderPending(null);
        } catch (err: any) {
            setError(err?.message || "Не удалось сохранить выбор ордена");
            setOrderPending(null);
        } finally {
            setBusy(false);
        }
    };

    // ---- Login gate ----
    if (!team) {
        return (
            <div className="quiz-herald-root">
                <BackgroundParticles />
                <div className="quiz-herald-body quiz-herald-login-wrap">
                    <form className="quiz-login-card" onSubmit={handleLogin}>
                        <div className="quiz-login-subtitle">Математический турнир</div>
                        <h2 className="quiz-login-title">Экран герольда</h2>
                        <input
                            className="quiz-code-input"
                            placeholder="Введите код команды"
                            value={codeInput}
                            onChange={(e) => setCodeInput(e.target.value)}
                            inputMode="numeric"
                            autoComplete="one-time-code"
                            autoFocus
                        />
                        <button
                            type="submit"
                            className="quiz-buzz-btn quiz-login-submit"
                            disabled={busy || codeInput.trim().length === 0}
                        >
                            {busy ? "Вход…" : "Войти"}
                        </button>
                        {error && <div className="quiz-msg error">{error}</div>}
                    </form>
                </div>
            </div>
        );
    }

    const roundCfg = state?.rounds.find((r) => r.round === state.current_round);
    const currentOrderCfg = ORDER_OPTIONS.find((o) => o.value === myTeamState?.round3_order);

    // Helper for Round 1 & Tie-break buzz/grade state machine
    const renderBuzzControls = (isTiebreak: boolean) => {
        if (!state) return null;

        // The question is annulled — no buzzes or grades are accepted anymore
        if (state.question?.cancelled) {
            return (
                <div className="quiz-status-card muted">
                    <div className="quiz-status-title">Вопрос отменён</div>
                    <div className="quiz-status-desc">
                        Результаты вопроса аннулированы. Ожидайте следующий вопрос…
                    </div>
                </div>
            );
        }

        const queueIdx = state.buzz_queue.findIndex((b) => b.team_id === team.team_id);
        const entry = queueIdx >= 0 ? state.buzz_queue[queueIdx] : undefined;
        const myTurn = entry?.status === BuzzStatus.Pending && entry.is_active;
        const questionSolved =
            state.answer_revealed ||
            state.buzz_queue.some((b) => b.status === BuzzStatus.Correct) ||
            state.attempts_used >= state.attempts_max;

        // 1. Active turn: hide Buzz button, show full-size Yes / No grading buttons
        if (myTurn) {
            return (
                <div className="quiz-herald-stage">
                    <div className="quiz-status-banner active">
                        Отвечает ваша команда! Зафиксируйте результат:
                    </div>
                    <div className="quiz-herald-actions quiz-actions-split">
                        <button
                            className="quiz-grade-btn yes"
                            disabled={busy}
                            onClick={() => handleGrade("yes")}
                        >
                            <span className="quiz-btn-label">Да</span>
                            <span className="quiz-btn-sub">{isTiebreak ? "Верно" : "+2 балла"}</span>
                        </button>
                        <button
                            className="quiz-grade-btn no"
                            disabled={busy}
                            onClick={() => handleGrade("no")}
                        >
                            <span className="quiz-btn-label">Нет</span>
                            <span className="quiz-btn-sub">0 баллов</span>
                        </button>
                    </div>
                </div>
            );
        }

        // 2. Answered correctly
        if (entry?.status === BuzzStatus.Correct) {
            return (
                <div className="quiz-status-card success">
                    <div className="quiz-status-title">Верный ответ!</div>
                    <div className="quiz-status-desc">
                        {isTiebreak
                            ? "Ответ засчитан. Ожидайте решение ведущего."
                            : "Команде начислено +2 балла. Ожидайте следующий вопрос."}
                    </div>
                </div>
            );
        }

        // 3. Answered wrong — blocked until next question
        if (entry?.status === BuzzStatus.Wrong) {
            return (
                <div className="quiz-status-card blocked">
                    <div className="quiz-status-title">Попытка использована</div>
                    <div className="quiz-status-desc">
                        Кнопка заблокирована до конца текущего вопроса.
                    </div>
                </div>
            );
        }

        // 4. Another team holds the attempt — wait; outpaced teams may re-buzz after a wrong answer
        const activeTeam = state.buzz_queue.find((b) => b.is_active);
        if (activeTeam) {
            return (
                <div className="quiz-status-card waiting">
                    <div className="quiz-status-title">
                        Сейчас отвечает команда «{activeTeam.team_name}»
                    </div>
                    <div className="quiz-status-desc">
                        {entry?.status === BuzzStatus.Outpaced
                            ? "Если ответ не засчитают, вы сможете нажать кнопку снова."
                            : "Ожидайте своей очереди на ответ…"}
                    </div>
                </div>
            );
        }

        // 5. Question already closed
        if (questionSolved) {
            return (
                <div className="quiz-status-card muted">
                    <div className="quiz-status-title">Розыгрыш вопроса завершён</div>
                    <div className="quiz-status-desc">Ожидайте следующий вопрос…</div>
                </div>
            );
        }

        // 6. Ready to buzz (including re-buzz after being outpaced in an earlier window)
        return (
            <div className="quiz-herald-stage">
                <div className="quiz-attempts-pill">
                    Попыток использовано: {state.attempts_used} / {state.attempts_max}
                </div>
                <button
                    className="quiz-buzz-btn quiz-buzz-hero"
                    disabled={busy}
                    onClick={handleBuzz}
                >
                    НАЖАТЬ КНОПКУ
                </button>
            </div>
        );
    };

    // Single state-machine renderer (prevents overlapping screens)
    const renderBody = () => {
        if (!state) {
            return <div className="quiz-status-card muted">Подключение к серверу…</div>;
        }

        // 1. Tie-break mode
        if (state.tiebreak_active) {
            if (!myTeamState?.in_tiebreak) {
                return (
                    <div className="quiz-status-card muted">
                        <div className="quiz-status-title">Идёт переигровка</div>
                        <div className="quiz-status-desc">
                            В переигровке участвуют другие команды. Ожидайте…
                        </div>
                    </div>
                );
            }
            if (state.tiebreak_question === 0) {
                return (
                    <div className="quiz-status-card waiting">
                        <div className="quiz-status-title">Переигровка</div>
                        <div className="quiz-status-desc">Ожидание резервного вопроса…</div>
                    </div>
                );
            }
            return renderBuzzControls(true);
        }

        // 2. Round 3 eliminated (Game Over for this round)
        if (state.current_round === 3 && myTeamState?.round3_eliminated) {
            return (
                <div className="quiz-eliminated-card">
                    <div className="quiz-eliminated-icon">✕</div>
                    <div className="quiz-eliminated-title">
                        Команда завершила участие в туре
                    </div>
                    <div className="quiz-eliminated-score">
                        Набрано баллов в 3-м туре: <strong>{myTeamState.round3_score}</strong>
                    </div>
                    <div className="quiz-eliminated-note">
                        За оставшиеся вопросы тура начисляется 0 баллов.
                    </div>
                </div>
            );
        }

        // 3. Round 3 — Screen 3.1 (Order selection before Question 1)
        if (
            state.current_round === 3 &&
            state.current_question === 0 &&
            myTeamState &&
            myTeamState.round3_order === 0
        ) {
            return (
                <div className="quiz-order-screen">
                    <h3 className="quiz-order-heading">Выбор ордена</h3>
                    <p className="quiz-order-subheading">
                        Зафиксируйте выбор капитана. Изменить орден после подтверждения нельзя.
                    </p>
                    <div className="quiz-order-list">
                        {ORDER_OPTIONS.map((o) => (
                            <button
                                key={o.key}
                                className={`quiz-order-card ${o.key}`}
                                disabled={busy}
                                onClick={() => setOrderPending(o)}
                            >
                                <span className="quiz-order-title">{o.title}</span>
                                <span className="quiz-order-hint">{o.hint}</span>
                            </button>
                        ))}
                    </div>
                </div>
            );
        }

        // 4. Waiting for round start
        if (!state.round_active) {
            return (
                <div className="quiz-status-card muted">
                    <div className="quiz-status-title">Ожидание начала тура</div>
                    <div className="quiz-status-desc">
                        Интерфейс обновится автоматически, когда ведущий запустит тур.
                    </div>
                </div>
            );
        }

        // 5. Waiting for Question 1 (after order is chosen in Round 3, or in Round 1/2)
        if (state.current_question === 0) {
            return (
                <div className="quiz-status-card waiting">
                    <div className="quiz-status-title">Ожидание первого вопроса…</div>
                    {state.current_round === 3 && currentOrderCfg && (
                        <div className="quiz-status-desc">
                            Выбран {currentOrderCfg.shortTitle.toLowerCase()} ({currentOrderCfg.hint}).
                        </div>
                    )}
                </div>
            );
        }

        // 6. Round 1: Brain-ring
        if (state.current_round === 1) {
            return renderBuzzControls(false);
        }

        // 7. Round 2: What? Where? When? (3 grading buttons)
        if (state.current_round === 2) {
            if (state.question?.cancelled) {
                return (
                    <div className="quiz-status-card muted">
                        <div className="quiz-status-title">Вопрос отменён</div>
                        <div className="quiz-status-desc">
                            Результаты вопроса аннулированы. Ожидайте следующий вопрос…
                        </div>
                    </div>
                );
            }
            // grading opens only after the question time is over (or the host closes it)
            if (state.timer.remaining_ms > 0) {
                return (
                    <div className="quiz-status-card waiting">
                        <div className="quiz-status-title">Идёт время на ответ</div>
                        <div className="quiz-status-desc">
                            Оценка бланка откроется после окончания времени.
                        </div>
                    </div>
                );
            }
            return (
                <div className="quiz-herald-stage">
                    <div className="quiz-stage-label">Оценка бланка команды</div>
                    <div className="quiz-herald-actions quiz-actions-stack">
                        <button
                            className={`quiz-grade-btn correct${lastGrade === "correct" ? " selected" : ""}`}
                            disabled={busy}
                            onClick={() => handleGrade("correct")}
                        >
                            <span className="quiz-btn-label">Правильно</span>
                            <span className="quiz-btn-sub">+4 балла</span>
                        </button>
                        <button
                            className={`quiz-grade-btn partial${lastGrade === "partial" ? " selected" : ""}`}
                            disabled={busy}
                            onClick={() => handleGrade("partial")}
                        >
                            <span className="quiz-btn-label">Частично</span>
                            <span className="quiz-btn-sub">+2 балла</span>
                        </button>
                        <button
                            className={`quiz-grade-btn wrong${lastGrade === "wrong" ? " selected" : ""}`}
                            disabled={busy}
                            onClick={() => handleGrade("wrong")}
                        >
                            <span className="quiz-btn-label">Неправильно</span>
                            <span className="quiz-btn-sub">0 баллов</span>
                        </button>
                    </div>
                </div>
            );
        }

        // 8. Round 3 — Screen 3.2 (Grading with lives)
        if (state.current_round === 3 && myTeamState && currentOrderCfg) {
            if (state.question?.cancelled) {
                return (
                    <div className="quiz-status-card muted">
                        <div className="quiz-status-title">Вопрос отменён</div>
                        <div className="quiz-status-desc">
                            Результаты вопроса аннулированы. Ожидайте следующий вопрос…
                        </div>
                    </div>
                );
            }
            const maxMistakes = currentOrderCfg.maxMistakes;
            const livesLeft = myTeamState.round3_lives_left;

            return (
                <div className="quiz-herald-stage">
                    <div className="quiz-lives-box">
                        {maxMistakes === 0 ? (
                            <div className="quiz-lives-warning">
                                Без права на ошибку (1-я ошибка — выбывание)
                            </div>
                        ) : (
                            <>
                                <div className="quiz-lives-label">
                                    Право на ошибку: {livesLeft} из {maxMistakes}
                                </div>
                                <div className="quiz-lives" aria-label={`Осталось жизней: ${livesLeft}`}>
                                    {"♥".repeat(Math.max(livesLeft, 0))}
                                    {"♡".repeat(Math.max(maxMistakes - livesLeft, 0))}
                                </div>
                            </>
                        )}
                    </div>

                    <div className="quiz-herald-actions quiz-actions-split">
                        <button
                            className={`quiz-grade-btn correct${lastGrade === "correct" ? " selected" : ""}`}
                            disabled={busy}
                            onClick={() => handleGrade("correct")}
                        >
                            <span className="quiz-btn-label">Верно</span>
                            <span className="quiz-btn-sub">+{currentOrderCfg.points} б.</span>
                        </button>
                        <button
                            className={`quiz-grade-btn wrong${lastGrade === "wrong" ? " selected" : ""}`}
                            disabled={busy}
                            onClick={() => handleGrade("wrong")}
                        >
                            <span className="quiz-btn-label">Неверно</span>
                            <span className="quiz-btn-sub">
                                {livesLeft === 0 ? "Выбывание" : "−1 право"}
                            </span>
                        </button>
                    </div>
                </div>
            );
        }

        return <div className="quiz-status-card muted">Ожидание действий ведущего…</div>;
    };

    return (
        <div className="quiz-herald-root">
            <BackgroundParticles />

            <header className="quiz-herald-header">
                <div className="quiz-herald-top-row">
                    <div className="quiz-herald-team">
                        <span className="quiz-team-num">№{team.number}</span>
                        <span className="quiz-team-name">{team.name}</span>
                    </div>
                    <div className="quiz-herald-top-right">
                        {myTeamState && (
                            <div className="quiz-score-pill" title="Текущая сумма баллов">
                                {myTeamState.total_score} б.
                            </div>
                        )}
                        <button
                            type="button"
                            className="quiz-logout-btn"
                            onClick={() => setLogoutConfirm(true)}
                            title="Сменить команду"
                            aria-label="Сменить команду"
                        >
                            Выйти
                        </button>
                    </div>
                </div>

                <div className="quiz-herald-meta-row">
                    <div className="quiz-herald-meta">
                        {state?.tiebreak_active
                            ? "Переигровка"
                            : roundCfg
                              ? `Тур ${roundCfg.round} · «${roundCfg.name}»`
                              : "Ожидание начала"}
                        {state &&
                            !state.tiebreak_active &&
                            state.current_question > 0 &&
                            ` · Вопрос №${state.current_question}`}
                        {state &&
                            state.tiebreak_active &&
                            state.tiebreak_question > 0 &&
                            ` · Резервный №${state.tiebreak_question}`}
                    </div>

                    {myTeamState && myTeamState.round3_order > 0 && (
                        <div
                            className={`quiz-herald-badge ${
                                ["", "red", "yellow", "green"][myTeamState.round3_order]
                            }`}
                        >
                            {Round3OrderNames[myTeamState.round3_order]}
                        </div>
                    )}
                </div>
            </header>

            <main className="quiz-herald-body">
                {renderBody()}
                {error && <div className="quiz-msg error">{error}</div>}
                {notice && <div className="quiz-msg ok">{notice}</div>}
            </main>

            {/* Order confirmation modal */}
            {orderPending && (
                <div className="quiz-modal-backdrop" role="dialog" aria-modal="true">
                    <div className="quiz-modal-card">
                        <div className="quiz-modal-title">Подтвердите выбор</div>
                        <div className={`quiz-modal-order-pill ${orderPending.key}`}>
                            {orderPending.shortTitle}
                        </div>
                        <div className="quiz-modal-desc">
                            {orderPending.hint}.
                            <br />
                            <strong>Изменить выбор после подтверждения будет нельзя.</strong>
                        </div>
                        <div className="quiz-herald-actions quiz-actions-split">
                            <button
                                className="quiz-grade-btn yes"
                                disabled={busy}
                                onClick={handleOrderConfirm}
                            >
                                Подтвердить
                            </button>
                            <button
                                className="quiz-grade-btn no"
                                disabled={busy}
                                onClick={() => setOrderPending(null)}
                            >
                                Отмена
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Logout confirmation modal */}
            {logoutConfirm && (
                <div className="quiz-modal-backdrop" role="dialog" aria-modal="true">
                    <div className="quiz-modal-card">
                        <div className="quiz-modal-title">Сменить команду?</div>
                        <div className="quiz-modal-desc">
                            Вы выйдете из пульта команды №{team.number} — «{team.name}».
                        </div>
                        <div className="quiz-herald-actions quiz-actions-split">
                            <button className="quiz-grade-btn no" onClick={handleLogout}>
                                Выйти
                            </button>
                            <button
                                className="quiz-grade-btn neutral"
                                onClick={() => setLogoutConfirm(false)}
                            >
                                Отмена
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
