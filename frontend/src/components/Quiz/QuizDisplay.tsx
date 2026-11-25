import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { useQuizState } from "../helpers/useQuizState";
import { BuzzStatus, type QuizTeamState } from "../../types/quizTypes";
import BackgroundParticles from "../common/BackgroundParticles";
import "../../assets/Quiz.css";

function formatTime(totalSeconds: number): string {
    const m = Math.floor(totalSeconds / 60);
    const s = totalSeconds % 60;
    return `${m}:${s.toString().padStart(2, "0")}`;
}

interface BoardRow {
    team: QuizTeamState;
    value: number;
}

function BoardColumn({ title, rows }: { title: string; rows: BoardRow[] }) {
    // Each column computes its own scale so single-round bars aren't dwarfed by cumulative totals
    const max = Math.max(...rows.map((r) => r.value), 1);

    return (
        <div className="quiz-board-column">
            <div className="quiz-board-title">{title}</div>
            <div className="quiz-board-rows">
                {rows.map(({ team, value }, index) => {
                    const rank = index + 1;
                    const pct = value > 0 ? Math.max((value / max) * 100, 10) : 0;
                    return (
                        <div
                            className={`quiz-bar-row${rank <= 3 ? ` rank-${rank}` : ""}`}
                            key={team.team_id}
                        >
                            <div className="quiz-bar-rank">{rank}</div>
                            <div className="quiz-bar-name" title={team.name}>
                                <span className="quiz-bar-num">№{team.number}</span> {team.name}
                            </div>
                            <div className="quiz-bar-track">
                                <div
                                    className={`quiz-bar-fill${value === 0 ? " zero" : ""}`}
                                    style={{ width: value > 0 ? `${pct}%` : "3.2rem" }}
                                >
                                    {value}
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

function Leaderboard({ round, teams }: { round: number; teams: QuizTeamState[] }) {
    const sortedBy = (fn: (t: QuizTeamState) => number): BoardRow[] =>
        teams
            .map((t) => ({ team: t, value: fn(t) }))
            .sort(
                (a, b) =>
                    b.value - a.value ||
                    b.team.tiebreak_score - a.team.tiebreak_score ||
                    a.team.number - b.team.number,
            );

    if (round === 1) {
        const rows = sortedBy((t) => t.round1_score);
        return (
            <div className="quiz-board single">
                <BoardColumn title="Итоги 1-го тура" rows={rows} />
            </div>
        );
    }

    if (round === 2) {
        const roundRows = sortedBy((t) => t.round2_score);
        const totalRows = sortedBy((t) => t.round1_score + t.round2_score);
        return (
            <div className="quiz-board double">
                <BoardColumn title="Результаты 2-го тура" rows={roundRows} />
                <BoardColumn title="Общий зачёт (1 + 2 туры)" rows={totalRows} />
            </div>
        );
    }

    const roundRows = sortedBy((t) => t.round3_score);
    const totalRows = sortedBy((t) => t.total_score);
    return (
        <div className="quiz-board double">
            <BoardColumn title="Результаты 3-го тура" rows={roundRows} />
            <BoardColumn title="Итоговый рейтинг турнира" rows={totalRows} />
        </div>
    );
}

export default function QuizDisplay() {
    const { eventId } = useParams<{ eventId: string }>();
    const { state } = useQuizState(eventId);

    const [remainingMs, setRemainingMs] = useState<number>(0);
    const [soundUnlocked, setSoundUnlocked] = useState<boolean>(false);

    const deadlineRef = useRef<number>(0);
    const runningRef = useRef<boolean>(false);
    const beepedRef = useRef<string>("");
    const audioRef = useRef<HTMLAudioElement | null>(null);

    const unlockAudio = () => {
        try {
            if (!audioRef.current) {
                audioRef.current = new Audio("/start.mp3");
                audioRef.current.volume = 0.6;
            }
            const audio = audioRef.current;
            audio.muted = true;
            audio
                .play()
                .then(() => {
                    audio.pause();
                    audio.currentTime = 0;
                    audio.muted = false;
                    setSoundUnlocked(true);
                })
                .catch(() => {
                    /* browser still blocked */
                });
        } catch {
            /* audio not available */
        }
    };

    const playTimeoutSound = () => {
        try {
            const audio = audioRef.current ?? new Audio("/start.mp3");
            audio.volume = 0.6;
            audio.currentTime = 0;
            audio.play().catch(() => {
                /* autoplay blocked or audio unavailable */
            });
        } catch {
            /* audio unavailable */
        }
    };

    // Track local deadline so the timer ticks smoothly between broadcasts
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
        }, 150);
        return () => clearInterval(tick);
    }, []);

    // Audio cue when the timer hits zero
    useEffect(() => {
        if (!state || !state.timer.running) return;
        const activeQuestion = state.tiebreak_active
            ? state.tiebreak_question
            : state.current_question;
        if (remainingMs > 0 || activeQuestion === 0) return;
        const key = `${state.tiebreak_active ? "tb" : state.current_round}:${activeQuestion}`;
        if (beepedRef.current === key) return;
        beepedRef.current = key;
        playTimeoutSound();
    }, [remainingMs, state]);

    if (!state) {
        return (
            <div className="quiz-display-root">
                <BackgroundParticles />
                <div className="quiz-display-idle">Загрузка табло…</div>
            </div>
        );
    }

    const roundCfg = state.rounds.find((r) => r.round === state.current_round);
    const secondsLeft = Math.ceil(remainingMs / 1000);
    const totalDurationSec = state.current_round === 3 && !state.tiebreak_active ? 120 : 60;
    const progressPct = Math.min(100, Math.max(0, (remainingMs / (totalDurationSec * 1000)) * 100));

    const showQuestionScreen = !state.leaderboard_shown && Boolean(state.question);
    const isPaused =
        showQuestionScreen && !state.timer.running && remainingMs > 0 && !state.answer_revealed;
    const isExpired = showQuestionScreen && remainingMs === 0;
    const isWarning = showQuestionScreen && secondsLeft <= 10 && secondsLeft > 0;
    const isCaution = showQuestionScreen && secondsLeft <= 20 && secondsLeft > 10;

    const timerClass = [
        "quiz-display-timer",
        isExpired ? "expired" : "",
        isWarning ? "warning" : "",
        isCaution ? "caution" : "",
        isPaused ? "paused" : "",
    ]
        .filter(Boolean)
        .join(" ");

    const showBuzzFooter =
        showQuestionScreen && (state.current_round === 1 || state.tiebreak_active);

    return (
        <div className="quiz-display-root">
            <BackgroundParticles />

            <header className="quiz-display-header">
                <div className="quiz-display-round">
                    {state.tiebreak_active
                        ? "Переигровка"
                        : roundCfg
                          ? `Тур ${roundCfg.round} · «${roundCfg.name}»`
                          : "Математический турнир"}
                </div>

                <div className="quiz-display-qnum">
                    {state.question && !state.leaderboard_shown
                        ? state.tiebreak_active
                            ? `Резервный вопрос №${state.question.number}`
                            : `Вопрос №${state.question.number}`
                        : state.leaderboard_shown
                          ? "Турнирная таблица"
                          : ""}
                </div>

                <div className="quiz-display-header-right">
                    {showQuestionScreen && (
                        <div className={timerClass}>
                            {isPaused && <span className="quiz-timer-badge">ПАУЗА</span>}
                            {isExpired && <span className="quiz-timer-badge expired">ВРЕМЯ</span>}
                            <span className="quiz-timer-digits">{formatTime(secondsLeft)}</span>
                        </div>
                    )}
                    {!soundUnlocked && (
                        <button
                            type="button"
                            className="quiz-sound-unlock"
                            onClick={unlockAudio}
                            title="Нажмите один раз перед турниром, чтобы разрешить звуковой сигнал таймера"
                        >
                            🔇 Вкл. звук
                        </button>
                    )}
                </div>
            </header>

            {showQuestionScreen && (
                <div className="quiz-timer-progress-track">
                    <div
                        className={`quiz-timer-progress-fill${
                            isExpired || isWarning ? " danger" : isCaution || isPaused ? " caution" : ""
                        }`}
                        style={{ width: `${progressPct}%` }}
                    />
                </div>
            )}

            <main className="quiz-display-body">
                {state.leaderboard_shown ? (
                    <Leaderboard round={state.current_round} teams={state.teams} />
                ) : !state.question ? (
                    <div className="quiz-display-idle">
                        <div className="quiz-idle-title">
                            {state.round_active ? (
                                "Ожидание вопроса…"
                            ) : (
                                <>
                                    Добро пожаловать на
                                    <br />
                                    Третий региональный математический турнир школьных команд
                                    <br />
                                    «Королевская битва»!
                                </>
                            )}
                        </div>
                        {roundCfg && (
                            <div className="quiz-idle-subtitle">
                                Тур {roundCfg.round}: «{roundCfg.name}»
                            </div>
                        )}
                    </div>
                ) : (
                    <div
                        className={`quiz-stage-card${
                            state.answer_revealed ? " with-answer" : ""
                        }`}
                    >
                        <div className="quiz-display-question">{state.question.text}</div>

                        {state.answer_revealed && state.question.answer && (
                            <div className="quiz-display-answer">
                                <div className="quiz-display-answer-label">Правильный ответ</div>
                                <div className="quiz-display-answer-text">
                                    {state.question.answer}
                                </div>
                                {state.question.explanation && (
                                    <div className="quiz-display-explanation">
                                        {state.question.explanation}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                )}
            </main>

            {showBuzzFooter && (
                <footer className="quiz-display-footer">
                    <div className="quiz-display-queue">
                        {state.buzz_queue.length === 0 ? (
                            <div className="quiz-queue-empty">Ожидание нажатия кнопки…</div>
                        ) : (
                            state.buzz_queue.map((b, i) => (
                                <div
                                    key={b.team_id}
                                    className={`quiz-queue-chip${
                                        b.is_active ? " active" : ""
                                    }${b.status === BuzzStatus.Wrong ? " wrong" : ""}${
                                        b.status === BuzzStatus.Correct ? " correct" : ""
                                    }${b.status === BuzzStatus.Outpaced ? " outpaced" : ""}`}
                                >
                                    <span className="quiz-queue-pos">#{i + 1}</span>
                                    <span className="quiz-queue-name">{b.team_name}</span>
                                    {i > 0 && (
                                        <span className="quiz-queue-offset">+{b.offset_ms} мс</span>
                                    )}
                                </div>
                            ))
                        )}
                    </div>

                    <div className="quiz-display-attempts">
                        <span className="quiz-attempts-label">Попытки:</span>
                        <div className="quiz-attempts-dots">
                            {Array.from({ length: state.attempts_max }).map((_, idx) => (
                                <span
                                    key={idx}
                                    className={`quiz-attempt-dot${
                                        idx < state.attempts_used ? " used" : ""
                                    }`}
                                />
                            ))}
                        </div>
                        <span className="quiz-attempts-count">
                            {state.attempts_used} / {state.attempts_max}
                        </span>
                    </div>
                </footer>
            )}
        </div>
    );
}
