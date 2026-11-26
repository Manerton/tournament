import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { useQuizState } from "../helpers/useQuizState";
import { useQuizSounds } from "../helpers/useQuizSounds";
import { quizSoundEngine } from "../helpers/quizSoundEngine";
import {
    QUIZ_ROUND_RULES,
    QUIZ_RULES_FOOTER,
    QUIZ_TIEBREAK_RULES,
    type QuizRoundRules,
} from "../helpers/quizRules";
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

type RulesSlideView = "overview" | "1" | "2" | "3" | "tiebreak";

/** Renders **bold** markers from the rules config. */
function renderRuleText(text: string) {
    return text.split("**").map((part, i) => (i % 2 === 1 ? <strong key={i}>{part}</strong> : part));
}

function RoundRulesSlide({ rules }: { rules: QuizRoundRules }) {
    return (
        <div className="quiz-rules-slide">
            <div className="quiz-rules-head">
                <div className="quiz-rules-title">
                    Тур {rules.round} · «{rules.name}»
                </div>
                <div className="quiz-rules-format">{rules.format}</div>
            </div>

            <div className="quiz-rules-params">
                {rules.params.map((p) => (
                    <span className="quiz-rules-pill" key={p}>
                        {p}
                    </span>
                ))}
            </div>

            {rules.note && <div className="quiz-rules-note">{rules.note}</div>}

            {rules.bullets.length > 0 && (
                <ol className="quiz-rules-bullets">
                    {rules.bullets.map((b, i) => (
                        <li key={i}>{renderRuleText(b)}</li>
                    ))}
                </ol>
            )}

            {rules.scores && (
                <div className="quiz-rules-scores">
                    {rules.scores.map((s) => (
                        <div className={`quiz-rules-score tone-${s.tone}`} key={s.label}>
                            <span className="quiz-rules-score-label">{s.label}</span>
                            <span className="quiz-rules-score-points">{s.points}</span>
                        </div>
                    ))}
                </div>
            )}

            {rules.orders && (
                <div className="quiz-rules-orders">
                    {rules.orders.map((o) => (
                        <div className={`quiz-rules-order tone-${o.tone}`} key={o.name}>
                            <div className="quiz-rules-order-name">{o.name}</div>
                            <div className="quiz-rules-order-points">{o.points}</div>
                            <div className="quiz-rules-order-mistakes">{o.mistakes}</div>
                            <div className="quiz-rules-order-note">({o.note})</div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

function OverviewRulesSlide() {
    return (
        <div className="quiz-rules-slide">
            <div className="quiz-rules-head">
                <div className="quiz-rules-title">Правила турнира</div>
                <div className="quiz-rules-format">Три тура «Королевской битвы»</div>
            </div>

            <div className="quiz-rules-overview">
                {QUIZ_ROUND_RULES.map((r) => (
                    <div className="quiz-rules-col" key={r.round}>
                        <div className="quiz-rules-col-round">Тур {r.round}</div>
                        <div className="quiz-rules-col-name">«{r.name}»</div>
                        <div className="quiz-rules-col-format">{r.format}</div>
                        <div className="quiz-rules-col-params">
                            {r.params.map((p) => (
                                <span className="quiz-rules-pill" key={p}>
                                    {p}
                                </span>
                            ))}
                        </div>
                        <div className="quiz-rules-col-highlight">{r.highlight}</div>
                    </div>
                ))}
            </div>

            <div className="quiz-rules-footer">{QUIZ_RULES_FOOTER}</div>
        </div>
    );
}

function TiebreakRulesSlide() {
    return (
        <div className="quiz-rules-slide">
            <div className="quiz-rules-head">
                <div className="quiz-rules-title">{QUIZ_TIEBREAK_RULES.title}</div>
                <div className="quiz-rules-format">{QUIZ_TIEBREAK_RULES.subtitle}</div>
            </div>

            <div className="quiz-rules-params">
                {QUIZ_TIEBREAK_RULES.params.map((p) => (
                    <span className="quiz-rules-pill" key={p}>
                        {p}
                    </span>
                ))}
            </div>

            <ol className="quiz-rules-bullets">
                {QUIZ_TIEBREAK_RULES.bullets.map((b, i) => (
                    <li key={i}>{renderRuleText(b)}</li>
                ))}
            </ol>
        </div>
    );
}

function RulesSlide({ view }: { view: RulesSlideView }) {
    if (view === "overview") return <OverviewRulesSlide />;
    if (view === "tiebreak") return <TiebreakRulesSlide />;
    const rules = QUIZ_ROUND_RULES.find((r) => r.round === Number(view));
    return rules ? <RoundRulesSlide rules={rules} /> : null;
}

export default function QuizDisplay() {
    const { eventId } = useParams<{ eventId: string }>();
    const { state } = useQuizState(eventId);

    const [remainingMs, setRemainingMs] = useState<number>(0);
    const [soundUnlocked, setSoundUnlocked] = useState<boolean>(false);
    const [soundMuted, setSoundMuted] = useState<boolean>(quizSoundEngine.isMuted());

    const deadlineRef = useRef<number>(0);
    const runningRef = useRef<boolean>(false);

    useQuizSounds(state, remainingMs);

    const unlockAudio = async () => {
        const ok = await quizSoundEngine.unlock();
        setSoundUnlocked(ok);
        if (ok) quizSoundEngine.playQuestionStart();
    };

    const toggleMute = () => {
        const next = !quizSoundEngine.isMuted();
        quizSoundEngine.setMuted(next);
        setSoundMuted(next);
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

    // Projector modes in priority order: leaderboard > rules (manual, then
    // automatic at question 0) > question > idle splash.
    const manualRulesView: RulesSlideView | null =
        state.rules_view && state.rules_view !== "none" ? state.rules_view : null;
    const rulesSlide: RulesSlideView | null = state.leaderboard_shown
        ? null
        : (manualRulesView ??
          (state.tiebreak_active && state.tiebreak_question === 0
              ? "tiebreak"
              : state.current_round > 0 && state.current_question === 0
                ? (String(state.current_round) as RulesSlideView)
                : null));

    const showQuestionScreen = !state.leaderboard_shown && !rulesSlide && Boolean(state.question);
    const isPaused =
        showQuestionScreen &&
        state.timer.started &&
        !state.timer.running &&
        remainingMs > 0 &&
        !state.answer_revealed;
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
                    {showQuestionScreen && state.question
                        ? state.tiebreak_active
                            ? `Резервный вопрос №${state.question.number}`
                            : `Вопрос №${state.question.number}`
                        : state.leaderboard_shown
                          ? "Турнирная таблица"
                          : rulesSlide
                            ? "Правила"
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
                ) : rulesSlide ? (
                    <RulesSlide view={rulesSlide} />
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

                        {state.question.cancelled && (
                            <div className="quiz-display-cancelled">
                                Вопрос отменён — баллы за него не начисляются
                            </div>
                        )}

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
                                {state.question.criteria && (
                                    <div className="quiz-display-criteria">
                                        <span className="quiz-display-criteria-label">
                                            Критерии оценки
                                        </span>
                                        {state.question.criteria}
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

            {!soundUnlocked ? (
                <button
                    type="button"
                    className="quiz-sound-pill"
                    onClick={unlockAudio}
                    title="Нажмите один раз перед турниром, чтобы разрешить звук на табло"
                >
                    🔇 Нажмите, чтобы включить звук
                </button>
            ) : (
                <div className="quiz-sound-controls">
                    <button
                        type="button"
                        onClick={toggleMute}
                        title={soundMuted ? "Включить звук" : "Выключить звук"}
                    >
                        {soundMuted ? "🔇" : "🔊"}
                    </button>
                    <button
                        type="button"
                        onClick={() => quizSoundEngine.playQuestionStart()}
                        title="Проверить громкость в зале"
                    >
                        Тест
                    </button>
                </div>
            )}
        </div>
    );
}
