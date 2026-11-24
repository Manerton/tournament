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

function playQuestionStartSound() {
    try {
        const audio = new Audio("/start.mp3");
        audio.volume = 0.5;
        audio.play().catch(() => {
            /* autoplay blocked or audio not available */
        });
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
    // tiebreak_score is the secondary sort key: it orders teams tied on the
    // primary value without ever changing the scores themselves
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
    const runningRef = useRef<boolean>(false);
    const beepedRef = useRef<string>("");

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
        }, 200);
        return () => clearInterval(tick);
    }, []);

    // audio cue when the timer hits zero
    useEffect(() => {
        if (!state || !state.timer.running) return;
        const activeQuestion = state.tiebreak_active ? state.tiebreak_question : state.current_question;
        if (remainingMs > 0 || activeQuestion === 0) return;
        const key = `${state.tiebreak_active ? "tb" : state.current_round}:${activeQuestion}`;
        if (beepedRef.current === key) return;
        beepedRef.current = key;
        playQuestionStartSound();
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
                    {state.tiebreak_active
                        ? "Переигровка"
                        : roundCfg
                          ? `Раунд ${roundCfg.round} — «${roundCfg.name}»`
                          : "Квиз-шоу"}
                </div>
                <div className="quiz-display-qnum">
                    {state.question
                        ? state.tiebreak_active
                            ? `Резервный вопрос №${state.question.number}`
                            : `Вопрос №${state.question.number}`
                        : ""}
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

                        {(state.current_round === 1 || state.tiebreak_active) &&
                            state.buzz_queue.length > 0 && (
                            <div className="quiz-display-queue">
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
                        )}

                        {(state.current_round === 1 || state.tiebreak_active) && (
                            <div className="quiz-display-attempts">
                                Попыток использовано: {state.attempts_used} / {state.attempts_max}
                            </div>
                        )}

                        {state.answer_revealed && state.question.answer && (
                            <div className="quiz-display-answer">
                                <div className="quiz-display-answer-label">Правильный ответ</div>
                                {state.question.answer}
                                {state.question.explanation && (
                                    <div className="quiz-display-explanation">{state.question.explanation}</div>
                                )}
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
