import { useEffect, useState } from "react";
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

const ORDER_OPTIONS = [
    { key: "red", value: 1, title: "1. Красный", hint: "0 ошибок / 5 очков" },
    { key: "yellow", value: 2, title: "2. Жёлтый", hint: "1 ошибка / 3 очка" },
    { key: "green", value: 3, title: "3. Зелёный", hint: "2 ошибки / 1 очко" },
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
    const [orderPending, setOrderPending] = useState<string | null>(null);

    useEffect(() => {
        if (eventId) setTeam(getStoredTeam(eventId));
    }, [eventId]);

    // reset the local grade highlight on a new question
    useEffect(() => {
        setLastGrade("");
    }, [state?.current_round, state?.current_question]);

    if (!eventId) return null;

    const myTeamState = state?.teams.find((t) => t.team_id === team?.team_id);

    const handleLogin = async () => {
        setError("");
        setBusy(true);
        try {
            const res = await axiosHeraldLogin(eventId, codeInput.trim());
            const stored = { ...res, code: codeInput.trim() };
            localStorage.setItem(`quiz_herald_${eventId}`, JSON.stringify(stored));
            setTeam(stored);
        } catch (e: any) {
            setError(e.message);
        } finally {
            setBusy(false);
        }
    };

    const handleBuzz = async () => {
        if (!team) return;
        setError("");
        setNotice("");
        setBusy(true);
        try {
            await axiosQuizBuzz(eventId, team.code);
            setNotice("Нажатие зафиксировано!");
        } catch (e: any) {
            setError(e.message);
        } finally {
            setBusy(false);
        }
    };

    const handleGrade = async (grade: string) => {
        if (!team) return;
        setError("");
        setNotice("");
        setBusy(true);
        try {
            const res = await axiosHeraldGrade(eventId, team.code, grade);
            setLastGrade(grade);
            if (res.eliminated) {
                setNotice("Команда завершила участие в раунде.");
            } else if (res.points_awarded > 0) {
                setNotice(`Начислено очков: ${res.points_awarded}`);
            } else {
                setNotice("Оценка сохранена.");
            }
        } catch (e: any) {
            setError(e.message);
        } finally {
            setBusy(false);
        }
    };

    const handleOrderConfirm = async () => {
        if (!team || !orderPending) return;
        setError("");
        setBusy(true);
        try {
            await axiosHeraldChooseOrder(eventId, team.code, orderPending);
            setOrderPending(null);
        } catch (e: any) {
            setError(e.message);
            setOrderPending(null);
        } finally {
            setBusy(false);
        }
    };

    // ---- login gate ----
    if (!team) {
        return (
            <div className="quiz-herald-root">
                <BackgroundParticles />
                <div className="quiz-herald-body">
                    <h2>Экран герольда</h2>
                    <input
                        className="quiz-code-input"
                        placeholder="Код"
                        value={codeInput}
                        onChange={(e) => setCodeInput(e.target.value)}
                        inputMode="numeric"
                    />
                    <button
                        className="quiz-buzz-btn"
                        disabled={busy || codeInput.trim().length === 0}
                        onClick={handleLogin}
                    >
                        Войти
                    </button>
                    {error && <div className="quiz-msg error">{error}</div>}
                </div>
            </div>
        );
    }

    const roundCfg = state?.rounds.find((r) => r.round === state.current_round);

    return (
        <div className="quiz-herald-root">
            <BackgroundParticles />
            <div className="quiz-herald-header">
                <div className="quiz-herald-team">
                    Команда №{team.number} — {team.name}
                </div>
                <div className="quiz-herald-meta">
                    {roundCfg
                        ? `Раунд ${roundCfg.round} — «${roundCfg.name}»`
                        : "Ожидание начала"}
                    {state && state.current_question > 0 && ` · Вопрос №${state.current_question}`}
                </div>
                {myTeamState && myTeamState.round3_order > 0 && (
                    <div
                        className={`quiz-herald-badge ${
                            ["", "red", "yellow", "green"][myTeamState.round3_order]
                        }`}
                    >
                        Орден: {Round3OrderNames[myTeamState.round3_order]}
                    </div>
                )}
            </div>

            <div className="quiz-herald-body">
                {!state && <div className="quiz-msg">Загрузка…</div>}

                {state && myTeamState?.round3_eliminated && (
                    <div className="quiz-overlay-eliminated">
                        <div>Команда завершила участие в третьем раунде.</div>
                        <div>Набрано очков: {myTeamState.round3_score}</div>
                    </div>
                )}

                {state && !state.round_active && (
                    <div className="quiz-msg">Ожидание начала раунда…</div>
                )}

                {state && state.round_active && state.current_question === 0 && (
                    <div className="quiz-msg">Ожидание первого вопроса…</div>
                )}

                {/* Round 3 order selection */}
                {state &&
                    state.current_round === 3 &&
                    state.current_question === 0 &&
                    myTeamState &&
                    myTeamState.round3_order === 0 && (
                        <>
                            <h3>Выберите орден</h3>
                            {ORDER_OPTIONS.map((o) => (
                                <button
                                    key={o.key}
                                    className={`quiz-order-card ${o.key}`}
                                    onClick={() => setOrderPending(o.key)}
                                >
                                    {o.title}
                                    <br />
                                    <small>{o.hint}</small>
                                </button>
                            ))}
                        </>
                    )}

                {/* Round 1: buzz + yes/no */}
                {state && state.round_active && state.current_round === 1 && state.current_question > 0 && (
                    <>
                        <button
                            className="quiz-buzz-btn"
                            disabled={busy}
                            onClick={handleBuzz}
                        >
                            НАЖАТЬ КНОПКУ
                        </button>
                        {(() => {
                            const entry = state.buzz_queue.find(
                                (b) => b.team_id === team.team_id,
                            );
                            const myTurn =
                                entry?.status === BuzzStatus.Pending && entry.is_active;
                            return (
                                <div className="quiz-herald-actions">
                                    <button
                                        className="quiz-grade-btn yes"
                                        disabled={busy || !myTurn}
                                        onClick={() => handleGrade("yes")}
                                    >
                                        Да (+2)
                                    </button>
                                    <button
                                        className="quiz-grade-btn no"
                                        disabled={busy || !myTurn}
                                        onClick={() => handleGrade("no")}
                                    >
                                        Нет (0)
                                    </button>
                                </div>
                            );
                        })()}
                    </>
                )}

                {/* Round 2: three-tier grading */}
                {state && state.round_active && state.current_round === 2 && state.current_question > 0 && (
                    <div className="quiz-herald-actions">
                        <button
                            className={`quiz-grade-btn correct${
                                lastGrade === "correct" ? " selected" : ""
                            }`}
                            disabled={busy}
                            onClick={() => handleGrade("correct")}
                        >
                            Верно (+4)
                        </button>
                        <button
                            className={`quiz-grade-btn partial${
                                lastGrade === "partial" ? " selected" : ""
                            }`}
                            disabled={busy}
                            onClick={() => handleGrade("partial")}
                        >
                            Частично (+2)
                        </button>
                        <button
                            className={`quiz-grade-btn wrong${
                                lastGrade === "wrong" ? " selected" : ""
                            }`}
                            disabled={busy}
                            onClick={() => handleGrade("wrong")}
                        >
                            Неверно (0)
                        </button>
                    </div>
                )}

                {/* Round 3: correct/incorrect with lives */}
                {state &&
                    state.round_active &&
                    state.current_round === 3 &&
                    state.current_question > 0 &&
                    myTeamState &&
                    myTeamState.round3_order > 0 && (
                        <>
                            <div className="quiz-lives">
                                {"♥".repeat(myTeamState.round3_lives_left)}
                                {"♡".repeat(
                                    Math.max(
                                        (myTeamState.round3_order === 1
                                            ? 0
                                            : myTeamState.round3_order === 2
                                              ? 1
                                              : 2) - myTeamState.round3_lives_left,
                                        0,
                                    ),
                                )}
                            </div>
                            <div className="quiz-herald-actions">
                                <button
                                    className="quiz-grade-btn correct"
                                    disabled={busy}
                                    onClick={() => handleGrade("correct")}
                                >
                                    Верно
                                </button>
                                <button
                                    className="quiz-grade-btn wrong"
                                    disabled={busy}
                                    onClick={() => handleGrade("wrong")}
                                >
                                    Неверно
                                </button>
                            </div>
                        </>
                    )}

                {error && <div className="quiz-msg error">{error}</div>}
                {notice && <div className="quiz-msg ok">{notice}</div>}
            </div>

            {/* order confirmation modal */}
            {orderPending && (
                <div className="quiz-overlay-eliminated">
                    <div>Вы уверены?</div>
                    <div className="text-body-secondary" style={{ marginTop: "1vh" }}>
                        Этот выбор нельзя будет изменить.
                    </div>
                    <div className="quiz-herald-actions" style={{ marginTop: "3vh" }}>
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
            )}
        </div>
    );
}
