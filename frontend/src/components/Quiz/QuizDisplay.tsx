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

function playBeep() {
    try {
        const ctx = new AudioContext();
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.frequency.value = 880;
        gain.gain.setValueAtTime(0.4, ctx.currentTime);
        osc.start();
        osc.stop(ctx.currentTime + 1.2);
        osc.onended = () => ctx.close();
    } catch {
        /* audio not available */
    }
}

interface BoardRow {
    team: QuizTeamState;
    value: number;
}

function BoardColumn({ title, rows, max }: { title: string; rows: BoardRow[]; max: number }) {
    return (
        <div className="quiz-board-column">
            <div className="quiz-board-title">{title}</div>
            {rows.map(({ team, value }) => (
                <div className="quiz-bar-row" key={team.team_id}>
                    <div className="quiz-bar-name" title={team.name}>
                        {team.number}. {team.name}
                    </div>
                    <div className="quiz-bar-track">
                        <div
                            className="quiz-bar-fill"
                            style={{ width: max > 0 ? `${Math.max((value / max) * 100, 6)}%` : "6%" }}
                        >
                            {value}
                        </div>
                    </div>
                </div>
            ))}
        </div>
    );
}

function Leaderboard({ round, teams }: { round: number; teams: QuizTeamState[] }) {
    const sortedBy = (fn: (t: QuizTeamState) => number): BoardRow[] =>
        teams
            .map((t) => ({ team: t, value: fn(t) }))
            .sort((a, b) => b.value - a.value || a.team.number - b.team.number);

    if (round === 1) {
        const rows = sortedBy((t) => t.round1_score);
        const max = Math.max(...rows.map((r) => r.value), 1);
        return (
            <div className="quiz-board">
                <BoardColumn title="Итоги первого раунда" rows={rows} max={max} />
            </div>
        );
    }

    if (round === 2) {
        const roundRows = sortedBy((t) => t.round2_score);
        const totalRows = sortedBy((t) => t.round1_score + t.round2_score);
        const max = Math.max(
            ...roundRows.map((r) => r.value),
            ...totalRows.map((r) => r.value),
            1,
        );
        return (
            <div className="quiz-board">
                <BoardColumn title="Второй раунд" rows={roundRows} max={max} />
                <BoardColumn title="Общий зачёт (1 + 2)" rows={totalRows} max={max} />
            </div>
        );
    }

    const roundRows = sortedBy((t) => t.round3_score);
    const totalRows = sortedBy((t) => t.total_score);
    const max = Math.max(
        ...roundRows.map((r) => r.value),
        ...totalRows.map((r) => r.value),
        1,
    );
    return (
        <div className="quiz-board">
            <BoardColumn title="Третий раунд" rows={roundRows} max={max} />
            <BoardColumn title="Финальное положение" rows={totalRows} max={max} />
        </div>
    );
}

export default function QuizDisplay() {
    const { eventId } = useParams<{ eventId: string }>();
    const { state } = useQuizState(eventId);

    const [remainingMs, setRemainingMs] = useState<number>(0);
    const deadlineRef = useRef<number>(0);
    const beepedRef = useRef<string>("");

    // track a local deadline so the timer ticks smoothly between broadcasts
    useEffect(() => {
        if (!state) return;
        deadlineRef.current = Date.now() + state.timer.remaining_ms;
        setRemainingMs(state.timer.remaining_ms);
    }, [state]);

    useEffect(() => {
        const tick = setInterval(() => {
            const rem = Math.max(deadlineRef.current - Date.now(), 0);
            setRemainingMs(rem);
        }, 200);
        return () => clearInterval(tick);
    }, []);

    // audio cue when the timer hits zero
    useEffect(() => {
        if (!state || !state.timer.running) return;
        if (remainingMs > 0 || state.current_question === 0) return;
        const key = `${state.current_round}:${state.current_question}`;
        if (beepedRef.current === key) return;
        beepedRef.current = key;
        playBeep();
    }, [remainingMs, state]);

    if (!state) {
        return (
            <div className="quiz-display-root">
                <BackgroundParticles />
            </div>
        );
    }

    const roundCfg = state.rounds.find((r) => r.round === state.current_round);
    const secondsLeft = Math.ceil(remainingMs / 1000);

    return (
        <div className="quiz-display-root">
            <BackgroundParticles />
            <div className="quiz-display-header">
                <div className="quiz-display-round">
                    {roundCfg ? `Раунд ${roundCfg.round} — «${roundCfg.name}»` : "Квиз-шоу"}
                </div>
                <div className="quiz-display-qnum">
                    {state.question ? `Вопрос №${state.question.number}` : ""}
                </div>
            </div>

            <div className="quiz-display-body">
                {state.leaderboard_shown ? (
                    <Leaderboard round={state.current_round} teams={state.teams} />
                ) : !state.question ? (
                    <div className="quiz-display-idle">
                        {state.round_active
                            ? "Ожидание вопроса…"
                            : "Добро пожаловать на квиз-шоу!"}
                    </div>
                ) : (
                    <>
                        <div className="quiz-display-question">{state.question.text}</div>

                        {state.current_round === 1 && state.buzz_queue.length > 0 && (
                            <div className="quiz-display-queue">
                                {state.buzz_queue.map((b, i) => (
                                    <div
                                        key={b.team_id}
                                        className={`quiz-queue-chip${
                                            b.is_active ? " active" : ""
                                        }${b.status === BuzzStatus.Wrong ? " wrong" : ""}${
                                            b.status === BuzzStatus.Correct ? " correct" : ""
                                        }`}
                                    >
                                        {i + 1}. {b.team_name}
                                    </div>
                                ))}
                            </div>
                        )}

                        {state.current_round === 1 && (
                            <div className="quiz-display-attempts">
                                Попыток использовано: {state.attempts_used} / {state.attempts_max}
                            </div>
                        )}

                        {state.answer_revealed && state.question.answer && (
                            <div className="quiz-display-answer">
                                <div className="quiz-display-answer-label">Правильный ответ</div>
                                {state.question.answer}
                            </div>
                        )}

                        <div
                            className={`quiz-display-timer${
                                secondsLeft <= 10 && secondsLeft > 0 ? " warning" : ""
                            }`}
                        >
                            {formatTime(secondsLeft)}
                        </div>
                    </>
                )}
            </div>
        </div>
    );
}
